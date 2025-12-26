package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	DefaultTronNode = "https://api.trongrid.io"
)

// Witness represents the structure of a witness returned by the Tron API
type Witness struct {
	Address       string `json:"address"`
	VoteCount     int64  `json:"voteCount"`
	PubKey        string `json:"pubKey"`
	URL           string `json:"url"`
	TotalProduced int64  `json:"totalProduced"`
	TotalMissed   int64  `json:"totalMissed"`
	LatestBlock   int64  `json:"latestBlockNum"`
	IsJobs        bool   `json:"isJobs"`
}

// WitnessListResponse is the wrapper for the witness list API response
type WitnessListResponse struct {
	Witnesses []Witness `json:"witnesses"`
}

// NextMaintenanceResponse is the wrapper for the maintenance time API response
type NextMaintenanceResponse struct {
	Num int64 `json:"num"`
}

func getNextMaintenanceTime(nodeURL string) (time.Time, error) {
	url := fmt.Sprintf("%s/wallet/getnextmaintenancetime", nodeURL)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var result NextMaintenanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return time.Time{}, err
	}

	// Result is in milliseconds
	return time.Unix(result.Num/1000, (result.Num%1000)*1000000), nil
}

func getWitnessList(nodeURL string) ([]Witness, error) {
	url := fmt.Sprintf("%s/wallet/getpaginatednowwitnesslist", nodeURL)
	// Fetch top 28 SRs
	payload := []byte(`{"offset": 0, "limit": 28}`)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to call Tron API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tron API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var result WitnessListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v (body: %s)", err, string(body))
	}

	return result.Witnesses, nil
}

func formatComma(n int64) string {
	in := fmt.Sprintf("%d", n)
	var out strings.Builder
	if n < 0 {
		out.WriteByte('-')
		in = in[1:]
	}
	l := len(in)
	for i, c := range in {
		if i > 0 && (l-i)%3 == 0 {
			out.WriteByte(',')
		}
		out.WriteRune(c)
	}
	return out.String()
}

func sendToSlack(webhookURL string, witnesses []Witness, prevVotes map[string]int64) error {
	var buffer bytes.Buffer
	buffer.WriteString("*TRON SR Status Update (Maintenance Period)*\n")
	buffer.WriteString(fmt.Sprintf("Time: %s (UTC)\n\n", time.Now().UTC().Format(time.RFC1123)))

	buffer.WriteString("```\n")
	buffer.WriteString(fmt.Sprintf("%-3s | %-30s | %-15s | %-15s | %-8s\n", "#", "SR Name/URL", "Current Votes", "Prev Votes", "Change"))
	buffer.WriteString("-----------------------------------------------------------------------------------------------------\n")

	for i, w := range witnesses {
		name := w.URL
		if len(name) > 30 {
			name = name[:27] + "..."
		}
		if name == "" {
			name = w.Address
		}

		prev := prevVotes[w.Address]
		diff := w.VoteCount - prev

		diffStr := formatComma(diff)
		if diff >= 0 {
			diffStr = "+" + diffStr
		}

		if prev == 0 {
			diffStr = "-"
		}

		prevStr := formatComma(prev)
		if prev == 0 {
			prevStr = "-"
		}

		buffer.WriteString(fmt.Sprintf("%-3d | %-30s | %-15s | %-15s | %-8s\n",
			i+1, name, formatComma(w.VoteCount), prevStr, diffStr))
	}
	buffer.WriteString("```\n")

	payload := map[string]string{
		"text": buffer.String(),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send to slack: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: No .env file found, using system environment variables")
	}

	tronNode := os.Getenv("TRON_NODE")
	if tronNode == "" {
		tronNode = DefaultTronNode
	}

	slackWebhook := os.Getenv("SLACK_WEBHOOK")
	if slackWebhook == "" {
		fmt.Println("Error: SLACK_WEBHOOK environment variable is not set")
		fmt.Println("Usage: SLACK_WEBHOOK=https://hooks.slack.com/... [TRON_NODE=...] go run main.go")
		os.Exit(1)
	}

	fmt.Printf("Starting SR monitor.\nNode: %s\nSlack Webhook: %s\n", tronNode, "[REDACTED]")

	// Map to track votes: Address -> VoteCount
	lastVotes := make(map[string]int64)

	updateLastVotes := func(witnesses []Witness) {
		for _, w := range witnesses {
			lastVotes[w.Address] = w.VoteCount
		}
	}

	// Initial check
	fmt.Println("Performing initial check...")
	witnesses, err := getWitnessList(tronNode)
	if err != nil {
		fmt.Printf("Initial check failed: %v\n", err)
	} else {
		fmt.Printf("Successfully fetched %d witnesses. Sending to Slack...\n", len(witnesses))
		if err := sendToSlack(slackWebhook, witnesses, lastVotes); err != nil {
			fmt.Printf("Failed to send initial update to Slack: %v\n", err)
		} else {
			fmt.Println("Initial update sent successfully.")
			updateLastVotes(witnesses)
		}
	}

	fmt.Println("Monitoring for maintenance periods via getnextmaintenancetime...")

	for {
		nextTime, err := getNextMaintenanceTime(tronNode)
		if err != nil {
			fmt.Printf("Error fetching next maintenance time: %v, retrying in 1 minute...\n", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		// Calculate trigger time: next maintenance time + 1 minute
		triggerTime := nextTime.Add(1 * time.Minute)
		now := time.Now().UTC()
		waitDuration := triggerTime.Sub(now)

		if waitDuration > 0 {
			fmt.Printf("Next maintenance time: %s (UTC). Waiting %v until %s...\n",
				nextTime.Format(time.RFC1123), waitDuration.Truncate(time.Second), triggerTime.Format(time.RFC1123))
			time.Sleep(waitDuration)
		} else {
			fmt.Printf("Maintenance time %s has already passed. Checking now...\n", nextTime.Format(time.RFC1123))
		}

		fmt.Printf("[%s] Maintenance period reached (+1m). Fetching SR list...\n", time.Now().UTC().Format(time.RFC3339))
		witnesses, err := getWitnessList(tronNode)
		if err != nil {
			fmt.Printf("Error fetching witness list: %v\n", err)
		} else {
			if err := sendToSlack(slackWebhook, witnesses, lastVotes); err != nil {
				fmt.Printf("Error sending to Slack: %v\n", err)
			} else {
				fmt.Println("Successfully sent SR list to Slack")
				updateLastVotes(witnesses)
			}
		}

		// Wait a bit before checking for the NEXT maintenance time to avoid double-triggering
		time.Sleep(2 * time.Minute)
	}
}
