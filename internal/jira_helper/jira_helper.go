package jira_helper

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/andygrunwald/go-jira"
	ginkgoTypes "github.com/onsi/ginkgo/v2/types"

	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/ginkgo_helper"
	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/utils"
)

type JiraInfo struct {
	BaseURL   string // jira base URL
	Project   string // jira Project name
	Board     string // jira Board Name
	Component string // if not empty, any jira filed issue will have this as component
	Username  string // jira username
	Password  string // jira password
	DryRun    bool   // indicates if this is a dryRun
}

// VerifyInfo verifies provided jira info are correct
func VerifyInfo(ctx context.Context, info *JiraInfo) error {
	if info == nil {
		return fmt.Errorf("VerifyInfo passed nil pointer")
	}

	jiraClient, err := getJiraClient(info)
	if err != nil {
		return fmt.Errorf("failed to get jira client. Error %v", err)
	}
	if jiraClient == nil {
		return fmt.Errorf("failed to get jira client")
	}

	project, err := getJiraProject(ctx, jiraClient, info)
	if err != nil {
		return fmt.Errorf("failed to get jira project. Error %v", err)
	}
	if project == nil {
		return fmt.Errorf("failed to get jira project")
	}

	board, err := getJiraBoard(ctx, jiraClient, project.Key, info)
	if err != nil {
		return fmt.Errorf("failed to get jira board. Error %v", err)
	}
	if board == nil {
		return fmt.Errorf("failed to get jira board")
	}

	sprint, err := getJiraActiveSprint(ctx, jiraClient, fmt.Sprintf("%d", board.ID))
	if err != nil {
		return err
	}
	if sprint == nil {
		return fmt.Errorf("failed to get jira active sprint")
	}

	return nil
}

// FileJiraIssuesForFailedTests files an issue, if needed, for a failed test.
// If filed, an issue is added to active sprint.
// Before filing a new issue, this method search if one already exists. If so,
// it simply adds a comment with run id.
// - report is the list of tests
// - runID is current run id
func FileJiraIssuesForFailedTests(ctx context.Context, report *ginkgoTypes.Report, runID int64,
	info *JiraInfo) error {
	jiraClient, err := getJiraClient(info)
	if err != nil || jiraClient == nil {
		msg := "Failed to get jira client"
		utils.Byf(msg)
		return fmt.Errorf("%s", msg)
	}

	project, err := getJiraProject(ctx, jiraClient, info)
	if err != nil || project == nil {
		msg := "Failed to get jira project"
		utils.Byf(msg)
		return fmt.Errorf("%s", msg)
	}

	board, err := getJiraBoard(ctx, jiraClient, project.Key, info)
	if err != nil || board == nil {
		msg := "Failed to get jira board"
		utils.Byf(msg)
		return fmt.Errorf("%s", msg)
	}

	activeSprint, err := getJiraActiveSprint(ctx, jiraClient, fmt.Sprintf("%d", board.ID))
	if err != nil || activeSprint == nil {
		msg := "Failed to get active sprint"
		utils.Byf(msg)
		return fmt.Errorf("%s", msg)
	}

	openIssues, err := GetOpenE2EJiraIssue(ctx, info)
	if err != nil {
		msg := "Failed to get open jira issue"
		utils.Byf(msg)
		return fmt.Errorf("%s", msg)
	}

	priority := jira.Priority{Name: "P1"}
	for i := range report.SpecReports {
		testReport := report.SpecReports[i]
		testName, maintainer := ginkgo_helper.GetTestNameAndMaintainer(&testReport)

		// If SynchronizeBeforeSuite fails, all nodes will report an issue. But only
		// one will have stack trace set.
		if testReport.Failed() && testReport.FailureLocation().FullStackTrace != "" {
			if openIssue := FindExistingIssue(openIssues, &testReport); openIssue != nil {
				utils.Byf(fmt.Sprintf("Adding comment to issue for test %s", testName))
				if info.DryRun {
					continue
				}
				addCommentToIssue(ctx, jiraClient, openIssue.ID, fmt.Sprintf("%d", runID), &testReport)
				moveIssueToSprint(ctx, jiraClient, activeSprint.ID, openIssue.ID)
			} else {
				utils.Byf(fmt.Sprintf("Filing issue for test %s", testName))
				if info.DryRun {
					continue
				}
				_ = createIssue(ctx, jiraClient, activeSprint, &priority, project.Key,
					info.Component, maintainer, fmt.Sprintf("%d", runID), &testReport)
			}
		}
	}

	return nil
}

