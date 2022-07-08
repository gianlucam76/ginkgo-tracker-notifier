package process_result_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gopkg.in/yaml.v2"

	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/utils"
	"github.com/gianlucam76/ginkgo-tracker-notifier/process_result"
)

const (
	// Envirnoment variable containing ut configuration.
	CONFIG_FILE = "TEST_CONFIG_FILE"

	ELASTIC_URL   = "ELASTIC_URL"
	ELASTIC_INDEX = "ELASTIC_INDEX"

	WEBEX_AUTH_TOKEN = "WEBEX_AUTH_TOKEN"
	WEBEX_ROOM       = "WEBEX_ROOM"

	SLACK_AUTH_TOKEN = "SLACK_AUTH_TOKEN"
	SLACK_CHANNEL    = "SLACK_CHANNEL"

	JIRA_BASE_URL = "JIRA_BASE_URL"
	JIRA_PROJECT  = "JIRA_PROJECT"
	JIRA_BOARD    = "JIRA_BOARD"
	JIRA_USERNAME = "JIRA_USERNAME"
	JIRA_PASSWORD = "JIRA_PASSWORD"
)

// utConfig defines the configuration of an unit test environment.
type utConfig struct {
	// Jira contains jira configuration
	Jira map[string]string `json:"jira,omitempty"`

	// Webex contains webex configuration
	Webex map[string]string `json:"webex,omitempty"`

	// Slack contains slack configuration
	Slack map[string]string `json:"slack,omitempty"`

	// Elastic contains elastic configuration
	Elastic map[string]string `json:"elastic,omitempty"`
}

var _ = Describe("Test verify methods", func() {
	config := &utConfig{}
	BeforeEach(func() {
		utils.Init(true)
		config = prepareEnvironment()
	})

	It("VerifyElasticInfo reports no error when correct info are provided", Label("ELASTIC"), func() {
		Expect(config.Elastic).ToNot(BeEmpty())
		c := &process_result.Options{}
		c.ElasticInfo = &process_result.ElasticInfo{
			URL:   config.GetElasticVariable(ELASTIC_URL),
			Index: config.GetElasticVariable(ELASTIC_INDEX),
		}
		Expect(process_result.VerifyElasticInfo(context.TODO(), c)).To(BeNil())
	})

	It("VerifyElasticInfo reports an error when URL is incorrect", Label("ELASTIC"), func() {
		Expect(config.Elastic).ToNot(BeEmpty())
		elastUrl := "issues.elastic.org/"
		c := &process_result.Options{}
		c.ElasticInfo = &process_result.ElasticInfo{
			URL:   elastUrl,
			Index: config.GetElasticVariable(ELASTIC_INDEX),
		}
		Expect(process_result.VerifyElasticInfo(context.TODO(), c)).ToNot(BeNil())
	})

	It("VerifyElasticInfo reports an error when index does not exist", Label("ELASTIC"), func() {
		Expect(config.Elastic).ToNot(BeEmpty())
		elastIndex := "1234-abcd"
		c := &process_result.Options{}
		c.ElasticInfo = &process_result.ElasticInfo{
			URL:   config.GetElasticVariable(ELASTIC_URL),
			Index: elastIndex,
		}
		Expect(process_result.VerifyElasticInfo(context.TODO(), c)).ToNot(BeNil())
	})

	It("VerifySlackInfo reports no error when correct info are provided", Label("SLACK"), func() {
		Expect(config.Slack).ToNot(BeEmpty())
		c := &process_result.Options{}
		c.SlackInfo = &process_result.SlackInfo{
			AuthToken: config.GetSlackVariable(SLACK_AUTH_TOKEN),
			Channel:   config.GetSlackVariable(SLACK_CHANNEL),
		}
		Expect(process_result.VerifySlackInfo(context.TODO(), c)).To(BeNil())
	})

	It("VerifySlackInfo reports error when wrong token is provided", Label("SLACK"), func() {
		Expect(config.Slack).ToNot(BeEmpty())
		slackAuthToken := "abc"
		c := &process_result.Options{}
		c.SlackInfo = &process_result.SlackInfo{
			AuthToken: slackAuthToken,
			Channel:   config.GetSlackVariable(SLACK_CHANNEL),
		}
		Expect(process_result.VerifySlackInfo(context.TODO(), c)).ToNot(BeNil())
	})

	It("VerifySlackInfo reports error when wrong channel name is provided", Label("SLACK"), func() {
		Expect(config.Slack).ToNot(BeEmpty())
		c := &process_result.Options{}
		c.SlackInfo = &process_result.SlackInfo{
			AuthToken: config.GetSlackVariable(SLACK_AUTH_TOKEN),
			Channel:   "non-existing",
		}
		Expect(process_result.VerifySlackInfo(context.TODO(), c)).ToNot(BeNil())
	})

	It("VerifyWebexInfo reports no error when correct info are provided", Label("WEBEX"), func() {
		Expect(config.Webex).ToNot(BeEmpty())
		c := &process_result.Options{}
		c.WebexInfo = &process_result.WebexInfo{
			AuthToken: config.GetWebexVariable(WEBEX_AUTH_TOKEN),
			Room:      config.GetWebexVariable(WEBEX_ROOM),
		}
		Expect(process_result.VerifyWebexInfo(context.TODO(), c)).To(BeNil())
	})

	It("VerifyWebexInfo reports error when auth token is incorrect", Label("WEBEX"), func() {
		Expect(config.Webex).ToNot(BeEmpty())
		webexAuthToken := "123"
		c := &process_result.Options{}
		c.WebexInfo = &process_result.WebexInfo{
			AuthToken: webexAuthToken,
			Room:      config.GetWebexVariable(WEBEX_ROOM),
		}
		Expect(process_result.VerifyWebexInfo(context.TODO(), c)).ToNot(BeNil())
	})

	It("VerifyWebexInfo reports error when room is incorrect", Label("WEBEX"), func() {
		Expect(config.Webex).ToNot(BeEmpty())
		webexRoom := "test-1234-abcd"
		c := &process_result.Options{}
		c.WebexInfo = &process_result.WebexInfo{
			AuthToken: config.GetWebexVariable(WEBEX_AUTH_TOKEN),
			Room:      webexRoom,
		}
		Expect(process_result.VerifyWebexInfo(context.TODO(), c)).ToNot(BeNil())
	})

	It("VerifyJiraInfo reports no error when correct info are provided", Label("JIRA"), func() {
		Expect(config.Jira).ToNot(BeEmpty())
		c := &process_result.Options{}
		c.JiraInfo = &process_result.JiraInfo{
			BaseURL:  config.GetJiraVariable(JIRA_BASE_URL),
			Project:  config.GetJiraVariable(JIRA_PROJECT),
			Board:    config.GetJiraVariable(JIRA_BOARD),
			Username: config.GetJiraVariable(JIRA_USERNAME),
			Password: config.GetJiraVariable(JIRA_PASSWORD),
		}
		Expect(process_result.VerifyJiraInfo(context.TODO(), c)).To(BeNil())
	})
})

