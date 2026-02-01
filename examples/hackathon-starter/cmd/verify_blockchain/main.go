package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/becomeliminal/nim-go-sdk/examples/hackathon-starter/blockchain"
)

func main() {
	// 1. Initialize Chain
	// Note: We are using a local instance here for testing, not the actual global one from main.go
	// This tests the package logic, not the main integration directly.
	chain := blockchain.NewBlockchain()
	fmt.Printf("✅ Blockchain initialized. Genesis: %s\n", chain.Chain[0].Hash)

	// 2. Mock adding mock data
	generateAndRecordBatch(chain, 1, "USD")
	generateAndRecordBatch(chain, 2, "EUR")

	// 3. Verify Chain
	if !chain.IsChainValid() {
		log.Fatal("❌ Chain invalid!")
	}
	fmt.Println("✅ Chain validity check passed.")

	// 4. Print Blocks
	for i, b := range chain.Chain {
		fmt.Printf("🧱 Block %d: [Prev: %.10s...] [Hash: %.10s...] DataLen: %d\n",
			i, b.PrevHash, b.Hash, len(b.Data))
	}
}

func generateAndRecordBatch(chain *blockchain.Blockchain, batchID int, currency string) {
	// Simulate data
	data := map[string]interface{}{
		"batch_id":  batchID,
		"currency":  currency,
		"tx_count":  5,
		"timestamp": time.Now().String(),
	}
	bytes, _ := json.Marshal(data)

	// Add to chain
	newBlock := chain.AddBlock(string(bytes))
	fmt.Printf("⛏️  Mined Block %d in %dms\n", newBlock.Index, newBlock.ProofTime_ms)
}
