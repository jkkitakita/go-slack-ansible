#!/bin/bash

LOGFILE="./log/app.log"

readonly PROCNAME=${0##*/}
function log() {
  local fname=${BASH_SOURCE[1]##*/}
  echo -e "$(date '+%Y-%m-%dT%H:%M:%#S') ${PROCNAME} (${fname}:${BASH_LINENO[0]}:${FUNCNAME[1]}) $@" | tee -a ${LOGFILE}
}

export GOPATH=/home/ec2-user/go
export PATH=$PATH:/home/ec2-user/go/bin:/usr/local/bin:/bin:/usr/bin

cd $HOME/go/src/go-slack-ansible

git reset --hard origin/master && \
git pull origin master && \
make bin/go-slack-ansible && \
pkill -f "./bin/go-slack-ansible" && \
./bin/go-slack-ansible && \
log "[INFO] go-slack-ansible is restarted..."

exit
