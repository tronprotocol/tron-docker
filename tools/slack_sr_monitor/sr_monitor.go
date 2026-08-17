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
)

const (
	DefaultTronNode     = "https://api.trongrid.io"
	monitorStatePath    = "logs/sr_monitor_state.json"
	monitorStateVersion = 1
)

type monitorState struct {
	Version   int              `json:"version"`
	UpdatedAt time.Time        `json:"updated_at"`
	LastVotes map[string]int64 `json:"last_votes"`
	LastTop27 []string         `json:"last_top_27"`
}

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

func getAccountNameFromNodes(nodeURLs []string, address string) string {
	payload := map[string]interface{}{
		"address": address,
		"visible": true,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error: failed to marshal account name payload for %s: %v", address, err)
		return ""
	}

	var errors []string
	for _, nodeURL := range nodeURLs {
		body, _, err := postToTron([]string{nodeURL}, "/wallet/getaccount", jsonPayload, 5*time.Second)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to unmarshal account JSON: %v", nodeURL, err))
			continue
		}

		if val, ok := data["account_name"]; ok {
			if str, ok := val.(string); ok {
				return str
			}
			return fmt.Sprintf("%v", val)
		}

		return ""
	}

	log.Printf("Error: failed to request account name for %s: %s", address, strings.Join(errors, "; "))
	return ""
}

func parseTronNodes() []string {
	raw := os.Getenv("TRON_NODES")
	if raw == "" {
		raw = os.Getenv("TRON_NODE")
	}
	if raw == "" {
		raw = DefaultTronNode
	}

	var nodes []string
	for _, node := range strings.Split(raw, ",") {
		node = strings.TrimSpace(node)
		if node != "" {
			nodes = append(nodes, strings.TrimRight(node, "/"))
		}
	}
	if len(nodes) == 0 {
		return []string{DefaultTronNode}
	}
	return nodes
}

func getSlackWebhook(specificEnv string) string {
	if webhook := os.Getenv(specificEnv); webhook != "" {
		return webhook
	}
	return os.Getenv("SLACK_WEBHOOK")
}

// getPhoneAlertURL returns the Phone_AWS API Gateway endpoint used by slack_sr_guard to
// trigger a phone call directly. Optional; when unset, the phone alert is skipped.
func getPhoneAlertURL() string {
	return os.Getenv("PHONE_ALERT_URL")
}

func postToTron(nodeURLs []string, path string, payload []byte, timeout time.Duration) ([]byte, string, error) {
	client := &http.Client{Timeout: timeout}
	var errors []string

	for _, nodeURL := range nodeURLs {
		resp, err := client.Post(nodeURL+path, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", nodeURL, err))
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to read response body: %v", nodeURL, readErr))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			errors = append(errors, fmt.Sprintf("%s: status %d: %s", nodeURL, resp.StatusCode, string(body)))
			continue
		}

		return body, nodeURL, nil
	}

	return nil, "", fmt.Errorf("all TRON nodes failed for %s: %s", path, strings.Join(errors, "; "))
}

// WitnessListResponse is the wrapper for the witness list API response
type WitnessListResponse struct {
	Witnesses []Witness `json:"witnesses"`
}

// NextMaintenanceResponse is the wrapper for the maintenance time API response
type NextMaintenanceResponse struct {
	Num int64 `json:"num"`
}

func getNextMaintenanceTimeFromNodes(nodeURLs []string) (time.Time, error) {
	var errors []string
	for _, nodeURL := range nodeURLs {
		body, _, err := postToTron([]string{nodeURL}, "/wallet/getnextmaintenancetime", nil, 10*time.Second)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		var result NextMaintenanceResponse
		if err := json.Unmarshal(body, &result); err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to unmarshal next maintenance JSON: %v", nodeURL, err))
			continue
		}

		return time.Unix(result.Num/1000, (result.Num%1000)*1000000), nil
	}

	return time.Time{}, fmt.Errorf("all TRON nodes failed for /wallet/getnextmaintenancetime: %s", strings.Join(errors, "; "))
}

