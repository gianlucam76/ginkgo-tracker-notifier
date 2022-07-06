package ginkgo_helper

import (
	"fmt"
	"strings"

	ginkgoTypes "github.com/onsi/ginkgo/v2/types"
)

// GetDescription returns the description for a jira issue.
func GetDescription(testReport *ginkgoTypes.SpecReport) string {
	summary := GetSummary(testReport)
	failureLocation := getFailureLocation(testReport.Failure.Location.FullStackTrace)

	// Use only test name and the name of the method where failure happened
	return fmt.Sprintf("Test %q failed: %q", summary, failureLocation)
}

// GetSummary returns, for a given test, the summary of the corresponding Jira bug (if one
// were to be created)
func GetSummary(testReport *ginkgoTypes.SpecReport) string {
	var summary string
	if testReport.LeafNodeText != "" {
		summary = testReport.LeafNodeText
	} else {
		summary = testReport.LeafNodeType.String()
	}
	return summary
}

// GetTestNameAndMaintainer returns test name and maintainer.
// - maintainer is only available if Label maintainer is defined
// - test name is the value of the Label name if defined, or LeafNodeText
func GetTestNameAndMaintainer(testReport *ginkgoTypes.SpecReport) (testName, maintainer string) {
	const infoSize = 2
	for i := range testReport.LeafNodeLabels {
		if strings.Contains(testReport.LeafNodeLabels[i], "maintainer") {
			info := strings.Split(testReport.LeafNodeLabels[i], ":")
			if len(info) == infoSize {
				maintainer = info[1]
			}
		}
		if strings.Contains(testReport.LeafNodeLabels[i], "name") {
			info := strings.Split(testReport.LeafNodeLabels[i], ":")
			if len(info) == infoSize {
				testName = info[1]
			}
		}
	}

	if testName == "" {
		if testReport.LeafNodeText != "" {
			testName = strings.ReplaceAll(testReport.LeafNodeText, " ", "_")
		} else {
			testName = testReport.LeafNodeType.String()
		}
	}

	return
}

// IsTestSerial returns true if a test was run in serial.
// NodeTypeSynchronizedBeforeSuite and NodeTypeSynchronizedAfterSuite have no labels
// but run in serial
func IsTestSerial(testReport *ginkgoTypes.SpecReport) bool {
	if testReport.LeafNodeType == ginkgoTypes.NodeTypeSynchronizedBeforeSuite ||
		testReport.LeafNodeType == ginkgoTypes.NodeTypeSynchronizedAfterSuite {
		return true
	}
	return testReport.IsSerial || testReport.IsInOrderedContainer
}

// getFailureLocation extracts the name of the method where failure happened
func getFailureLocation(stackTrace string) string {
	addressIndex := strings.Index(stackTrace, "0x")
	if addressIndex != -1 {
		return stackTrace[:addressIndex]
	}

	return ""
}