// getJiraClient returns a new Jira API client.
func getJiraClient(info *JiraInfo) (*jira.Client, error) {
	var jiraClient *jira.Client
	var err error
	if info.Username != "" && info.Password != "" {
		tp := jira.BasicAuthTransport{
			Username: info.Username,
			Password: info.Password,
		}
		jiraClient, err = jira.NewClient(tp.Client(), info.BaseURL)
	} else {
		jiraClient, err = jira.NewClient(nil, info.BaseURL)
	}

	if err != nil {
		utils.Byf(fmt.Sprintf("Failed to get jira client. Err: %v", err))
		return nil, err
	}

	return jiraClient, nil
}

// getJiraProject returns the jira.Project with name projectName
func getJiraProject(ctx context.Context, jiraClient *jira.Client, info *JiraInfo) (*jira.Project, error) {
	url := fmt.Sprintf("rest/api/2/project/%s", info.Project)
	req, _ := jiraClient.NewRequestWithContext(ctx, "GET", url, nil)
	project := &jira.Project{}
	if resp, err := jiraClient.Do(req, project); err != nil {
		body, _ := io.ReadAll(resp.Body)
		utils.Byf(fmt.Sprintf("Failed to get project with name: %s. Error: %v. Response: %s", info.Project, err, string(body)))
		return nil, err
	}

	return project, nil
}

// getJiraBoard returns board with name boardName in project projectKey
// returns the board if only one is found or an error if any occurs.
// Returns nil if no board is found or more than one is found
func getJiraBoard(ctx context.Context, jiraClient *jira.Client, projectKey string, info *JiraInfo) (*jira.Board, error) {
	boardListOptions := &jira.BoardListOptions{ProjectKeyOrID: projectKey, Name: info.Board}
	boardList, resp, err := jiraClient.Board.GetAllBoardsWithContext(ctx, boardListOptions)
	if err != nil {
		body, _ := io.ReadAll(resp.Body)
		utils.Byf(fmt.Sprintf("Failed to get board list. Error %v. Response: %s", err, string(body)))
		return nil, err
	}

	if boardList.Values == nil {
		utils.Byf(fmt.Sprintf("Got not result for GetAllBoards with projectKey: %s and boardName: %s ", projectKey, info.Board))
		return nil, nil
	}

	if len(boardList.Values) != 1 {
		utils.Byf(fmt.Sprintf("Got more than one result for GetAllBoards with projectKey: %s and boardName: %s ", projectKey, info.Board))
		utils.Byf(fmt.Sprintf("Result: %v", boardList.Values))
		return nil, nil
	}

	utils.Byf(fmt.Sprintf("Board %s found", boardList.Values[0].Name))
	return &boardList.Values[0], nil
}

// getJiraActiveSprint returns the active sprint for passed in board
// Returns active sprint if found or an error if any occurs.
// If no sprint is currently active, latest active sprint will be returned
func getJiraActiveSprint(ctx context.Context, jiraClient *jira.Client, boardID string) (*jira.Sprint, error) {
	if jiraClient == nil {
		msg := "jiraClient is nil"
		utils.Byf(msg)
		return nil, fmt.Errorf("%s", msg)
	}

	sprints, _, err := jiraClient.Board.GetAllSprintsWithContext(ctx, boardID)
	if err != nil {
		utils.Byf(fmt.Sprintf("Failed to get board list. Error: %v", err))
		return nil, err
	}

	var activeSprint *jira.Sprint
	now := time.Now()
	for i := range sprints {
		if sprints[i].StartDate != nil && sprints[i].EndDate != nil {
			if sprints[i].StartDate.Before(now) && sprints[i].EndDate.After(now) {
				return &sprints[i], nil
			} else if sprints[i].StartDate.Before(now) {
				if activeSprint == nil {
					activeSprint = &sprints[i]
				} else if sprints[i].EndDate.After(*activeSprint.EndDate) {
					activeSprint = &sprints[i]
				}
			}
		}
	}

	return activeSprint, nil
}

// getJiraIssues finds all issues matching passed jql
func getJiraIssues(ctx context.Context, jiraClient *jira.Client, jql string) ([]jira.Issue, error) {
	issues, _, err := jiraClient.Issue.SearchWithContext(ctx, jql, nil)
	if err != nil {
		utils.Byf(fmt.Sprintf("Failed to get all issues matching jql:%s. Error: %v", jql, err))
		return nil, err
	}

	return issues, nil
}

