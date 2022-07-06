package process_result

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2" // nolint: golint,stylecheck // ginkgo pattern
	ginkgoTypes "github.com/onsi/ginkgo/v2/types"

	"github.com/andygrunwald/go-jira"

	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/elastic_helper"
	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/ginkgo_helper"
	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/jira_helper"
	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/utils"
	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/webex_helper"
)

type info struct {
	elasticInfo *ElasticInfo
	webexInfo   *WebexInfo
	jiraInfo    *JiraInfo
	runID       int64
	dryRun      bool
}

type WebexInfo struct {
	AuthToken string // webex auth token
	Room      string // webex room
}

type ElasticInfo struct {
	URL   string // elastic DB URL
	Index string // elastic DB Index
}

type JiraInfo struct {
	BaseURL   string // jira base URL
	Project   string // jira Project name
	Board     string // jira Board Name
	Component string // if not empty, any jira filed issue will have this as component
	Username  string // jira username. This is required field. When a test fails, a bug is filed.
	// If a bug is already open for the failed test, no new bug will be open. Simply
	// a new comment will added. jql to search for open bug uses also reporter = username
	Password string // jira password
}

// Register register ReportAfterSuite (named afterSuiteReport) when called.
// Args:
// - runID: is the CI/e2e run ID. This is used when storing result to elastic DB and/or when
// filing jira issues;
// - enableLogs: if set will allow some messages to logged. Otherwise no message is logged;
// - dryRun: indicates whether this is a dry run. dryRun overrides enableLogs to true.
//   In a dry run:
// 		. no result will be stored in elastic DB (only a log containing the results will be displayed)
//      . no message will be sent to webex room (only a log containing the message will be displayed)
//		. no jira bug will be filed or comment added (only a log will be displayed for each jira bug that would have been created/modified)
// - elasticInfo: if provided, test suite results will be stored to an elastic DB;
// - webexInfo: if provided a webex message will be sent for each run with at least one failed test;
// - jiraInfo: if provided, a jira bug will be filed, if one does not exist already, for each failed
// test.
func Register(ctx context.Context, runID int64, enableLogs bool, dryRun bool,
	elasticInfo *ElasticInfo,
	webexInfo *WebexInfo,
	jiraInfo *JiraInfo,
) error {
	c := &info{}

	utils.Init(enableLogs)

	if runID == 0 {
		return fmt.Errorf("runID cannot be 0")
	}
	c.runID = runID

	if dryRun {
		utils.Init(true)
		c.dryRun = true
	}

	if elasticInfo != nil {
		if err := setElasticInfo(ctx, c, elasticInfo); err != nil {
			return err
		}
	}

	if webexInfo != nil {
		if err := setWebexInfo(ctx, c, webexInfo); err != nil {
			return err
		}
	}

	if jiraInfo != nil {
		if err := setJiraInfo(ctx, c, jiraInfo); err != nil {
			return err
		}
	}

	afterSuiteReport := func(report ginkgoTypes.Report) {
		if c.elasticInfo != nil {
			utils.Byf(fmt.Sprintf("Save results to elastic db. Run %d", runID))
			elastic_helper.StoreResults(&report, c.runID, c.getElasticInfo())
		}

		var openIssues []jira.Issue
		if c.jiraInfo != nil {
			utils.Byf(fmt.Sprintf("File jira issue for failed tests. Run %d", runID))
			_ = jira_helper.FileJiraIssuesForFailedTests(context.TODO(), &report, c.runID, c.getJiraInfo())

			openIssues, _ = jira_helper.GetOpenE2EJiraIssue(context.TODO(), c.getJiraInfo())
		}

		msg := prepareMessage(&report, c, openIssues)

		if c.webexInfo != nil {
			utils.Byf(fmt.Sprintf("Send failed tests notification to webex room %s", c.webexInfo.Room))
			sendWebexNotification(&report, c, msg)
		}
	}

	ReportAfterSuite("afterSuiteReport", afterSuiteReport)
	return nil
}

// sendWebexNotification send a message for each failed test.
func sendWebexNotification(report *ginkgoTypes.Report, c *info, msg string) {
	utils.Byf("Eventually sending Webex notifications")

	webex_helper.SendWebexMessage(c.getWebexInfo(), msg)
}

func prepareMessage(report *ginkgoTypes.Report, c *info, openIssues []jira.Issue) string {
	msg := ""
	for i := range report.SpecReports {
		specReport := report.SpecReports[i]
		if specReport.Failed() {
			testText := specReport.FullText()
			if testText == "" {
				testText = ginkgo_helper.GetSummary(&specReport)
			}
			msg += fmt.Sprintf("Test: %q failed in run %d ", testText, c.runID)
			if openIssue := jira_helper.FindExistingIssue(openIssues, &specReport); openIssue != nil {
				msg += fmt.Sprintf("current jira issue %s", openIssue.Key)
			}
			msg += "  \n"
		}
	}

	return msg
}

func (i *info) getWebexInfo() *webex_helper.WebexInfo {
	return &webex_helper.WebexInfo{
		AuthToken: i.webexInfo.AuthToken,
		Room:      i.webexInfo.Room,
		DryRun:    i.dryRun,
	}
}

func (i *info) getElasticInfo() *elastic_helper.ElasticInfo {
	return &elastic_helper.ElasticInfo{
		URL:    i.elasticInfo.URL,
		Index:  i.elasticInfo.Index,
		DryRun: i.dryRun,
	}
}

func (i *info) getJiraInfo() *jira_helper.JiraInfo {
	return &jira_helper.JiraInfo{
		BaseURL:   i.jiraInfo.BaseURL,
		Project:   i.jiraInfo.Project,
		Board:     i.jiraInfo.Board,
		Component: i.jiraInfo.Component,
		Username:  i.jiraInfo.Username,
		Password:  i.jiraInfo.Password,
		DryRun:    i.dryRun,
	}
}

func setElasticInfo(ctx context.Context, c *info, elasticInfo *ElasticInfo) error {
	c.elasticInfo = elasticInfo
	if err := elastic_helper.VerifyInfo(ctx, c.getElasticInfo()); err != nil {
		return fmt.Errorf("failed to verify elastic info. Error: %v", err)
	}
	return nil
}

func setWebexInfo(ctx context.Context, c *info, webexInfo *WebexInfo) error {
	c.webexInfo = webexInfo
	if err := webex_helper.VerifyInfo(c.getWebexInfo()); err != nil {
		return fmt.Errorf("failed to verify webex info. Error: %v", err)
	}
	return nil
}

func setJiraInfo(ctx context.Context, c *info, jiraInfo *JiraInfo) error {
	c.jiraInfo = jiraInfo
	if err := jira_helper.VerifyInfo(ctx, c.getJiraInfo()); err != nil {
		return fmt.Errorf("failed to verify jira info. Error: %v", err)
	}
	return nil
}
