package process_result

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2" // nolint: golint,stylecheck // ginkgo pattern
	ginkgoTypes "github.com/onsi/ginkgo/v2/types"

	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/elastic_helper"
	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/webex_helper"
)

type info struct {
	elasticInfo *elastic_helper.ElasticInfo
	webexInfo   *webex_helper.WebexInfo
	runID       int64
}

func Register(ctx context.Context, runID int64,
	elasticInfo *elastic_helper.ElasticInfo,
	webexInfo *webex_helper.WebexInfo,
) error {
	c := &info{}

	if runID == 0 {
		return fmt.Errorf("runID cannot be 0")
	}
	c.runID = runID

	if elasticInfo != nil {
		if err := elastic_helper.VerifyInfo(ctx, elasticInfo); err != nil {
			return fmt.Errorf("failed to verify elastic info. Error: %v", err)
		}
		c.elasticInfo = elasticInfo
	}

	if webexInfo != nil {
		if err := webex_helper.VerifyInfo(webexInfo); err != nil {
			return fmt.Errorf("failed to verify webex info. Error: %v", err)
		}
		c.webexInfo = webexInfo
	}

	afterSuiteReport := func(report ginkgoTypes.Report) {
		if c.elasticInfo != nil {
			By(fmt.Sprintf("Save results to elastic db. Run %d", runID))
			elastic_helper.StoreResults(&report, c.runID, c.elasticInfo)
		}

		if c.webexInfo != nil {
			By(fmt.Sprintf("Send failed tests notification to webex room %s", c.webexInfo.Room))
			sendWebexNotification(&report, c)
		}
	}

	ReportAfterSuite("afterSuiteReport", afterSuiteReport)
	return nil
}

// sendWebexNotification send a message for each failed test.
func sendWebexNotification(report *ginkgoTypes.Report, c *info) {
	By("Eventually sending Webex notifications")
	msg := ""
	for i := range report.SpecReports {
		specReport := report.SpecReports[i]
		if specReport.Failed() {
			msg += fmt.Sprintf("Test %s failed in run %d", specReport.FullText(), c.runID)
		}
	}

	webex_helper.SendWebexMessage(c.webexInfo, msg)
}
