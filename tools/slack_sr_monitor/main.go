package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

func setupLogging() {
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		_ = os.Mkdir(logDir, 0755)
	}

	logPath := filepath.Join(logDir, "slack_sr_tools.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		fmt.Printf("Error opening log file: %v, falling back to stdout only\n", err)
	} else {
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	setupLogging()

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	tronNodes := parseTronNodes()
	mode := "both"
	if len(os.Args) > 1 {
		mode = strings.ToLower(os.Args[1])
	}

	switch mode {
	case "monitor", "sr-monitor", "slack-sr-monitor":
		slackWebhook := getSlackWebhook("SLACK_SR_MONITOR_WEBHOOK")
		if slackWebhook == "" {
			log.Println("Error: SLACK_SR_MONITOR_WEBHOOK or SLACK_WEBHOOK environment variable is not set")
			log.Println("Usage: SLACK_SR_MONITOR_WEBHOOK=https://hooks.slack.com/... [TRON_NODES=...] go run . monitor")
			os.Exit(1)
		}
		runSRMonitor(tronNodes, slackWebhook)
	case "guard", "sr-guard", "slack-sr-guard":
		slackWebhook := getSlackWebhook("SLACK_SR_GUARD_WEBHOOK")
		if slackWebhook == "" {
			log.Println("Error: SLACK_SR_GUARD_WEBHOOK or SLACK_WEBHOOK environment variable is not set")
			log.Println("Usage: SLACK_SR_GUARD_WEBHOOK=https://hooks.slack.com/... [TRON_NODES=...] go run . guard")
			os.Exit(1)
		}
		runSRGuard(tronNodes, slackWebhook, getPhoneAlertURL())
	case "both", "":
		monitorWebhook := getSlackWebhook("SLACK_SR_MONITOR_WEBHOOK")
		guardWebhook := getSlackWebhook("SLACK_SR_GUARD_WEBHOOK")
		if monitorWebhook == "" || guardWebhook == "" {
			log.Println("Error: SLACK_SR_MONITOR_WEBHOOK/SLACK_SR_GUARD_WEBHOOK or SLACK_WEBHOOK environment variable is not set")
			log.Println("Usage: SLACK_WEBHOOK=https://hooks.slack.com/... [TRON_NODES=...] go run .")
			os.Exit(1)
		}
		phoneAlertURL := getPhoneAlertURL()

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			runSRMonitor(tronNodes, monitorWebhook)
		}()
		go func() {
			defer wg.Done()
			runSRGuard(tronNodes, guardWebhook, phoneAlertURL)
		}()
		wg.Wait()
	default:
		log.Printf("Error: unknown mode %q", mode)
		log.Println("Usage: go run . [monitor|guard|both]")
		os.Exit(1)
	}
}
