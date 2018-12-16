package ansible

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/nlopes/slack"
	"github.com/joho/godotenv"
	l "go-slack-ansible/logger"
	"go.uber.org/zap"
)

func TerminatePlaybook(hosts, envDeploy string) error {
	return terminatePlaybook(hosts, envDeploy)
}

func terminatePlaybook(h, env string) error {
	if err := godotenv.Load(); err != nil {
		l.Logger.Error("[ERROR] Invalid godotenv", zap.Error(err))
		return nil
	}

	var cmd *exec.Cmd
	var args string

	args = fmt.Sprintf("/usr/local/bin/ansible-playbook cancel-ec2.yml --limit=%s -v", h)
	cmd = exec.Command("/bin/bash", "-c", args)
	cmd.Dir = DeployPath

	trace(cmd)

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	defer cmd.Process.Kill()

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from", r)
			}
		}()
		io.WriteString(stdin, "values written to stdin are passed to cmd's standard input")
		_, err := io.Copy(stdin, os.Stdin)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			message := &slack.WebhookMessage{
				Text: line,
			}
			slack.PostWebhook(os.Getenv("SLACK_MONITOR_WEBHOOK_URL"), message)
		}
		if e, ok := err.(*os.PathError); ok && e.Err == syscall.EPIPE {
			io.WriteString(stdin, "values written to stdin are passed to cmd's standard input")
		} else if err != nil {
			log.Println("failed to write to STDIN", err)
		}
		stdin.Close()
		wg.Done()
		G()
	}()
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			message := &slack.WebhookMessage{}
			if strings.Contains(line, "\"failed\": true") {
				message = &slack.WebhookMessage{
					Text: line,
				}
				slack.PostWebhook(os.Getenv("SLACK_MONITOR_WEBHOOK_URL"), message)
			}
		}
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from", r)
			}
		}()
		io.Copy(os.Stdout, stdout)
		stdout.Close()
		wg.Done()
		G()
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from", r)
			}
		}()
		io.Copy(os.Stderr, stderr)
		stderr.Close()
		wg.Done()
		G()
	}()
	wg.Wait()

	return nil
}
