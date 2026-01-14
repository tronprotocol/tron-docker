package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	DisplayName   string `json:"-"`
}

func getAccountName(nodeURL string, address string) string {
	url := fmt.Sprintf("%s/wallet/getaccount", nodeURL)
	payload := map[string]interface{}{
		"address": address,
		"visible": true,
	}
	jsonPayload, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error: failed to request account name for %s: %v", address, err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error: failed to read account response body for %s: %v", address, err)
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Error: failed to unmarshal account JSON for %s: %v", address, err)
		return ""
	}

	if val, ok := data["account_name"]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", val)
	}

	return ""
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

	payload := []byte(`{"offset": 0, "limit": 28, "visible": true}`)

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

	// For each witness, try to get the account name in parallel
	log.Printf("Fetching account names for %d witnesses in parallel...\n", len(result.Witnesses))
	var wg sync.WaitGroup
	for i := range result.Witnesses {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			w := &result.Witnesses[idx]
			accName := getAccountName(nodeURL, w.Address)
			if accName != "" {
				if decoded, err := hex.DecodeString(accName); err == nil {
					w.DisplayName = string(decoded)
				} else {
					w.DisplayName = accName
				}
			}
			if w.DisplayName == "" {
				w.DisplayName = w.URL
			}
		}(i)
	}
	wg.Wait()

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

