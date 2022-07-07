If you are using [Ginkgo](https://onsi.github.io/ginkgo), this package can be of help.

This package, registered from a ginkgo test suite, provides support for:
1. Store test-suite results to an elastic DB;
2. Send notification for failed tests in a test-suite to a webex channel/slack channel;
3. File Jira issue for each failed tests in a test-suite.

You can pick what you need (all or a just a subset of supported features).

## Installing

### *go get*

    $ go get -u gianlucam76/ginkgo-tracker-notifier

## Example

```
import (
 	"ginkgo_helper "github.com/gianlucam76/ginkgo-tracker-notifier/process_result"
 )
```

```
func TestLib(t *testing.T) {
	RegisterFailHandler(Fail)

	suiteConfig, reporterConfig := GinkgoConfiguration()

 	webexInfo := ginkgo_helper.WebexInfo{
		AuthToken: "YOUR WEBEX AUTH TOKEN",
		Room:      "YOUR WEBEX ROOM",
	}

	slackInfo := ginkgo_helper.SlackInfo{
		AuthToken: "YOUR SLACK AUTH TOKEN",
		Channel:   "YOUR SLACK CHANNEL",
	}
  
 	elasticInfo := ginkgo_helper.ElasticInfo{
		URL:        "YOUR ELASTIC URL",
		Index:      "YOUR ELASTIC INDEX",
	}
  
	Expect(ginkgo_helper.Register(context.TODO(),
		12345,        // run ID: your CI run id
		false,        // enableLogs: if disabled no logs will printed.
		false,        // dryRun: if enables, this package will print all it would do:
		              // - log what it would store to the provided elastic DB without storing anything;
  		              // - log what message it would send to the provided webex room, without sending any message;
			      // - log what message it would send to the provided slack channel, without sending any message;
 		              // - log what Jira issues it would file, without filing any issue.
		&elasticInfo, // elastic info: test results will be stored into an elastic DB;
		&webexInfo,   // webex info: a webex message will be sent for failed tests in this suite;
		&slackInfo,   // slack info: a slack message will be sent for failed tests in this suite;
		nil,          // no jira. No Jira issue will be failed for failed tests in this suite;
	)).To(Succeed())


	RunSpecs(t, "Controllers Suite", suiteConfig, reporterConfig)
}
```
