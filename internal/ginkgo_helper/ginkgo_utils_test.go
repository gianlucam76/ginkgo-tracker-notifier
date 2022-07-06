package ginkgo_helper_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ginkgoTypes "github.com/onsi/ginkgo/v2/types"

	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/ginkgo_helper"
)

func getReport() *ginkgoTypes.SpecReport {
	return &ginkgoTypes.SpecReport{
		LeafNodeType: ginkgoTypes.NodeTypeIt,
		State:        ginkgoTypes.SpecStatePassed,
		NumAttempts:  1,
		RunTime:      time.Second,
		LeafNodeText: "return correct data based on labels",
	}
}

func getFullStackTrace() string {
	return `golang.cisco.com/cloudstack/test/framework.waitForSveltosServerRequestConditions({0x400e5e8, 0xc000144008}, {0x40625c0, 0xc0008390a0}, {0xc001902a00, 0x50}, 0xc000473600, {0xc00507f1c0, 0x2, 0x2})
	/root/src/e2e/framework/clusterapi_helpers.go:3131 +0x33e
golang.cisco.com/cloudstack/test/framework.WaitForSveltosServerRequestStatus({0x400e5e8, 0xc000144008}, {0x40625c0, 0xc0008390a0}, 0xc0006fc380, 0xc002f57200, 0xc001028b70, 0xc00076dae0)
	/root/src/e2e/framework/clusterapi_helpers.go:1736 +0x8e5
golang.cisco.com/cloudstack/test/framework.WaitForMachineToBeProvisioned({0x400e5e8, 0xc000144008}, {0x40625c0, 0xc0008390a0}, 0xc0006fc380, 0xc002f57200, 0x1)
	/root/src/e2e/framework/clusterapi_helpers.go:520 +0x4f1
golang.cisco.com/cloudstack/test/e2e/failuredomain.fdSpec.func2()
	/root/src/e2e/e2e/failuredomain/failuredomain.go:172 +0x1146`
}

var _ = Describe("GinkgoUtils", func() {
	It("GetSummary returns LeafNodeText when set", func() {
		report := getReport()
		summary := ginkgo_helper.GetSummary(report)
		Expect(summary).ToNot(BeEmpty())
		Expect(summary).To(Equal(report.LeafNodeText))
	})

	It("GetSummary returns LeafNodeType when LeafNodeText is not set", func() {
		report := getReport()
		report.LeafNodeText = ""
		summary := ginkgo_helper.GetSummary(report)
		Expect(summary).ToNot(BeEmpty())
		Expect(summary).To(Equal(report.LeafNodeType.String()))
	})

	It("GetFailureLocation returns method where failure happened", func() {
		report := getReport()
		report.Failure.Location.FullStackTrace = getFullStackTrace()
		failureLocation := ginkgo_helper.GetFailureLocation(report.Failure.Location.FullStackTrace)
		Expect(failureLocation).To(Equal("golang.cisco.com/cloudstack/test/framework.waitForSveltosServerRequestConditions({"))
	})

	It("GetDescription returns description for a report", func() {
		report := getReport()
		report.Failure.Location.FullStackTrace = getFullStackTrace()
		description := ginkgo_helper.GetDescription(report)
		expectedDescription := fmt.Sprintf("Test %q failed: %q",
			ginkgo_helper.GetSummary(report),
			ginkgo_helper.GetFailureLocation(report.Failure.Location.FullStackTrace))
		Expect(description).To(Equal(expectedDescription))
	})

	It("IsTestSerial returns true for serial test", func() {
		report := getReport()
		report.IsSerial = true
		Expect(ginkgo_helper.IsTestSerial(report)).To(BeTrue())
	})

	It("IsTestSerial returns true for an ordered test", func() {
		report := getReport()
		report.IsInOrderedContainer = true
		Expect(ginkgo_helper.IsTestSerial(report)).To(BeTrue())
	})

	It("IsTestSerial returns true for an SynchronizedBeforeSuite", func() {
		report := getReport()
		report.LeafNodeType = ginkgoTypes.NodeTypeSynchronizedBeforeSuite
		Expect(ginkgo_helper.IsTestSerial(report)).To(BeTrue())
	})

	It("IsTestSerial returns true for an SynchronizedAfterSuite", func() {
		report := getReport()
		report.LeafNodeType = ginkgoTypes.NodeTypeSynchronizedAfterSuite
		Expect(ginkgo_helper.IsTestSerial(report)).To(BeTrue())
	})

	It("IsTestSerial returns false for non serial, non ordered test", func() {
		report := getReport()
		Expect(ginkgo_helper.IsTestSerial(report)).To(BeFalse())
	})

	It("GetTestNameAndMaintainer returns maintainer and test name whe labels are set", func() {
		report := getReport()
		report.LeafNodeLabels = []string{"name:verify-labels", "maintainer:user-a"}
		testName, maintainer := ginkgo_helper.GetTestNameAndMaintainer(report)
		Expect(testName).To(Equal("verify-labels"))
		Expect(maintainer).To(Equal("user-a"))
	})

	It("GetTestNameAndMaintainer returns LeafNodeText when name label is missing", func() {
		report := getReport()
		report.LeafNodeLabels = []string{"maintainer:user-a"}
		testName, _ := ginkgo_helper.GetTestNameAndMaintainer(report)
		Expect(testName).To(Equal(strings.ReplaceAll(report.LeafNodeText, " ", "_")))
	})

	It("GetTestNameAndMaintainer returns LeafNodeType when name label is missing and LeafNodeText is not set", func() {
		report := getReport()
		report.LeafNodeLabels = []string{"maintainer:user-a"}
		report.LeafNodeText = ""
		testName, _ := ginkgo_helper.GetTestNameAndMaintainer(report)
		Expect(testName).To(Equal(report.LeafNodeType.String()))
	})
})