func sendToSlack(webhookURL string, witnesses []Witness, prevVotes map[string]int64, prevSRs []string) error {
	var summary strings.Builder
	summary.WriteString("*TRON SR Status Update (Maintenance Period)*\n")
	summary.WriteString(fmt.Sprintf("Time: %s (UTC)\n\n", time.Now().UTC().Format(time.RFC1123)))

	// 1. Calculate SR changes for the summary
	if len(prevSRs) > 0 {
		currentTop27 := make(map[string]bool)
		for i := 0; i < 27 && i < len(witnesses); i++ {
			currentTop27[witnesses[i].Address] = true
		}

		prevTop27 := make(map[string]bool)
		for _, addr := range prevSRs {
			prevTop27[addr] = true
		}

		var entered []string
		var left []string

		// Who enter
		for i := 0; i < 27 && i < len(witnesses); i++ {
			w := witnesses[i]
			if !prevTop27[w.Address] {
				name := w.DisplayName
				if name == "" {
					name = w.Address
				}
				entered = append(entered, name)
			}
		}

		// Who left
		for _, addr := range prevSRs {
			if !currentTop27[addr] {
				name := addr
				for _, w := range witnesses {
					if w.Address == addr {
						name = w.DisplayName
						if name == "" {
							name = w.Address
						}
						break
					}
				}
				left = append(left, name)
			}
		}

		if len(entered) > 0 || len(left) > 0 {
			summary.WriteString("*SR Replacement Detected:*\n")
			if len(entered) > 0 {
				summary.WriteString(fmt.Sprintf(">:inbox_tray: *Entered:* %s\n", strings.Join(entered, ", ")))
			}
			if len(left) > 0 {
				summary.WriteString(fmt.Sprintf(">:outbox_tray: *Left:* %s\n", strings.Join(left, ", ")))
			}
		} else {
			summary.WriteString("*Top 27 SRs remain unchanged.*")
		}
	} else {
		summary.WriteString("*First check, initializing SR list*")
	}

	// Calculate gap between 27th and 28th SR (always show this if we have enough witnesses)
	if len(witnesses) >= 28 {
		v27 := witnesses[26].VoteCount
		v28 := witnesses[27].VoteCount
		gap := v27 - v28
		summary.WriteString(fmt.Sprintf("\n*Gap (27 vs 28):* `%s` votes", formatComma(gap)))
	}

	// 2. Build detailed list for attachment
	var details strings.Builder
	details.WriteString("```\n")
	for i, w := range witnesses {
		name := w.DisplayName
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

		details.WriteString(fmt.Sprintf("%2d. %-25s Current: %15s  Change: %10s\n",
			i+1, name, formatComma(w.VoteCount), diffStr))
	}
	details.WriteString("```")

	// 3. Construct Payload with Attachments
	payload := map[string]interface{}{
		"text": summary.String(),
		"attachments": []map[string]interface{}{
			{
				"title": "Full Witness List Details",
				"text":  details.String(),
				"color": "#36a64f",
			},
		},
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

func updateLastStatus(witnesses []Witness, lastVotes map[string]int64) []string {
	var top27 []string
	for i, w := range witnesses {
		lastVotes[w.Address] = w.VoteCount
		if i < 27 {
			top27 = append(top27, w.Address)
		}
	}
	return top27
}

func main() {
	// Ensure logs directory exists
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		_ = os.Mkdir(logDir, 0755)
	}

	// Setup logging to both file and stdout
	logPath := filepath.Join(logDir, "sr_monitor.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v, falling back to stdout only\n", err)
	} else {
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	tronNode := os.Getenv("TRON_NODE")
	if tronNode == "" {
		tronNode = DefaultTronNode
	}

	slackWebhook := os.Getenv("SLACK_WEBHOOK")
	if slackWebhook == "" {
		log.Println("Error: SLACK_WEBHOOK environment variable is not set")
		log.Println("Usage: SLACK_WEBHOOK=https://hooks.slack.com/... [TRON_NODE=...] go run main.go")
		os.Exit(1)
	}

	log.Printf("Starting SR monitor.\nNode: %s\nSlack Webhook: %s\n", tronNode, slackWebhook)

	// Map to track votes: Address -> VoteCount
	lastVotes := make(map[string]int64)
	var lastTop27 []string

	// Initial check
	log.Println("Performing initial check...")
	witnesses, err := getWitnessList(tronNode)
	if err != nil {
		log.Printf("Initial check failed: %v\n", err)
	} else {
		log.Printf("Successfully fetched %d witnesses. Sending to Slack...\n", len(witnesses))
		if err := sendToSlack(slackWebhook, witnesses, lastVotes, lastTop27); err != nil {
			log.Printf("Failed to send initial update to Slack: %v\n", err)
		} else {
			lastTop27 = updateLastStatus(witnesses, lastVotes)
		}
	}

	log.Println("Monitoring for maintenance periods via getnextmaintenancetime...")

	for {
		nextTime, err := getNextMaintenanceTime(tronNode)
		if err != nil {
			log.Printf("Error fetching next maintenance time: %v, retrying in 1 minute...\n", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		// Calculate trigger time: next maintenance time + 1 minute
		triggerTime := nextTime.Add(1 * time.Minute)
		now := time.Now().UTC()
		waitDuration := triggerTime.Sub(now)

		if waitDuration > 0 {
			log.Printf("Next maintenance time: %s (UTC). Waiting %v until %s...\n",
				nextTime.Format(time.RFC1123), waitDuration.Truncate(time.Second), triggerTime.Format(time.RFC1123))
			time.Sleep(waitDuration)
		} else {
			log.Printf("Maintenance time %s has already passed. Checking now...\n", nextTime.Format(time.RFC1123))
		}

		log.Printf("[%s] Maintenance period reached (+1m). Fetching SR list...\n", time.Now().UTC().Format(time.RFC3339))
		witnesses, err := getWitnessList(tronNode)
		if err != nil {
			log.Printf("Error fetching witness list: %v\n", err)
		} else {
			if err := sendToSlack(slackWebhook, witnesses, lastVotes, lastTop27); err != nil {
				log.Printf("Error sending to Slack: %v\n", err)
			} else {
				log.Println("Successfully sent SR list to Slack")
				lastTop27 = updateLastStatus(witnesses, lastVotes)
			}
		}

		// Wait a bit before checking for the NEXT maintenance time to avoid double-triggering
		time.Sleep(2 * time.Minute)
	}
}
