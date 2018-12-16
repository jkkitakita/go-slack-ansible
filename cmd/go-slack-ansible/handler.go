package main

import (
	"go-slack-ansible/ansible"
	"go-slack-ansible/ansible/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	l "go-slack-ansible/logger"

	"go.uber.org/zap"

	"github.com/nlopes/slack"
)

// interactionHandler handles interactive message response.
type interactionHandler struct {
	slackClient       *slack.Client
	verificationToken string
	config            *config.Config
}

// MySlackMessage is ...
type MySlackMessage struct {
	slack.Message
}

func (h interactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		l.Logger.Error(fmt.Sprintf("[ERROR] Invalid method", r.Method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Logger.Error("[ERROR] Failed to read request body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		l.Logger.Error("[ERROR] Failed to unespace request body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.AttachmentActionCallback
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		l.Logger.Error("[ERROR] Failed to decode json message from slack", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Only accept message from slack with valid token
	if message.Token != h.verificationToken {
		l.Logger.Error(fmt.Sprintf("[ERROR] Invalid token: %s", message.Token))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conf, err := config.LoadConfig()
	if err != nil {
		l.Logger.Error("[ERROR] Invalid config", zap.Error(err))
		return
	}

	action := message.Actions[0]
	switch originalMessage := message.OriginalMessage; action.Name {
	case ansible.ActionSelectType:
		selectValue := action.SelectedOptions[0].Value
		// Overwrite original drop down message.
		originalMessage.ReplaceOriginal = true
		originalMessage.Attachments[0].Text = fmt.Sprintln("選んでください。 :thinking_face: ")
		attachmentActionText := originalMessage.Attachments[0].Actions[0].Text
		updateAttachmentFields(&MySlackMessage{originalMessage}, attachmentActionText, selectValue)
		originalMessage.Attachments[0].Actions = []slack.AttachmentAction{
			{
				Name:    ansible.ActionDeployConfirm,
				Text:    "デプロイしますか？",
				Type:    "button",
				Style: 	 "primary",
			},
			{
				Name:    ansible.ActionTerminateConfirm,
				Text:    "terminateしますか？",
				Type:    "button",
				Style:   "danger",
			},
			{
				Name:  ansible.ActionCancel,
				Text:  "Cancel",
				Type:  "button",
				Style: "danger",
			},
		}

		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(originalMessage)
		return
	case ansible.ActionDeployConfirm:
		// Overwrite original drop down message.
		originalMessage.ReplaceOriginal = true
		originalMessage.Attachments[0].Text = fmt.Sprintln("deploy start! OK? :thinking_face: ")
		var attachmentActionText, selectValue string
		if originalMessage.Attachments[0].Actions[0].Type == "select" {
			attachmentActionText = originalMessage.Attachments[0].Actions[0].Text
			selectValue = action.SelectedOptions[0].Value
		}
		updateAttachmentFields(&MySlackMessage{originalMessage}, attachmentActionText, selectValue)
		originalMessage.Attachments[0].Actions = updateAttachmentAction(w, message.OriginalMessage)
		
		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(originalMessage)
		return
	case ansible.ActionTerminateConfirm:
		// Overwrite original drop down message.
		originalMessage.ReplaceOriginal = true
		originalMessage.Attachments[0].Text = fmt.Sprintln("terminate start! OK? :thinking_face: ")
		originalMessage.Attachments[0].Actions = []slack.AttachmentAction{
			{
				Name:  ansible.ActionTerminateStart,
				Text:  "terminate",
				Type:  "button",
				Style: "danger",
			},
			{
				Name:  ansible.ActionCancel,
				Text:  "Cancel",
				Type:  "button",
				Style: "primary",
			},
		}
		
		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(originalMessage)
		return
	case ansible.ActionTerminateStart:
		originalMessage.ReplaceOriginal = true
		title := fmt.Sprintf(":o: <@%s> terminate start!!", message.User.ID)
		var host string
		for _, f := range originalMessage.Attachments[0].Fields {
			switch {
			case f.Title == "host":
				host = f.Value
			default:
				fmt.Sprintf("Field of [%s] not found.\n", f.Title)
			}
		}
		responseMessage(w, message.OriginalMessage, title, "")
		go ansible.TerminatePlaybook(host, envDeploy)
		return
	case ansible.ActionDeployStart:
		originalMessage.ReplaceOriginal = true
		title := fmt.Sprintf(":o: <@%s> deploy start!!", message.User.ID)
		var host, branch, tags, skipTags string
		for _, f := range originalMessage.Attachments[0].Fields {
			switch {
			case f.Title == "host":
				host = f.Value
			case f.Title == "branch":
				branch = f.Value
			case f.Title == "tags":
				tags = f.Value
			case f.Title == "skipTags":
				skipTags = f.Value
			default:
				fmt.Sprintf("Field of [%s] not found.\n", f.Title)
			}
		}
		responseMessage(w, message.OriginalMessage, title, "")
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Recovered from", r)
				}
			}()
			ansible.RunPlaybook(host, branch, tags, skipTags)
			ansible.G()
		}()

		return
	case ansible.ActionCancel:
		title := fmt.Sprintf(":x: <@%s> canceled the request", message.User.ID)
		responseMessage(w, message.OriginalMessage, title, "")
		return
	case ansible.ActionSelectBranch:
		originalMessage.ReplaceOriginal = true
		var host string
		for _, f := range originalMessage.Attachments[0].Fields {
			switch {
			case f.Title == "host":
				host = f.Value
			default:
				fmt.Sprintf("Field of [%s] not found.\n", f.Title)
			}
		}
		for _, repo := range conf.HostAndRepos {
			if repo.Host == strings.ToLower(host) {
				ansible.ResponseMessageBranches(w, message.OriginalMessage, repo.Repo, host)
				return
			}
		}
		return
	case ansible.ActionSelectTag:
		taskTags := []slack.AttachmentActionOption{}
		for _, t := range conf.TaskTags {
			taskTags = append(taskTags, slack.AttachmentActionOption{Text: t, Value: t})
		}
		ansible.ResponseMessageTags(w, message.OriginalMessage, taskTags)
		return
	case ansible.ActionAddTag:
		taskTags := []slack.AttachmentActionOption{}
		for _, t := range conf.TaskTags {
			taskTags = append(taskTags, slack.AttachmentActionOption{Text: t, Value: t})
		}
		ansible.ResponseMessageTags(w, message.OriginalMessage, taskTags)
		return
	case ansible.ActionSelectSkipTag:
		taskSkipTags := []slack.AttachmentActionOption{}
		for _, t := range conf.TaskTags {
			taskSkipTags = append(taskSkipTags, slack.AttachmentActionOption{Text: t, Value: t})
		}
		ansible.ResponseMessageSkipTags(w, message.OriginalMessage, taskSkipTags)
		return
	case ansible.ActionAddSkipTag:
		taskSkipTags := []slack.AttachmentActionOption{}
		for _, t := range conf.TaskTags {
			taskSkipTags = append(taskSkipTags, slack.AttachmentActionOption{Text: t, Value: t})
		}
		ansible.ResponseMessageSkipTags(w, message.OriginalMessage, taskSkipTags)
		return
	default:
		l.Logger.Error(fmt.Sprintf("[ERROR] Invalid action was submitted", action.Name))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// responseMessage response to the original slackbutton enabled message.
// It removes button and replace it with message which indicate how bot will work
func responseMessage(w http.ResponseWriter, original slack.Message, title, value string) {
	original.ReplaceOriginal = true
	original.Attachments[0].Text = fmt.Sprintf("%v", title)
	original.Attachments[0].Actions = []slack.AttachmentAction{}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)
}

func updateAttachmentFields(original *MySlackMessage, attachmentActionText, selectValue string) {
	attachmentActionTexts := make([]string, 0, 10)
	for _, t := range original.Attachments[0].Fields {
		attachmentActionTexts = append(attachmentActionTexts, t.Title)
	}
	if !contains(attachmentActionTexts, attachmentActionText) && attachmentActionText != "" && selectValue != "" {
		original.Attachments[0].Fields = append(original.Attachments[0].Fields, slack.AttachmentField{Title: attachmentActionText, Value: selectValue, Short: true})
	}
	for _, f := range original.Attachments[0].Fields {
		if attachmentActionText == f.Title {
			f.Value = selectValue
		}
	}
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

func updateAttachmentAction(w http.ResponseWriter, original slack.Message) []slack.AttachmentAction {
	var originalFieldsTitles []string
	var originalAttachmentAction []slack.AttachmentAction
	originalFields := original.Attachments[0].Fields
	for _, r := range originalFields {
		originalFieldsTitles = append(originalFieldsTitles, r.Title)
	}
	if envDeploy == "it" {
		originalAttachmentAction = []slack.AttachmentAction{
			{ Name:  ansible.ActionDeployStart, Text:  "deploy", Type:  "button", Style: "primary" },
			{ Name:  ansible.ActionCancel, Text:  "cancel", Type:  "button", Style: "danger" },
			{ Name:  ansible.ActionSelectBranch, Text:  "branch", Type:  "button"},
			{ Name:  ansible.ActionSelectTag, Text:  "tags", Type:  "button"},
			{ Name:  ansible.ActionSelectSkipTag, Text:  "skipTags", Type:  "button"},
		}
	} else {
		originalAttachmentAction = []slack.AttachmentAction{
			{ Name:  ansible.ActionDeployStart, Text:  "deploy", Type:  "button", Style: "primary" },
			{ Name:  ansible.ActionCancel, Text:  "cancel", Type:  "button", Style: "danger" },
			{ Name:  ansible.ActionSelectTag, Text:  "tags", Type:  "button"},
			{ Name:  ansible.ActionSelectSkipTag, Text:  "skipTags", Type:  "button"},
		}
	}
	for i := 0; i < len(originalAttachmentAction); i++ {
		if contains(originalFieldsTitles, originalAttachmentAction[i].Text) {
			originalAttachmentAction[i].Text = ""
		}
	}
	return originalAttachmentAction
}
