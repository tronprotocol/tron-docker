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
	pollInterval      = time.Minute
	topSRLimit        = 27
	fetchLimit        = 28
	guardStatePath    = "logs/sr_guard_state.json"
	guardStateVersion = 1
)

type guardState struct {
	Version       int               `json:"version"`
	UpdatedAt     time.Time         `json:"updated_at"`
	PreviousTop27 []string          `json:"previous_top_27"`
	NameCache     map[string]string `json:"name_cache"`
}

func getGuardWitnessListFromNodes(nodeURLs []string) ([]Witness, error) {
	payload := []byte(fmt.Sprintf(`{"offset":0,"limit":%d,"visible":true}`, fetchLimit))
	var errors []string
	for _, nodeURL := range nodeURLs {
		body, _, err := postToTron([]string{nodeURL}, "/wallet/getpaginatednowwitnesslist", payload, 10*time.Second)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		var result WitnessListResponse
		if err := json.Unmarshal(body, &result); err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to unmarshal JSON: %v", nodeURL, err))
			continue
		}

		return result.Witnesses, nil
	}

	return nil, fmt.Errorf("all TRON nodes failed for /wallet/getpaginatednowwitnesslist: %s", strings.Join(errors, "; "))
}

func getGuardAccountNameFromNodes(nodeURLs []string, address string) string {
	payload := map[string]interface{}{
		"address": address,
		"visible": true,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("failed to marshal account name payload for %s: %v", address, err)
		return ""
	}

	var errors []string
	for _, nodeURL := range nodeURLs {
		body, _, err := postToTron([]string{nodeURL}, "/wallet/getaccount", jsonPayload, 5*time.Second)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		name, ok := parseAccountName(address, body)
		if !ok {
			errors = append(errors, fmt.Sprintf("%s: invalid account JSON", nodeURL))
			continue
		}

		return name
	}

	log.Printf("failed to request account name for %s: %s", address, strings.Join(errors, "; "))
	return ""
}

func parseAccountName(address string, body []byte) (string, bool) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("failed to unmarshal account JSON for %s: %v", address, err)
		return "", false
	}

	val, ok := data["account_name"]
	if !ok {
		return "", true
	}

	str, ok := val.(string)
	if !ok {
		str = fmt.Sprintf("%v", val)
	}
	decoded, err := hex.DecodeString(str)
	if err != nil {
		return str, true
	}
	return string(decoded), true
}

func fillDisplayNames(nodeURLs []string, witnesses []Witness, nameCache map[string]string) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := range witnesses {
		if name, ok := nameCache[witnesses[i].Address]; ok {
			witnesses[i].DisplayName = name
			continue
		}

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			witness := witnesses[idx]
			name := getGuardAccountNameFromNodes(nodeURLs, witness.Address)
			if name == "" {
				name = witness.URL
			}
			if name == "" {
				name = witness.Address
			}

			mu.Lock()
			nameCache[witness.Address] = name
			witnesses[idx].DisplayName = name
			mu.Unlock()
		}(i)
	}

	wg.Wait()
}

func topAddresses(witnesses []Witness) []string {
	limit := topSRLimit
	if len(witnesses) < limit {
		limit = len(witnesses)
	}

	addresses := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		addresses = append(addresses, witnesses[i].Address)
	}
	return addresses
}

func findTopSRChanges(previous []string, current []Witness) ([]Witness, []string) {
	previousSet := make(map[string]bool, len(previous))
	for _, address := range previous {
		previousSet[address] = true
	}

	currentSet := make(map[string]bool, topSRLimit)
	var entered []Witness
	for i := 0; i < topSRLimit && i < len(current); i++ {
		witness := current[i]
		currentSet[witness.Address] = true
		if !previousSet[witness.Address] {
			entered = append(entered, witness)
		}
	}

	var left []string
	for _, address := range previous {
		if !currentSet[address] {
			left = append(left, address)
		}
	}

	return entered, left
}

func witnessName(w Witness) string {
	if w.DisplayName != "" {
		return w.DisplayName
	}
	if w.URL != "" {
		return w.URL
	}
	return w.Address
}

