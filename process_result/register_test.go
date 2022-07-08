package process_result_test

import (
	"fmt"
	"reflect"
	"time"

	"github.com/andygrunwald/go-jira"
	. "github.com/onsi/ginkgo/v2"
	ginkgoTypes "github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/ginkgo_helper"
	"github.com/gianlucam76/ginkgo-tracker-notifier/process_result"
)

func getSpecReport() []ginkgoTypes.SpecReport {
	return []ginkgoTypes.SpecReport{
		{
			LeafNodeType: ginkgoTypes.NodeTypeIt,
			State:        ginkgoTypes.SpecStatePassed,
			NumAttempts:  1,
			RunTime:      time.Second,
		},
		{
			LeafNodeType:            ginkgoTypes.NodeTypeIt,
			State:                   ginkgoTypes.SpecStateFailed,
			NumAttempts:             1,
			RunTime:                 time.Second,
			LeafNodeText:            "return correct data based on labels",
			ContainerHierarchyTexts: []string{"Verify Labels", "Filter on Labels"},
		},
		{
			LeafNodeType:            ginkgoTypes.NodeTypeIt,
			State:                   ginkgoTypes.SpecStateFailed,
			NumAttempts:             1,
			RunTime:                 time.Second,
			LeafNodeText:            "return ordered list",
			ContainerHierarchyTexts: []string{"Verify list methods"},
		},
		{
			LeafNodeType:            ginkgoTypes.NodeTypeSynchronizedBeforeSuite,
			State:                   ginkgoTypes.SpecStateFailed,
			NumAttempts:             1,
			RunTime:                 time.Second,
			LeafNodeText:            "",
			ContainerHierarchyTexts: nil,
		},
		{
			LeafNodeType: ginkgoTypes.NodeTypeIt,
			State:        ginkgoTypes.SpecStateSkipped,
			NumAttempts:  1,
			RunTime:      time.Second,
		},
	}
}

var _ = Describe("PrepareMessage", func() {
	It("Prepare correct message when Jira issues are not present", func() {
		report := ginkgoTypes.Report{
			SpecReports: getSpecReport(),
		}
		c := &process_result.Options{}
		setter := process_result.WithRunID(int64(65512))
		setter(c)

		message := process_result.PrepareMessage(&report, c, nil)
		Expect(message).To(ContainSubstring("Test: \"Verify Labels Filter on Labels return correct data based on labels\" failed in run 65512"))
		Expect(message).To(ContainSubstring("Test: \"Verify list methods return ordered list\" failed in run 65512"))
		Expect(message).To(ContainSubstring("Test: \"SynchronizedBeforeSuite\" failed in run 65512"))
	})

	It("Prepare correct message when Jira issues are not present", func() {
		report := ginkgoTypes.Report{
			SpecReports: getSpecReport(),
		}
		c := &process_result.Options{}
		setter := process_result.WithRunID(int64(1623))
		setter(c)

		expected := make([]string, 0)
		openIssue := make([]jira.Issue, 0)
		for i := range report.SpecReports {
			specReport := &report.SpecReports[i]
			if specReport.Failed() {
				description := ginkgo_helper.GetDescription(specReport)
				issue := jira.Issue{
					Key: fmt.Sprintf("key%d", i),
					Fields: &jira.IssueFields{
						Description: description,
					},
				}
				openIssue = append(openIssue, issue)

				testText := specReport.FullText()
				if testText == "" {
					testText = ginkgo_helper.GetSummary(specReport)
				}
				expected = append(expected, fmt.Sprintf("Test: %q failed in run 1623 current jira issue %s", testText, issue.Key))
			}
		}

		message := process_result.PrepareMessage(&report, c, openIssue)
		for i := range expected {
			Expect(message).To(ContainSubstring(expected[i]))
		}
	})
})

