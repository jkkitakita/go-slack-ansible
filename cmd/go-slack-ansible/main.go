package main

import (
	"go-slack-ansible/ansible/config"
	l "go-slack-ansible/logger"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
	"go.uber.org/zap"
)

type envConfig struct {
	Port              string `envconfig:"PORT" default:"3000"`
	BotToken          string `envconfig:"BOT_TOKEN" required:"true"`
	VerificationToken string `envconfig:"VERIFICATION_TOKEN" required:"true"`
	BotID             string `envconfig:"BOT_ID" required:"true"`
}

type Server struct {
	envConfig []envConfig
	Logger    *zap.Logger
}

func main() {
	os.Exit(_main(os.Args[1:]))
}

func _main(args []string) int {
	err := godotenv.Load()
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		l.Logger.Error("[ERROR] Invalid envConfig", zap.Error(err))
		return 1
	}

	conf, err := config.LoadConfig()
	if err != nil {
		l.Logger.Error("[ERROR] Invalid config", zap.Error(err))
		return 1
	}

	l.Logger.Info("[INFO] Start slack event listening")
	client := slack.New(env.BotToken)
	slackListener := &SlackListener{
		client: client,
		botID:  env.BotID,
		// channelID: env.ChannelID,
		config: conf,
	}
	go slackListener.ListenAndResponse()

	http.Handle("/interaction", interactionHandler{
		verificationToken: env.VerificationToken,
	})

	l.Logger.Info(fmt.Sprintf("[INFO] Server listening on :%s", env.Port))
	if err := http.ListenAndServe(":"+env.Port, nil); err != nil {
		l.Logger.Error("[ERROR] Invalid ListenAndServe", zap.Error(err))
		return 1
	}

	return 0
}
