package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Block represents a single block in the chain
type Block struct {
	Index        int    `json:"index"`
	Timestamp    string `json:"timestamp"`
	Data         string `json:"data"`
	PrevHash     string `json:"prev_hash"`
	Hash         string `json:"hash"`
	Nonce        int    `json:"nonce"`
	Difficulty   int    `json:"difficulty"`
	ProofTime_ms int64  `json:"proof_time_ms"`
}

// CalculateHash creates a SHA256 hash of the block content
func (b *Block) CalculateHash() string {
	record := fmt.Sprintf("%d%s%s%s%d", b.Index, b.Timestamp, b.Data, b.PrevHash, b.Nonce)
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// MineBlock performs a simple Proof of Work
func (b *Block) MineBlock(difficulty int) {
	startTime := time.Now()
	target := ""
	for i := 0; i < difficulty; i++ {
		target += "0"
	}

	for {
		b.Hash = b.CalculateHash()
		// Check if hash starts with difficulty-number of zeros
		if len(b.Hash) >= difficulty && b.Hash[:difficulty] == target {
			break
		}
		b.Nonce++
	}
	b.ProofTime_ms = time.Since(startTime).Milliseconds()
}
