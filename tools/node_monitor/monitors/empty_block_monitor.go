package monitors

import (
	"context"
	"fmt"
	"log"
	"time"
)

// EmptyBlockMonitor monitors for empty blocks
type EmptyBlockMonitor struct {
	client       *TronClient
	interval     time.Duration
	lastBlockNum int64
	nodeLabel    string
}

// NewEmptyBlockMonitor creates a new empty block monitor
func NewEmptyBlockMonitor(client *TronClient, interval time.Duration, nodeLabel string) *EmptyBlockMonitor {
	return &EmptyBlockMonitor{
		client:    client,
		interval:  interval,
		nodeLabel: nodeLabel,
	}
}

// Start starts the empty block monitoring loop
func (m *EmptyBlockMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	log.Println("Empty block monitor started")

	// Do an initial check
	m.checkBlock(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Empty block monitor stopped")
			return
		case <-ticker.C:
			m.checkBlock(ctx)
		}
	}
}

// checkBlock checks blocks for emptiness
// It checks all blocks from lastBlockNum+1 to the latest block to avoid missing blocks
// On first run (lastBlockNum == 0), it only checks the current latest block to initialize
func (m *EmptyBlockMonitor) checkBlock(ctx context.Context) {
	// Get the latest block to know the current block number
	latestBlock, err := m.client.GetNowBlock()
	if err != nil {
		log.Printf("Failed to get latest block: %v", err)
		NodeMetrics.APICallErrors.WithLabelValues(m.nodeLabel, "getnowblock", "request_failed").Inc()
		return
	}

	latestBlockNum := latestBlock.BlockHeader.RawData.Number

	// Update last block metrics with the latest block info
	EmptyBlockMetrics.LastBlockNumber.WithLabelValues(m.nodeLabel).Set(float64(latestBlockNum))
	EmptyBlockMetrics.LastBlockTimestamp.WithLabelValues(m.nodeLabel).Set(float64(latestBlock.BlockHeader.RawData.Timestamp))

	// First run: only check the current latest block and initialize lastBlockNum
	if m.lastBlockNum == 0 {
		log.Printf("Initializing empty block monitor, checking current latest block: %d", latestBlockNum)
		m.checkSingleBlock(latestBlock, latestBlockNum)
		m.lastBlockNum = latestBlockNum
		NodeMetrics.LastCheckTime.WithLabelValues(m.nodeLabel, "empty_block").Set(float64(time.Now().Unix()))
		return
	}

	// If no new blocks since last check, skip
	if latestBlockNum <= m.lastBlockNum {
		return
	}

	// Check all blocks from lastBlockNum+1 to latestBlockNum to avoid missing any
	// This handles the case where multiple blocks are produced between checks
	for blockNum := m.lastBlockNum + 1; blockNum <= latestBlockNum; blockNum++ {
		var block *Block
		var err error

		// Use the latest block we already fetched if it's the one we're checking
		if blockNum == latestBlockNum {
			block = latestBlock
		} else {
			// Fetch the specific block by number
			block, err = m.client.GetBlockByNum(blockNum)
			if err != nil {
				log.Printf("Failed to get block %d: %v", blockNum, err)
				NodeMetrics.APICallErrors.WithLabelValues(m.nodeLabel, "getblockbynum", "request_failed").Inc()
				continue // Skip this block but continue checking others
			}
		}

		m.checkSingleBlock(block, blockNum)
	}

	// Update lastBlockNum to the latest block we've checked
	m.lastBlockNum = latestBlockNum
	NodeMetrics.LastCheckTime.WithLabelValues(m.nodeLabel, "empty_block").Set(float64(time.Now().Unix()))
}

// checkSingleBlock checks a single block for emptiness and updates metrics
func (m *EmptyBlockMonitor) checkSingleBlock(block *Block, blockNum int64) {
	witnessAddress := block.BlockHeader.RawData.WitnessAddress

	// Check if block is empty
	isEmpty := m.isBlockEmpty(block)

	if isEmpty {
		log.Printf("Empty block detected: block_num=%d, witness=%s", blockNum, witnessAddress)
		EmptyBlockMetrics.EmptyBlockCount.WithLabelValues(m.nodeLabel, witnessAddress).Inc()
		EmptyBlockMetrics.EmptyBlockDetected.WithLabelValues(
			m.nodeLabel,
			fmt.Sprintf("%d", blockNum),
			witnessAddress,
		).Set(1)

		// expose block hash for empty blocks
		blockHash := computeBlockHash(block)
		EmptyBlockMetrics.EmptyBlockInfo.WithLabelValues(
			m.nodeLabel,
			fmt.Sprintf("%d", blockNum),
			witnessAddress,
			blockHash,
		).Set(1)
	} else {
		// Reset the detection flag for non-empty blocks
		EmptyBlockMetrics.EmptyBlockDetected.WithLabelValues(
			m.nodeLabel,
			fmt.Sprintf("%d", blockNum),
			witnessAddress,
		).Set(0)
	}
}

// isBlockEmpty checks if a block is empty
// A block is considered empty if it has no transactions
func (m *EmptyBlockMonitor) isBlockEmpty(block *Block) bool {
	// A block is empty if it contains no transactions
	// Note: In Tron, even empty blocks might have witness transactions,
	// but for monitoring purposes, we consider a block empty if transactions array is empty
	return len(block.Transactions) == 0
}

// computeBlockHash returns the hash identifier of the block
// For now we simply use the BlockID field from the API response
func computeBlockHash(block *Block) string {
	return block.BlockID
}
