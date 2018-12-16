# go-slack-ansible

Slackbot to execute "ansible-playbook" as "deployment tool" from slack.

## Requirements

1. Tools and language
   - Slack
   - ansible（>= 2.3.1）
   - ansible-playbook
   - golang (>= 1.11)
2. Use "ansible-playbook" as deployment tools.

## Getting Started

1. Copy .env to .env.local on server to execute "ansible-playbook".
2. Add env variables for using "Slack Interactive messages" and "ansible-playbook" to .env.local.
https://api.slack.com/interactive-messages
3. Setup golang
```
make deps
```
4. Run Slack bot
```
make run

or

make build && ./bin/go-slack-ansible
```
