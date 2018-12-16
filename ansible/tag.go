package ansible

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nlopes/slack"
)

func ResponseMessageTags(w http.ResponseWriter, original slack.Message, taskTags []slack.AttachmentActionOption) {
	original.ReplaceOriginal = true
	original.Attachments[0].Text = fmt.Sprintf("%v", "tagを選択して下さい :label:")
	original.Attachments[0].Actions = []slack.AttachmentAction{
		{
			Name:    ActionDeployConfirm,
			Text:    "tags",
			Type:    "select",
			Value:   "tag",
			Options: taskTags,
		},
		{
			Name:  ActionCancel,
			Text:  "No",
			Type:  "button",
			Style: "danger",
		},
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)
}

func ResponseMessageSkipTags(w http.ResponseWriter, original slack.Message, taskSkipTags []slack.AttachmentActionOption) {
	original.ReplaceOriginal = true
	original.Attachments[0].Text = fmt.Sprintf("%v", "スキップするtagを選択して下さい :label: :dash: ")
	original.Attachments[0].Actions = []slack.AttachmentAction{
		{
			Name:    ActionDeployConfirm,
			Text:    "skipTags",
			Type:    "select",
			Value:   "tag",
			Options: taskSkipTags,
		},
		{
			Name:  ActionCancel,
			Text:  "No",
			Type:  "button",
			Style: "danger",
		},
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)
}
