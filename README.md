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
		ginkgo_helper.WithRunID(12345),         
		ginkgo_helper.WithElastic(elasticInfo),
		ginkgo_helper.WithWebex(webexInfo),
		ginkgo_helper.WithSlack(slackInfo),
	)).To(Succeed())


	RunSpecs(t, "Controllers Suite", suiteConfig, reporterConfig)
}
```

## Installing

### dry run

If you want to test info provided are correct, you can invoke ginkgo_helper.Register and set dryRun to true (use WithDryRun setter)
This will:
- log what results it would store into the (if provided) elastic DB without actually storing anything;
- log which message it would send to the (if provided) webex room without actually sending any message;
- log which message it would send to the (if provided) slack channel without actually sending any message;
- log which issues it would file to the (if provided) Jira project/board without actually filing any bug.

### make ut

Another option is to:

1. prepare a config file containing the necessary info. Here is an example
```
    JIRA_BASE_URL: "your jira base url"
    JIRA_PROJECT: "your jira project"
    JIRA_BOARD: "your jira board"
    JIRA_USERNAME: "your jira username"
    JIRA_PASSWORD: "your jira password"

webex:
    WEBEX_AUTH_TOKEN: "your webex auth token"
    WEBEX_ROOM: "your webex room"

slack:
    SLACK_AUTH_TOKEN: "your slack token"
    SLACK_CHANNEL: "your slack channel name"

elastic:
    ELASTIC_URL: "your elastic URL"
    ELASTIC_INDEX: "your elastic index"
~                             
```
2. export TEST_CONFIG_FILE=<file created at step #1 path>
3. make ut

This will verify provided info are correct (no jira bug will be filed, no message will be sent to webex/slack, no result will be stored to elastic DB).


