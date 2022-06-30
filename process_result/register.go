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
	elasticInfo *ElasticInfo
	webexInfo   *WebexInfo
	runID       int64
}

type WebexInfo struct {
	AuthToken string // webex auth token
	Room      string // webex room
}

type ElasticInfo struct {
	URL   string // elastic DB URL
	Index string // elastic DB Index
}

func Register(ctx context.Context, runID int64,
	elasticInfo *ElasticInfo,
	webexInfo *WebexInfo,
) error {
	c := &info{}

	if runID == 0 {
		return fmt.Errorf("runID cannot be 0")
	}
	c.runID = runID

	if elasticInfo != nil {
		c.elasticInfo = elasticInfo
		if err := elastic_helper.VerifyInfo(ctx, c.getElasticInfo()); err != nil {
			return fmt.Errorf("failed to verify elastic info. Error: %v", err)
		}
	}

	if webexInfo != nil {
		c.webexInfo = webexInfo
		if err := webex_helper.VerifyInfo(c.getWebexInfo()); err != nil {
			return fmt.Errorf("failed to verify webex info. Error: %v", err)
		}
	}

	afterSuiteReport := func(report ginkgoTypes.Report) {
		if c.elasticInfo != nil {
			By(fmt.Sprintf("Save results to elastic db. Run %d", runID))
			elastic_helper.StoreResults(&report, c.runID, c.getElasticInfo())
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

	webex_helper.SendWebexMessage(c.getWebexInfo(), msg)
}

func (i *info) getWebexInfo() *webex_helper.WebexInfo {
	return &webex_helper.WebexInfo{
		AuthToken: i.webexInfo.AuthToken,
		Room:      i.webexInfo.Room,
	}
}

func (i *info) getElasticInfo() *elastic_helper.ElasticInfo {
	return &elastic_helper.ElasticInfo{
		URL:   i.elasticInfo.URL,
		Index: i.elasticInfo.Index,
	}
}
