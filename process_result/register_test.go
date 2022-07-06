package process_result_test

import (
	"fmt"
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
		c := &process_result.Info{}
		c.SetRundId(int64(65512))

		message := process_result.PrepareMessage(&report, c, nil)
		Expect(message).To(ContainSubstring("Test: \"Verify Labels Filter on Labels return correct data based on labels\" failed in run 65512"))
		Expect(message).To(ContainSubstring("Test: \"Verify list methods return ordered list\" failed in run 65512"))
		Expect(message).To(ContainSubstring("Test: \"SynchronizedBeforeSuite\" failed in run 65512"))
	})

	It("Prepare correct message when Jira issues are not present", func() {
		report := ginkgoTypes.Report{
			SpecReports: getSpecReport(),
		}
		c := &process_result.Info{}
		c.SetRundId(int64(1623))

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