// FindExistingIssue finds if an already existing issue exists.
func FindExistingIssue(openIssues []jira.Issue, testReport *ginkgoTypes.SpecReport) *jira.Issue {
	// First search for a match with description.
	// Description contains both test name and failure location.
	description := ginkgo_helper.GetDescription(testReport)
	for i := range openIssues {
		if openIssues[i].Fields != nil && openIssues[i].Fields.Description == description {
			return &openIssues[i]
		}
	}

	// Search for an issue matching summary only
	summary := ginkgo_helper.GetSummary(testReport)
	for i := range openIssues {
		if openIssues[i].Fields != nil && openIssues[i].Fields.Summary == summary {
			return &openIssues[i]
		}
	}
	return nil
}

// createIssue creates new issue of type bug which will be added to sprint
// - Comments will contain run ID, failure message and full stack trace
// - Assignee is the user the bug will be assigned to
// - Reporter is the issue reporter
// Return the issue Key or empty an error occurred.
func createIssue(ctx context.Context, jiraClient *jira.Client, sprint *jira.Sprint, priority *jira.Priority,
	projectKey, componentName, assignee, runID string,
	testReport *ginkgoTypes.SpecReport) string {
	summary := ginkgo_helper.GetSummary(testReport)

	i := jira.Issue{
		Fields: &jira.IssueFields{
			Description: ginkgo_helper.GetDescription(testReport),
			Type: jira.IssueType{
				Name: "Bug",
			},
			Project: jira.Project{
				Key: projectKey,
			},
			Summary:  summary,
			Priority: priority,
		},
	}

	if componentName != "" {
		component := jira.Component{Name: componentName}
		i.Fields.Components = []*jira.Component{&component}
	}

	if assignee != "" {
		i.Fields.Assignee = &jira.User{Name: assignee}
	}

	issue, resp, err := jiraClient.Issue.CreateWithContext(ctx, &i)
	if err != nil {
		body, _ := io.ReadAll(resp.Body)
		utils.Byf(fmt.Sprintf("Failed to create issue. Error: %v. Resp %s", err, string(body)))
		return ""
	}

	utils.Byf(fmt.Sprintf("Created issue %s", issue.Key))

	addCommentToIssue(ctx, jiraClient, issue.ID, runID, testReport)

	moveIssueToSprint(ctx, jiraClient, sprint.ID, issue.ID)

	return issue.Key
}

// addCommentToIssue append comment to current open issue while also resetting sprint and priority.
// The new appended comment will contain buildEnvironment (VCS vs UCS), run ID, failure message and full stack trace
func addCommentToIssue(ctx context.Context, jiraClient *jira.Client, issueID string,
	runID string, testReport *ginkgoTypes.SpecReport) {
	comment := jira.Comment{
		Body: fmt.Sprintf("Run: %s\n\nFailure Location: %s\n\nFull Stack Trace %s",
			runID, testReport.Failure.Location.String(),
			testReport.Failure.Location.FullStackTrace),
	}

	if _, resp, err := jiraClient.Issue.AddCommentWithContext(ctx, issueID, &comment); err != nil {
		body, _ := io.ReadAll(resp.Body)
		utils.Byf(fmt.Sprintf("Failed to update issue %s. Error: %v. Resp %s", issueID, err, string(body)))
		return
	}

	utils.Byf("Update issue with comment")
}

func moveIssueToSprint(ctx context.Context, jiraClient *jira.Client, sprintID int, issueID string) {
	if resp, err := jiraClient.Sprint.MoveIssuesToSprintWithContext(ctx, sprintID, []string{issueID}); err != nil {
		body, _ := io.ReadAll(resp.Body)
		utils.Byf(fmt.Sprintf("Failed to update issue %s. Error: %v. Resp %s", issueID, err, string(body)))
		return
	}
	utils.Byf("Moved issue to sprint")
}

// GetOpenE2EJiraIssue returns issues filed by user (in jiraInfo)
func GetOpenE2EJiraIssue(ctx context.Context, info *JiraInfo) ([]jira.Issue, error) {
	jiraClient, err := getJiraClient(info)
	if err != nil || jiraClient == nil {
		utils.Byf("Failed to get jira client")
		return nil, fmt.Errorf("failed to get jira client")
	}

	jql := fmt.Sprintf("reporter = %s and type = Bug and Status NOT IN (Resolved,Closed)", info.Username)
	openIssues, err := getJiraIssues(ctx, jiraClient, jql)
	if err != nil {
		utils.Byf("Failed to get open jira issue")
		return nil, fmt.Errorf("failed to get open jira issue")
	}

	return openIssues, nil
}