// NowBlockResponse captures the timestamp of the node's latest block.
type NowBlockResponse struct {
	BlockHeader struct {
		RawData struct {
			Number    int64 `json:"number"`
			Timestamp int64 `json:"timestamp"`
		} `json:"raw_data"`
	} `json:"block_header"`
}

// A fresh TRON node should always be within a few block intervals (block time is 3s).
// Anything beyond this threshold means the node is still catching up to the chain head.
const (
	maxBlockAge            = 30 * time.Second
	freshnessRetryInterval = 2 * time.Minute
	freshnessRetryTimeout  = 2 * time.Hour
	witnessRetryInterval   = 30 * time.Second
	witnessRetryTimeout    = 10 * time.Minute
)

func getLatestBlockTimeFromNodes(nodeURLs []string) (time.Time, error) {
	var errors []string
	for _, nodeURL := range nodeURLs {
		body, _, err := postToTron([]string{nodeURL}, "/wallet/getnowblock", nil, 10*time.Second)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		var result NowBlockResponse
		if err := json.Unmarshal(body, &result); err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to unmarshal now block JSON: %v", nodeURL, err))
			continue
		}

		ts := result.BlockHeader.RawData.Timestamp
		return time.Unix(ts/1000, (ts%1000)*1000000), nil
	}

	return time.Time{}, fmt.Errorf("all TRON nodes failed for /wallet/getnowblock: %s", strings.Join(errors, "; "))
}

func getWitnessListFromNodes(nodeURLs []string) ([]Witness, error) {
	payload := []byte(`{"offset": 0, "limit": 28, "visible": true}`)
	var errors []string
	for _, nodeURL := range nodeURLs {
		body, _, err := postToTron([]string{nodeURL}, "/wallet/getpaginatednowwitnesslist", payload, 10*time.Second)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		witnesses, err := parseWitnessList(nodeURLs, body)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", nodeURL, err))
			continue
		}

		return witnesses, nil
	}

	return nil, fmt.Errorf("all TRON nodes failed for /wallet/getpaginatednowwitnesslist: %s", strings.Join(errors, "; "))
}

func waitForFreshNode(nodeURLs []string) error {
	deadline := time.Now().Add(freshnessRetryTimeout)

	for {
		blockTime, err := getLatestBlockTimeFromNodes(nodeURLs)
		if err != nil {
			log.Printf("Error fetching latest block: %v", err)
		} else if age := time.Since(blockTime); age <= maxBlockAge {
			log.Printf("Node is fresh (latest block at %s, age %s)", blockTime.UTC().Format(time.RFC3339), age.Truncate(time.Second))
			return nil
		} else {
			log.Printf("Node is %s behind chain head (latest block at %s), waiting %s...",
				age.Truncate(time.Second), blockTime.UTC().Format(time.RFC3339), freshnessRetryInterval)
		}

		if time.Now().Add(freshnessRetryInterval).After(deadline) {
			return fmt.Errorf("node did not become fresh within %s", freshnessRetryTimeout)
		}
		time.Sleep(freshnessRetryInterval)
	}
}

func getWitnessListWithRetry(nodeURLs []string) ([]Witness, error) {
	deadline := time.Now().Add(witnessRetryTimeout)

	for {
		witnesses, err := getWitnessListFromNodes(nodeURLs)
		if err == nil {
			return witnesses, nil
		}

		log.Printf("Error fetching witness list: %v", err)
		if time.Now().Add(witnessRetryInterval).After(deadline) {
			return nil, fmt.Errorf("failed to fetch witness list within %s: %v", witnessRetryTimeout, err)
		}
		log.Printf("Retrying witness list in %s...", witnessRetryInterval)
		time.Sleep(witnessRetryInterval)
	}
}

