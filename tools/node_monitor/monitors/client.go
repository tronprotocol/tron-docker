package monitors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TronClient is a client for interacting with Tron node API
type TronClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewTronClient creates a new Tron API client
func NewTronClient(baseURL string) *TronClient {
	return &TronClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Block represents a Tron block
type Block struct {
	BlockID      string        `json:"blockID"`
	BlockHeader  BlockHeader   `json:"block_header"`
	Transactions []Transaction `json:"transactions"`
}

// BlockHeader represents a Tron block header
type BlockHeader struct {
	RawData BlockRawData `json:"raw_data"`
}

// BlockRawData contains the raw block data
type BlockRawData struct {
	Number         int64  `json:"number"`
	TxTrieRoot     string `json:"txTrieRoot"`
	WitnessAddress string `json:"witness_address"`
	ParentHash     string `json:"parentHash"`
	Version        int32  `json:"version"`
	Timestamp      int64  `json:"timestamp"`
}

// Transaction represents a Tron transaction
type Transaction struct {
	TxID    string      `json:"txID"`
	RawData interface{} `json:"raw_data"`
}

// Witness represents a Super Representative (Witness)
type Witness struct {
	Address        string `json:"address"`
	URL            string `json:"url"`
	TotalProduced  int64  `json:"totalProduced"`
	TotalMissed    int64  `json:"totalMissed"`
	LatestBlockNum int64  `json:"latestBlockNum"`
	LatestSlotNum  int64  `json:"latestSlotNum"`
	IsJobs         bool   `json:"isJobs"`
	VoteCount      int64  `json:"voteCount"`
}

// WitnessListResponse represents the response from listwitnesses API
type WitnessListResponse struct {
	Witnesses []Witness `json:"witnesses"`
}

// GetNowBlock gets the latest block from the node
func (c *TronClient) GetNowBlock() (*Block, error) {
	url := fmt.Sprintf("%s/wallet/getnowblock", c.baseURL)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var block Block
	if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &block, nil
}

// GetBlockByNum gets a block by block number
func (c *TronClient) GetBlockByNum(num int64) (*Block, error) {
	url := fmt.Sprintf("%s/wallet/getblockbynum", c.baseURL)

	reqBody := map[string]interface{}{
		"num": num,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var block Block
	if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &block, nil
}

// ListWitnesses gets the list of witnesses (Super Representatives)
// Use /wallet/getpaginatednowwitnesslist to fetch the top 27 witnesses by voteCount.
func (c *TronClient) ListWitnesses() ([]Witness, error) {
	url := fmt.Sprintf("%s/wallet/getpaginatednowwitnesslist", c.baseURL)

	// Always fetch the first 27 witnesses (active SRs) ordered by voteCount.
	reqBody := map[string]interface{}{
		"offset": 0,
		"limit":  27,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response WitnessListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Witnesses, nil
}