// Use this method to set all necessary environment variables
// Use ginkgo label filter to select tests you want to run. Then
// only sets necessary environment variables
func prepareEnvironment() *utConfig {
	configFilename, ok := os.LookupEnv(CONFIG_FILE)
	Expect(ok).To(BeTrue())
	return loadE2EConfig(configFilename)
}

// loadE2EConfig loads the configuration for unit tests
func loadE2EConfig(input string) *utConfig {
	configData, err := os.ReadFile(input)
	Expect(err).ToNot(HaveOccurred(),
		"Failed to read the e2e test config file")

	Expect(configData).ToNot(BeEmpty(),
		"The e2e test config file should not be empty")

	config := &utConfig{}
	Expect(yaml.Unmarshal(configData, config)).To(Succeed(),
		"Failed to convert the e2e test config file to yaml")

	return config
}

// GetJiraVariable returns a Jira variable from ut config file.
func (c *utConfig) GetJiraVariable(varName string) string {
	value, ok := c.Jira[varName]
	Expect(ok).NotTo(BeFalse())
	return value
}

// GetElasticVariable returns an Elastic variable from the ut config file.
func (c *utConfig) GetElasticVariable(varName string) string {
	value, ok := c.Elastic[varName]
	Expect(ok).NotTo(BeFalse())
	return value
}

// GetWebexVariable returns a variable from the ut config file.
func (c *utConfig) GetWebexVariable(varName string) string {
	value, ok := c.Webex[varName]
	Expect(ok).NotTo(BeFalse())
	return value
}

// GetSlackVariable returns a variable from the ut config file.
func (c *utConfig) GetSlackVariable(varName string) string {
	value, ok := c.Slack[varName]
	Expect(ok).NotTo(BeFalse())
	return value
}
