package ansible

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/nlopes/slack"
	"golang.org/x/oauth2"
	"github.com/joho/godotenv"
	l "go-slack-ansible/logger"
	"go.uber.org/zap"
)

func ResponseMessageBranches(w http.ResponseWriter, original slack.Message, repo, host string) {
	original.ReplaceOriginal = true
	remoteBranches := []slack.AttachmentActionOption{}
	for _, q := range listRemoteBranches(repo) {
		remoteBranches = append(remoteBranches, slack.AttachmentActionOption{Text: *q.Name, Value: *q.Name, Description: host})
	}
	original.Attachments[0].Text = fmt.Sprintf("%v", "branchを選択して下さい。:octocat:")
	original.Attachments[0].Actions = []slack.AttachmentAction{
		{
			Name:    ActionDeployConfirm,
			Text:    "branch",
			Type:    "select",
			Value:   host,
			Options: remoteBranches,
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

func listRemoteBranches(repo string) []*github.Branch {
	if err := godotenv.Load(); err != nil {
		l.Logger.Error("[ERROR] Invalid godotenv", zap.Error(err))
		return nil
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("ANSIBLE_ROOT_PATH")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	opt := &github.ListOptions{}

	// list all branches for the authenticated user
	branches, _, err := client.Repositories.ListBranches(ctx, os.Getenv("GITHUB_ORG_NAME"), repo, opt)
	if err != nil {
		return nil
	}
	return branches
}
