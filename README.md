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

## Demo

![go-slack-ansible](https://user-images.githubusercontent.com/11452854/51100891-82d9b480-181b-11e9-90ad-55efa5502dd8.gif)

## Ref. Qiita
https://qiita.com/jkkitakita/items/0d6065f14fb81d1a3226