var _ = Describe("Setters", func() {
	It("WithLogs enables logs", func() {
		f := process_result.WithLogs()
		c := &process_result.Options{}
		f(c)
		Expect(c.EnableLogs).To(BeTrue())
	})

	It("WithLogs sets RunID", func() {
		runId := int64(4568)
		f := process_result.WithRunID(runId)
		c := &process_result.Options{}
		f(c)
		Expect(c.RunID).To(Equal(runId))
	})

	It("WithDryRun enables logs and sets DryRun", func() {
		f := process_result.WithDryRun()
		c := &process_result.Options{}
		f(c)
		Expect(c.EnableLogs).To(BeTrue())
		Expect(c.DryRun).To(BeTrue())
	})

	It("WithElastic sets ElasticInfo", func() {
		f := process_result.WithElastic(*getElasticInfo())
		c := &process_result.Options{
			JiraInfo:  getJiraInfo(),
			SlackInfo: getSlackInfo(),
			WebexInfo: getWebexInfo(),
		}
		f(c)
		Expect(c.ElasticInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.ElasticInfo, *getElasticInfo())).To(BeTrue())
		Expect(c.JiraInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.JiraInfo, *getJiraInfo())).To(BeTrue())
		Expect(c.SlackInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.SlackInfo, *getSlackInfo())).To(BeTrue())
		Expect(c.WebexInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.WebexInfo, *getWebexInfo())).To(BeTrue())
	})

	It("WithWebex sets WebexInfo", func() {
		f := process_result.WithWebex(*getWebexInfo())
		c := &process_result.Options{
			JiraInfo:    getJiraInfo(),
			SlackInfo:   getSlackInfo(),
			ElasticInfo: getElasticInfo(),
		}
		f(c)
		Expect(c.ElasticInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.ElasticInfo, *getElasticInfo())).To(BeTrue())
		Expect(c.JiraInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.JiraInfo, *getJiraInfo())).To(BeTrue())
		Expect(c.SlackInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.SlackInfo, *getSlackInfo())).To(BeTrue())
		Expect(c.WebexInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.WebexInfo, *getWebexInfo())).To(BeTrue())
	})

	It("WithSlack sets SlackInfo", func() {
		f := process_result.WithSlack(*getSlackInfo())
		c := &process_result.Options{
			JiraInfo:    getJiraInfo(),
			WebexInfo:   getWebexInfo(),
			ElasticInfo: getElasticInfo(),
		}
		f(c)
		Expect(c.ElasticInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.ElasticInfo, *getElasticInfo())).To(BeTrue())
		Expect(c.JiraInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.JiraInfo, *getJiraInfo())).To(BeTrue())
		Expect(c.SlackInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.SlackInfo, *getSlackInfo())).To(BeTrue())
		Expect(c.WebexInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.WebexInfo, *getWebexInfo())).To(BeTrue())
	})

	It("WithJira sets JiraInfo", func() {
		f := process_result.WithJira(*getJiraInfo())
		c := &process_result.Options{
			SlackInfo:   getSlackInfo(),
			WebexInfo:   getWebexInfo(),
			ElasticInfo: getElasticInfo(),
		}
		f(c)
		Expect(c.ElasticInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.ElasticInfo, *getElasticInfo())).To(BeTrue())
		Expect(c.JiraInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.JiraInfo, *getJiraInfo())).To(BeTrue())
		Expect(c.SlackInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.SlackInfo, *getSlackInfo())).To(BeTrue())
		Expect(c.WebexInfo).ToNot(BeNil())
		Expect(reflect.DeepEqual(*c.WebexInfo, *getWebexInfo())).To(BeTrue())
	})
})

func getWebexInfo() *process_result.WebexInfo {
	return &process_result.WebexInfo{
		Room:      "e2e result",
		AuthToken: "98765aaaaa",
	}
}

func getSlackInfo() *process_result.SlackInfo {
	return &process_result.SlackInfo{
		Channel:   "qa testing",
		AuthToken: "128io6827896",
	}
}

func getElasticInfo() *process_result.ElasticInfo {
	return &process_result.ElasticInfo{
		URL:   "https://elastic.org",
		Index: "cs_e2e",
	}
}

func getJiraInfo() *process_result.JiraInfo {
	return &process_result.JiraInfo{
		BaseURL:  "https://jira.org",
		Project:  "my project",
		Board:    "p1 board",
		Username: "username",
		Password: "password",
	}
}
