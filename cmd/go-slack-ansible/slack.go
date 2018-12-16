package main

import (
	"go-slack-ansible/ansible"
	"go-slack-ansible/ansible/config"
	"fmt"
	"os"
	"strings"

	l "go-slack-ansible/logger"

	"go.uber.org/zap"

	"github.com/nlopes/slack"
	"github.com/joho/godotenv"
)

// SlackListener is ...
type SlackListener struct {
	client *slack.Client
	botID  string
	config *config.Config
}

var envDeploy string

// ListenAndResponse listens slack events and response
// particular messages. It replies by slack message button.
func (s *SlackListener) ListenAndResponse() {
	rtm := s.client.NewRTM()

	// Start listening slack events
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(ev); err != nil {
				l.Logger.Error("[ERROR] Failed to handle message", zap.Error(err))
			}
		}
	}
}

// handleMesageEvent handles message events.
func (s *SlackListener) handleMessageEvent(ev *slack.MessageEvent) error {
	if err := godotenv.Load(); err != nil {
		l.Logger.Error("[ERROR] Invalid godotenv", zap.Error(err))
		return nil
	}

	channels := []string{
		os.Getenv("SandboxChannelID"),
		os.Getenv("RelaseSTGChannelID"),
		os.Getenv("RelasePRDChannelID"),
	}

	subtypes := []string{
		"bot_message",
		"message_changed",
	}

	// Only response in specific channel. Ignore else.
	if !contains(channels, ev.Channel) {
		l.Logger.Warn(fmt.Sprintf("[WARN] Invalid channel: %s", ev.Channel))
		return nil
	}

	if contains(subtypes, ev.Msg.SubType) {
		l.Logger.Info("[INFO] Bot message")
		return nil
	}

	// Only response mention to bot. Ignore else.
	if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s> ", s.botID)) {
		l.Logger.Warn(fmt.Sprintf("[WARN] Invalid bot: %s", ev.Msg.Text))
		return nil
	}

	// Parse message
	m := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[1:]
	if len(m) == 0 || m[0] != "deploy" {
		return fmt.Errorf("invalid message")
	}

	if setEnv(ev) == "none" {
		l.Logger.Warn(fmt.Sprintf("[WARN] Invalid channel: %s", ev.Channel))
		return nil
	}

	inventoryHosts := []slack.AttachmentActionOption{}
	for _, q := range ansible.InventoryScan(envDeploy) {
		inventoryHosts = append(inventoryHosts, slack.AttachmentActionOption{Text: q, Value: q})
	}

	attachment := slack.Attachment{
		Text:       "Please select the target host.",
		Color:      "#f9a41b",
		CallbackID: "beer",
		Actions: []slack.AttachmentAction{
			{
				Name:    ansible.ActionSelectType,
				Text:    "host",
				Type:    "select",
				Options: inventoryHosts,
			},
			{
				Name:  ansible.ActionCancel,
				Text:  "Cancel",
				Type:  "button",
				Style: "danger",
			},
		},
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			attachment,
		},
	}

	if _, _, err := s.client.PostMessage(ev.Channel, "", params); err != nil {
		return fmt.Errorf("failed to post message: %s", err)
	}

	return nil
}

func setEnv(ev *slack.MessageEvent) string {
	if ev.Channel == os.Getenv("SandboxChannelID") {
		envDeploy = "stg"
	} else if ev.Channel == os.Getenv("RelaseSTGChannelID") {
		envDeploy = "stg"
	} else if ev.Channel == os.Getenv("RelasePRDChannelID") {
		envDeploy = "prd"
	} else {
		return "none"
	}
	return envDeploy
}
