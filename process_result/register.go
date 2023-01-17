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
	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/slack_helper"
	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/utils"
	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/webex_helper"
)

type Options struct {
	ElasticInfo *ElasticInfo
	WebexInfo   *WebexInfo
	SlackInfo   *SlackInfo
	JiraInfo    *JiraInfo
	RunID       int64
	DryRun      bool
	EnableLogs  bool
}

type WebexInfo struct {
	AuthToken string // webex auth token
	Room      string // webex room
}

type SlackInfo struct {
	AuthToken string // slack auth token
	Channel   string // slack channel name
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

type Option func(*Options)

func WithLogs() Option {
	return func(args *Options) {
		args.EnableLogs = true
	}
}

func WithRunID(runID int64) Option {
	return func(args *Options) {
		args.RunID = runID
	}
}

func WithDryRun() Option {
	return func(args *Options) {
		args.EnableLogs = true
		args.DryRun = true
	}
}

func WithElastic(info ElasticInfo) Option {
	return func(args *Options) {
		args.ElasticInfo = &info
	}
}

func WithWebex(info WebexInfo) Option {
	return func(args *Options) {
		args.WebexInfo = &info
	}
}

func WithSlack(info SlackInfo) Option {
	return func(args *Options) {
		args.SlackInfo = &info
	}
}

func WithJira(info JiraInfo) Option {
	return func(args *Options) {
		args.JiraInfo = &info
	}
}

// Register register ReportAfterSuite (named afterSuiteReport) when called.
func Register(ctx context.Context, setters ...Option) error {
	c := &Options{}

	for _, setter := range setters {
		setter(c)
	}

	utils.Init(c.EnableLogs)

	if c.DryRun {
		utils.Init(true)
	}

	if c.ElasticInfo != nil {
		if err := verifyElasticInfo(ctx, c); err != nil {
			return err
		}
	}

	if c.WebexInfo != nil {
		if err := verifyWebexInfo(ctx, c); err != nil {
			return err
		}
	}

	if c.SlackInfo != nil {
		if err := verifySlackInfo(ctx, c); err != nil {
			return err
		}
	}

	if c.JiraInfo != nil {
		if err := verifyJiraInfo(ctx, c); err != nil {
			return err
		}
	}

	afterSuiteReport := func(report ginkgoTypes.Report) {
		if c.ElasticInfo != nil {
			utils.Byf(fmt.Sprintf("Save results to elastic db. Run %d", c.RunID))
			elastic_helper.StoreResults(&report, c.RunID, c.getElasticInfo())
		}

		var openIssues []jira.Issue
		if c.JiraInfo != nil {
			utils.Byf(fmt.Sprintf("File jira issue for failed tests. Run %d", c.RunID))
			_ = jira_helper.FileJiraIssuesForFailedTests(context.TODO(), &report, c.RunID, c.getJiraInfo())

			openIssues, _ = jira_helper.GetOpenE2EJiraIssue(context.TODO(), c.getJiraInfo())
		}

		msg := prepareMessage(&report, c, openIssues)

		if c.WebexInfo != nil {
			utils.Byf(fmt.Sprintf("Send failed tests notification to webex room %s", c.WebexInfo.Room))
			sendWebexNotification(&report, c, msg)
		}

		if c.SlackInfo != nil {
			utils.Byf(fmt.Sprintf("Send failed tests notification to slack channel %s", c.SlackInfo.Channel))
			sendSlackNotification(&report, c, msg)
		}
	}

	ReportAfterSuite("afterSuiteReport", afterSuiteReport)
	return nil
}

// sendWebexNotification send a message for each failed test.
func sendWebexNotification(report *ginkgoTypes.Report, c *Options, msg string) {
	utils.Byf("Eventually sending Webex notifications")

	webex_helper.SendWebexMessage(c.getWebexInfo(), msg)
}

// sendSlackNotification send a message for each failed test.
func sendSlackNotification(report *ginkgoTypes.Report, c *Options, msg string) {
	utils.Byf("Eventually sending Slack notifications")

	slack_helper.SendSlackMessage(c.getSlackInfo(), msg)
}

func prepareMessage(report *ginkgoTypes.Report, c *Options, openIssues []jira.Issue) string {
	msg := ""
	for i := range report.SpecReports {
		specReport := report.SpecReports[i]
		if specReport.Failed() {
			testText := specReport.FullText()
			if testText == "" {
				testText = ginkgo_helper.GetSummary(&specReport)
			}
			msg += fmt.Sprintf("Test: %q failed in run %d ", testText, c.RunID)
			if openIssue := jira_helper.FindExistingIssue(openIssues, &specReport); openIssue != nil {
				msg += fmt.Sprintf("current jira issue %s", openIssue.Key)
			}
			msg += "  \n"
		}
	}

	return msg
}

func (i *Options) getWebexInfo() *webex_helper.WebexInfo {
	return &webex_helper.WebexInfo{
		AuthToken: i.WebexInfo.AuthToken,
		Room:      i.WebexInfo.Room,
		DryRun:    i.DryRun,
	}
}

func (i *Options) getSlackInfo() *slack_helper.SlackInfo {
	return &slack_helper.SlackInfo{
		AuthToken: i.SlackInfo.AuthToken,
		Channel:   i.SlackInfo.Channel,
		DryRun:    i.DryRun,
	}
}

func (i *Options) getElasticInfo() *elastic_helper.ElasticInfo {
	return &elastic_helper.ElasticInfo{
		URL:    i.ElasticInfo.URL,
		Index:  i.ElasticInfo.Index,
		DryRun: i.DryRun,
	}
}

func (i *Options) getJiraInfo() *jira_helper.JiraInfo {
	return &jira_helper.JiraInfo{
		BaseURL:   i.JiraInfo.BaseURL,
		Project:   i.JiraInfo.Project,
		Board:     i.JiraInfo.Board,
		Component: i.JiraInfo.Component,
		Username:  i.JiraInfo.Username,
		Password:  i.JiraInfo.Password,
		DryRun:    i.DryRun,
	}
}

func verifyElasticInfo(ctx context.Context, c *Options) error {
	if err := elastic_helper.VerifyInfo(ctx, c.getElasticInfo()); err != nil {
		return fmt.Errorf("failed to verify elastic info. Error: %v", err)
	}
	return nil
}

func verifyWebexInfo(ctx context.Context, c *Options) error {
	if err := webex_helper.VerifyInfo(c.getWebexInfo()); err != nil {
		return fmt.Errorf("failed to verify webex info. Error: %v", err)
	}
	return nil
}

func verifySlackInfo(ctx context.Context, c *Options) error {
	if err := slack_helper.VerifyInfo(ctx, c.getSlackInfo()); err != nil {
		return fmt.Errorf("failed to verify slack info. Error: %v", err)
	}
	return nil
}

func verifyJiraInfo(ctx context.Context, c *Options) error {
	if err := jira_helper.VerifyInfo(ctx, c.getJiraInfo()); err != nil {
		return fmt.Errorf("failed to verify jira info. Error: %v", err)
	}
	return nil
}
