package slack_helper

import (
	"context"
	"fmt"

	"github.com/gianlucam76/ginkgo-tracker-notifier/internal/utils"
	"github.com/slack-go/slack"
)

type SlackInfo struct {
	AuthToken string // slack auth token
	Channel   string // slack channel name
	DryRun    bool   // indicates if this is a dryRun
}

// VerifyInfo verifies provided info (slack authorization token and channel name) are correct
func VerifyInfo(ctx context.Context, info *SlackInfo) error {
	api := slack.New(info.AuthToken)
	if api == nil {
		return fmt.Errorf("failed to get slack client")
	}

	if _, err := api.AuthTestContext(ctx); err != nil {
		return fmt.Errorf("auth test failed. Err: %v", err)
	}

	if _, err := getChannelID(info); err != nil {
		return err
	}

	return nil
}

// SendSlackMessage sends slack message to specified room.
// text is a markdown message
func SendSlackMessage(info *SlackInfo, text string) {
	utils.Byf(fmt.Sprintf("Get channel ID %s", info.Channel))
	api := slack.New(info.AuthToken)
	if api == nil {
		utils.Byf("failed to get slack client")
	}

	utils.Byf(fmt.Sprintf("Sending message to channel %s", info.Channel))

	channelID, err := getChannelID(info)
	if err != nil {
		utils.Byf(fmt.Sprintf("failed to get channel %s", info.Channel))
		return
	}

	if info.DryRun {
		utils.Byf("Send message %q to channel %s", text, info.Channel)
		return
	}

	_, _, err = api.PostMessage(channelID, slack.MsgOptionText(text, false))
	if err != nil {
		utils.Byf(fmt.Sprintf("Failed to send message. Error: %v", err))
	}
}

func getChannelID(info *SlackInfo) (string, error) {
	api := slack.New(info.AuthToken)
	for {
		cursor := ""
		channels, nextCursor, err := api.GetConversations(&slack.GetConversationsParameters{
			ExcludeArchived: true,
			Cursor:          cursor,
			Limit:           50,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get channels. Err: %v", err)
		}

		for i := range channels {
			if channels[i].Name == info.Channel {
				return channels[i].ID, nil
			}
		}

		cursor = nextCursor
		if cursor == "" {
			return "", fmt.Errorf("failed to find channel %s", info.Channel)
		}
	}
}
