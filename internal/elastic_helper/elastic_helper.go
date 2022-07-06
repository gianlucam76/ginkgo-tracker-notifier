package elastic_helper

import (
	"context"
	"fmt"
	"strings"
	"time"

	ginkgoTypes "github.com/onsi/ginkgo/v2/types"

	"github.com/gdexlab/go-render/render"
	elastic "github.com/olivere/elastic/v7"

	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/ginkgo_helper"
	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/utils"
)

type ElasticInfo struct {
	URL    string // elastic DB URL
	Index  string // elastic DB Index
	DryRun bool   // indicates if this is a dryRun
}

type ElasticResult struct {
	// Name is the name of the test
	Name string `json:"name"`
	// Description is the test description
	Description string `json:"description"`
	// Maintainer is the maintainer for a given test
	Maintainer string `json:"maintainer"`
	// DurationInMinutes is the duration of the test in minutes
	DurationInMinutes float64 `json:"durationInMinutes"`
	// Duration is the duration of the test in seconds
	DurationInSecond time.Duration `json:"durationInSeconds"`
	// Result indicates whether test passed or failed or it was skipped
	Result string `json:"result"`
	// Run is the sanity run id
	Run int64 `json:"run"`
	// StartTime is the time test started
	StartTime time.Time `json:"startTime"`
	// Serial indicates whether test was run in serial
	Serial bool `json:"serial"`
}

const (
	healthCheckInterval = 10 * time.Second
)

// VerifyInfo verifies provided info (elastic DB and Index) are correct
func VerifyInfo(ctx context.Context, info *ElasticInfo) error {
	if info == nil {
		return fmt.Errorf("VerifyInfo passed nil pointer")
	}

	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(info.URL),
		elastic.SetHealthcheckInterval(healthCheckInterval),
	)

	if err != nil {
		msg := fmt.Sprintf("Failed to create client to access es: %v", err)
		utils.Byf(msg)
		return fmt.Errorf("%s", msg)
	}

	exist, err := client.IndexExists(info.Index).Do(ctx)
	if err != nil {
		msg := fmt.Sprintf("Failed to check index %s existence err: %v", info.Index, err)
		utils.Byf(msg)
		return fmt.Errorf("%s", msg)
	}
	if !exist {
		msg := fmt.Sprintf("Index %s does not exist", info.Index)
		utils.Byf(msg)
		return fmt.Errorf("%s", msg)
	}

	return nil
}

// StoreResults store test results in elastic db.
// - report is the list of tests
// - buildID is current run id
// - buildEnvironment is current run environment
func StoreResults(report *ginkgoTypes.Report, runID int64, info *ElasticInfo) {
	ctx := context.TODO()

	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(info.URL),
		elastic.SetHealthcheckInterval(healthCheckInterval),
	)
	if err != nil {
		utils.Byf(fmt.Sprintf("Failed to create client to access es: %v", err))
		return
	}

	exist, err := client.IndexExists(info.Index).Do(ctx)
	if err != nil {
		utils.Byf(fmt.Sprintf("Failed to check index %s existence err: %v", info.Index, err))
		return
	}
	if !exist {
		utils.Byf(fmt.Sprintf("Index %s does not exist", info.Index))
		return
	}

	utils.Byf(fmt.Sprintf("Found %d tests", len(report.SpecReports)))

	for i := range report.SpecReports {
		testReport := report.SpecReports[i]

		storeResult(testReport, client, info.Index, runID, info.DryRun)
	}
}

// GetFailuresForRun returns failed test for a given run <buildEnvironment, buildID>
func GetFailuresForRun(buildID int, buildEnvironment, esURL, index string) (*elastic.SearchResult, error) {
	const maxResult = 200
	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(esURL),
		elastic.SetHealthcheckInterval(healthCheckInterval),
	)
	if err != nil {
		utils.Byf(fmt.Sprintf("Failed to create client to access es: %v", err))
		return nil, err
	}

	generalQ := elastic.NewBoolQuery().Should()

	// Filter by failed test
	generalQ.Filter(elastic.NewMatchQuery("result", ginkgoTypes.SpecStateFailed.String()))
	// Filter by run
	generalQ.Filter(elastic.NewMatchQuery("run", fmt.Sprintf("%d", buildID)))

	searchResult, err := client.Search().Index(index).Query(generalQ).Size(maxResult).
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		utils.Byf(fmt.Sprintf("Failed to run query %v", err))
		return nil, err
	}

	return searchResult, nil
}

func storeResult(testReport ginkgoTypes.SpecReport, client *elastic.Client, index string,
	runID int64, dryRun bool) {
	testName, maintainer := ginkgo_helper.GetTestNameAndMaintainer(&testReport)

	// E2E runs in parallel. GINKGO_NODES defines how many nodes.
	// That means there are multiple SynchronizedAfterSuite and SynchronizedAfterSuite
	// running, one per node. But only one SynchronizedBeforeSuite and one SynchronizedAfterSuite
	// actually do work. Store result only for the first one
	if testName == ginkgoTypes.NodeTypeSynchronizedBeforeSuite.String() ||
		testName == ginkgoTypes.NodeTypeSynchronizedAfterSuite.String() {
		if testReport.ParallelProcess != 1 {
			return
		}
	}

	r := ElasticResult{
		Name: testName,
		// Description is what allows us to find from a query in es for a failed test, the corresponding Jira bug
		Description:       ginkgo_helper.GetSummary(&testReport),
		DurationInMinutes: testReport.RunTime.Minutes(),
		DurationInSecond:  testReport.RunTime.Round(time.Second),
		Run:               runID,
		Maintainer:        maintainer,
		StartTime:         testReport.StartTime,
		Serial:            ginkgo_helper.IsTestSerial(&testReport),
	}
	r.Result = testReport.State.String()

	runInfo := fmt.Sprintf("run_%d_test_%s", runID, strings.TrimSpace(testName))
	if dryRun {
		utils.Byf("Run ID: %d Store ElasticResult %s", runID, render.AsCode(r))
		return
	}

	_, err := client.Index().Index(index).Id(runInfo).BodyJson(r).Do(context.TODO())
	if err != nil {
		utils.Byf(fmt.Sprintf("Failed to store result %s. Result %s", testName, r.Result))
	} else {
		utils.Byf(fmt.Sprintf("Stored result %s. Result %s", testName, r.Result))
	}
}