func sendAlert(webhookURL string, witnesses []Witness, entered []Witness, left []string, nameCache map[string]string) error {
	var text strings.Builder
	text.WriteString("*:warning: TRON Top 27 SR Change Detected*\n")
	text.WriteString(fmt.Sprintf("Time: %s\n", time.Now().UTC().Format(time.RFC1123)))

	if len(entered) > 0 {
		var names []string
		for _, witness := range entered {
			names = append(names, fmt.Sprintf("%s (`%s` votes)", witnessName(witness), formatComma(witness.VoteCount)))
		}
		text.WriteString(fmt.Sprintf(">:inbox_tray: *Entered Top 27:* %s\n", strings.Join(names, ", ")))
	}

	if len(left) > 0 {
		var names []string
		for _, address := range left {
			name := nameCache[address]
			if name == "" {
				name = address
			}
			names = append(names, name)
		}
		text.WriteString(fmt.Sprintf(">:outbox_tray: *Left Top 27:* %s\n", strings.Join(names, ", ")))
	}

	if len(witnesses) >= fetchLimit {
		gap := witnesses[26].VoteCount - witnesses[27].VoteCount
		text.WriteString(fmt.Sprintf("*Gap (27 vs 28):* `%s` votes\n", formatComma(gap)))
		text.WriteString(fmt.Sprintf("*27:* %s `%s` votes\n", witnessName(witnesses[26]), formatComma(witnesses[26].VoteCount)))
		text.WriteString(fmt.Sprintf("*28:* %s `%s` votes\n", witnessName(witnesses[27]), formatComma(witnesses[27].VoteCount)))
	}

	var details strings.Builder
	details.WriteString("```\n")
	for i, witness := range witnesses {
		details.WriteString(fmt.Sprintf("%2d. %-25s %15s\n", i+1, witnessName(witness), formatComma(witness.VoteCount)))
	}
	details.WriteString("```")

	payload := map[string]interface{}{
		"text": text.String(),
		"attachments": []map[string]interface{}{
			{
				"text":  details.String(),
				"color": "#e01e5a",
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

func loadGuardState() guardState {
	state := guardState{
		Version:   guardStateVersion,
		NameCache: make(map[string]string),
	}

	body, err := os.ReadFile(guardStatePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("failed to read guard state: %v", err)
		}
		return state
	}

	if err := json.Unmarshal(body, &state); err != nil {
		log.Printf("failed to parse guard state: %v", err)
		return guardState{Version: guardStateVersion, NameCache: make(map[string]string)}
	}
	if state.NameCache == nil {
		state.NameCache = make(map[string]string)
	}
	log.Printf("loaded guard state: top27=%d names=%d updated_at=%s", len(state.PreviousTop27), len(state.NameCache), state.UpdatedAt.UTC().Format(time.RFC3339))
	return state
}

func saveGuardState(previousTop27 []string, nameCache map[string]string) {
	state := guardState{
		Version:       guardStateVersion,
		UpdatedAt:     time.Now().UTC(),
		PreviousTop27: previousTop27,
		NameCache:     nameCache,
	}
	body, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Printf("failed to marshal guard state: %v", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(guardStatePath), 0755); err != nil {
		log.Printf("failed to create guard state directory: %v", err)
		return
	}
	if err := os.WriteFile(guardStatePath, body, 0644); err != nil {
		log.Printf("failed to write guard state: %v", err)
	}
}

func runSRGuard(tronNodes []string, slackWebhook string) {
	log.Printf("Starting SR guard. Nodes: %s Slack Webhook: %s Poll interval: %s", strings.Join(tronNodes, ", "), maskWebhook(slackWebhook), pollInterval)

	state := loadGuardState()
	nameCache := state.NameCache
	previousTop27 := state.PreviousTop27
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		witnesses, err := getGuardWitnessListFromNodes(tronNodes)
		if err != nil {
			log.Printf("failed to fetch witness list: %v", err)
		} else {
			fillDisplayNames(tronNodes, witnesses, nameCache)

			if len(previousTop27) == 0 {
				previousTop27 = topAddresses(witnesses)
				saveGuardState(previousTop27, nameCache)
				log.Printf("initialized Top 27 snapshot with %d witnesses", len(previousTop27))
			} else {
				entered, left := findTopSRChanges(previousTop27, witnesses)
				if len(entered) > 0 || len(left) > 0 {
					log.Printf("Top 27 change detected: entered=%d left=%d", len(entered), len(left))
					if err := sendAlert(slackWebhook, witnesses, entered, left, nameCache); err != nil {
						log.Printf("failed to send Slack alert: %v", err)
					} else {
						previousTop27 = topAddresses(witnesses)
						saveGuardState(previousTop27, nameCache)
						log.Println("sent Slack alert and updated Top 27 snapshot")
					}
				} else {
					previousTop27 = topAddresses(witnesses)
					saveGuardState(previousTop27, nameCache)
					log.Println("Top 27 unchanged")
				}
			}
		}

		<-ticker.C
	}
}