func parseWitnessList(nodeURLs []string, body []byte) ([]Witness, error) {
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
			accName := getAccountNameFromNodes(nodeURLs, w.Address)
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

func sendToSlack(webhookURL string, nodeURLs []string, witnesses []Witness, prevVotes map[string]int64, prevSRs []string) error {
	var summary strings.Builder
	summary.WriteString("*TRON SR Status Update (Maintenance Period)*")
	summary.WriteString(fmt.Sprintf("   Time: %s\n", time.Now().UTC().Format(time.RFC1123)))

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
				foundInCurrentList := false
				for _, w := range witnesses {
					if w.Address == addr {
						foundInCurrentList = true
						name = w.DisplayName
						if name == "" {
							name = w.Address
						}
						break
					}
				}
				if !foundInCurrentList {
					if accountName := getAccountNameFromNodes(nodeURLs, addr); accountName != "" {
						name = accountName
					}
				}
				left = append(left, name)
			}
		}

		// Build rank map for previous SRs
		prevRankMap := make(map[string]int)
		for i, addr := range prevSRs {
			prevRankMap[addr] = i + 1 // Rank is 1-based
		}

		// Detect rank changes within top 27
		var rankChanges []string
		for i := 0; i < 27 && i < len(witnesses); i++ {
			w := witnesses[i]
			if prevRank, ok := prevRankMap[w.Address]; ok {
				change := prevRank - (i + 1) // Positive = rank up, negative = rank down
				if change != 0 {
					direction := "↑"
					if change < 0 {
						direction = "↓"
						change = -change
					}
					name := w.DisplayName
					if name == "" {
						name = w.Address
					}
					rankChanges = append(rankChanges,
						fmt.Sprintf("%s: %d → %d (%s%d)", name, prevRank, i+1, direction, change))
				}
			}
		}

		// Output all changes (entering, leaving, and rank changes) in one block
		if len(entered) > 0 || len(left) > 0 || len(rankChanges) > 0 {
			summary.WriteString("*SR Changes Detected:*\n")
			if len(entered) > 0 {
				summary.WriteString(fmt.Sprintf(">:inbox_tray: *Entered:* %s\n", strings.Join(entered, ", ")))
			}
			if len(left) > 0 {
				summary.WriteString(fmt.Sprintf(">:outbox_tray: *Left:* %s\n", strings.Join(left, ", ")))
			}
			if len(rankChanges) > 0 {
				summary.WriteString(">*Rank Changes:*\n")
				for _, rc := range rankChanges {
					summary.WriteString(fmt.Sprintf(">  `%s`\n", rc))
				}
			}
		} else {
			summary.WriteString("*Top 27 SRs remain unchanged.*\n")
		}
	} else {
		summary.WriteString("*First check, initializing SR list*\n")
	}

	// Calculate gap between 27th and 28th SR (always show this if we have enough witnesses)
	if len(witnesses) >= 28 {
		v27 := witnesses[26].VoteCount
		v28 := witnesses[27].VoteCount
		gap := v27 - v28
		summary.WriteString(fmt.Sprintf("*Gap (27 vs 28):* `%s` votes", formatComma(gap)))
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

func sendSlackText(webhookURL string, text string) error {
	payload := map[string]interface{}{
		"text": text,
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

func loadMonitorState() monitorState {
	state := monitorState{
		Version:   monitorStateVersion,
		LastVotes: make(map[string]int64),
	}

	body, err := os.ReadFile(monitorStatePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to read monitor state: %v", err)
		}
		return state
	}

	if err := json.Unmarshal(body, &state); err != nil {
		log.Printf("Failed to parse monitor state: %v", err)
		return monitorState{Version: monitorStateVersion, LastVotes: make(map[string]int64)}
	}
	if state.LastVotes == nil {
		state.LastVotes = make(map[string]int64)
	}
	log.Printf("Loaded monitor state: votes=%d top27=%d updated_at=%s", len(state.LastVotes), len(state.LastTop27), state.UpdatedAt.UTC().Format(time.RFC3339))
	return state
}

func saveMonitorState(lastVotes map[string]int64, lastTop27 []string) {
	state := monitorState{
		Version:   monitorStateVersion,
		UpdatedAt: time.Now().UTC(),
		LastVotes: lastVotes,
		LastTop27: lastTop27,
	}
	body, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal monitor state: %v", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(monitorStatePath), 0755); err != nil {
		log.Printf("Failed to create monitor state directory: %v", err)
		return
	}
	if err := os.WriteFile(monitorStatePath, body, 0644); err != nil {
		log.Printf("Failed to write monitor state: %v", err)
	}
}

// The final path segment of a Slack webhook URL is the secret token; mask it before logging.
func maskWebhook(url string) string {
	idx := strings.LastIndex(url, "/")
	if idx < 0 || idx == len(url)-1 {
		return "***"
	}
	return url[:idx+1] + "***"
}

func runSRMonitor(tronNodes []string, slackWebhook string) {
	log.Printf("Starting SR monitor.\nNodes: %s\nSlack Webhook: %s\n", strings.Join(tronNodes, ", "), maskWebhook(slackWebhook))

	state := loadMonitorState()
	lastVotes := state.LastVotes
	lastTop27 := state.LastTop27

	if len(lastVotes) > 0 && len(lastTop27) > 0 {
		text := fmt.Sprintf("TRON SR monitor started with existing state. Previous state updated at `%s`. Next report will be sent after the next maintenance period.", state.UpdatedAt.UTC().Format(time.RFC3339))
		if err := sendSlackText(slackWebhook, text); err != nil {
			log.Printf("Failed to send startup log to Slack: %v\n", err)
		} else {
			log.Println("Sent startup log to Slack; keeping existing monitor state until next maintenance report")
		}
	} else {
		log.Println("No complete monitor state found. Performing initial check...")
		witnesses, err := getWitnessListFromNodes(tronNodes)
		if err != nil {
			log.Printf("Initial check failed: %v\n", err)
		} else {
			log.Printf("Successfully fetched %d witnesses. Sending to Slack...\n", len(witnesses))
			if err := sendToSlack(slackWebhook, tronNodes, witnesses, lastVotes, lastTop27); err != nil {
				log.Printf("Failed to send initial update to Slack: %v\n", err)
			} else {
				lastTop27 = updateLastStatus(witnesses, lastVotes)
				saveMonitorState(lastVotes, lastTop27)
			}
		}
	}

	log.Println("Monitoring for maintenance periods via getnextmaintenancetime...")

	for {
		nextTime, err := getNextMaintenanceTimeFromNodes(tronNodes)
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

		log.Printf("[%s] Maintenance period reached (+1m). Waiting for fresh node...\n", time.Now().UTC().Format(time.RFC3339))
		if err := waitForFreshNode(tronNodes); err != nil {
			log.Printf("Skipping this maintenance report: %v\n", err)
			time.Sleep(2 * time.Minute)
			continue
		}

		log.Printf("[%s] Fetching SR list...\n", time.Now().UTC().Format(time.RFC3339))
		witnesses, err := getWitnessListWithRetry(tronNodes)
		if err != nil {
			log.Printf("Skipping this maintenance report: %v\n", err)
		} else {
			if err := sendToSlack(slackWebhook, tronNodes, witnesses, lastVotes, lastTop27); err != nil {
				log.Printf("Error sending to Slack: %v\n", err)
			} else {
				log.Println("Successfully sent SR list to Slack")
				lastTop27 = updateLastStatus(witnesses, lastVotes)
				saveMonitorState(lastVotes, lastTop27)
			}
		}

		// Wait a bit before checking for the NEXT maintenance time to avoid double-triggering
		time.Sleep(2 * time.Minute)
	}
}
