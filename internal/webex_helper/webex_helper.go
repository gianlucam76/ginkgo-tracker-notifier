package webex_helper

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2" // nolint: golint,stylecheck // ginkgo pattern

	webexteams "github.com/jbogarin/go-cisco-webex-teams/sdk"
)

type WebexInfo struct {
	AuthToken string // webex auth token
	Room      string // webex room
}

// VerifyInfo verifies provided info (webex authorization token and room) are correct
func VerifyInfo(info *WebexInfo) error {
	c := getWebexClient(info.AuthToken)
	room, err := getRoom(c, info.Room)
	if err != nil {
		return fmt.Errorf("failed to get room %s. Err: %s", info.Room, err)
	}
	if room == nil {
		return fmt.Errorf("failed to get room %s", info.Room)
	}
	return nil
}

// getWebexClient returns a Webex client
func getWebexClient(authToken string) *webexteams.Client {
	c := webexteams.NewClient()
	c.SetAuthToken(authToken)
	return c
}

// getRoom returns the Webex room for a given room name
func getRoom(c *webexteams.Client, roomName string) (*webexteams.Room, error) {
	roomQueryParams := &webexteams.ListRoomsQueryParams{
		Max: 200,
	}
	rooms, _, err := c.Rooms.ListRooms(roomQueryParams)
	if err != nil {
		By(fmt.Sprintf("Failed to list rooms %v", err))
		return nil, err
	}

	for i := range rooms.Items {
		if rooms.Items[i].Title == roomName {
			return &rooms.Items[i], nil
		}
	}

	By(fmt.Sprintf("Failed to find room %s", roomName))
	return nil, nil
}

// SendWebexMessage sends webex message to specified room.
// text is a markdown message
func SendWebexMessage(info *WebexInfo, text string) {
	By(fmt.Sprintf("Get room %s", info.Room))
	c := getWebexClient(info.AuthToken)
	room, err := getRoom(c, info.Room)
	if err != nil {
		By(fmt.Sprintf("failed to get room %s. Error: %v", info.Room, err))
		return
	}
	if room == nil {
		By(fmt.Sprintf("failed to get room %s.", info.Room))
		return
	}

	By(fmt.Sprintf("Sending message to room %s", info.Room))
	message := &webexteams.MessageCreateRequest{
		Markdown: text,
		RoomID:   room.ID,
	}
	_, resp, err := c.Messages.CreateMessage(message)
	if err != nil {
		if resp != nil {
			By(fmt.Sprintf("Failed to send message. Error: %v. Response: %s", err, string(resp.Body())))
			return
		}
		By(fmt.Sprintf("Failed to send message. Error: %v", err))
	}
}
