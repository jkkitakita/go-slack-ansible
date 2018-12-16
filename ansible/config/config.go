package config

import (
	"io/ioutil"

	l "go-slack-ansible/logger"

	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	HostAndRepos []HostAndRepo `yaml:"host_and_repos"`
	TaskTags     []string      `yaml:"task_tags"`
}

type HostAndRepo struct {
	Host string `yaml:"host"`
	Repo string `yaml:"repo"`
}

// type TaskTag struct {
// 	Tag string `yaml:"tag"`
// }

func LoadConfig() (*Config, error) {
	buf, err := ioutil.ReadFile("./ansible/config.yaml")
	if err != nil {
		l.Logger.Error("[ERROR] Failed read config file", zap.Error(err))
		return nil, err
	}
	var config Config
	if err = yaml.Unmarshal(buf, &config); err != nil {
		l.Logger.Error("[ERROR] Failed yaml unmarshal", zap.Error(err))
		return nil, err
	}
	return &config, nil
}
