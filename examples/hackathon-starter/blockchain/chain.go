package blockchain

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Blockchain manages the chain of blocks
type Blockchain struct {
	Chain      []Block `json:"chain"`
	Difficulty int     `json:"difficulty"`
}

// NewBlockchain creates a new chain with a Genesis block
func NewBlockchain() *Blockchain {
	genesisBlock := Block{
		Index:      0,
		Timestamp:  time.Now().String(),
		Data:       "Genesis Block - Hackathon Starter",
		PrevHash:   "0",
		Difficulty: 2, // Low difficulty for demo speed
		Nonce:      0,
	}
	genesisBlock.MineBlock(2)

	return &Blockchain{
		Chain:      []Block{genesisBlock},
		Difficulty: 2,
	}
}

// AddBlock adds a new block with the given data to the chain
func (bc *Blockchain) AddBlock(data string) Block {
	prevBlock := bc.Chain[len(bc.Chain)-1]
	newBlock := Block{
		Index:      prevBlock.Index + 1,
		Timestamp:  time.Now().String(),
		Data:       data,
		PrevHash:   prevBlock.Hash,
		Difficulty: bc.Difficulty,
		Nonce:      0,
	}
	newBlock.MineBlock(bc.Difficulty)
	bc.Chain = append(bc.Chain, newBlock)
	return newBlock
}

// IsChainValid checks the integrity of the blockchain
func (bc *Blockchain) IsChainValid() bool {
	for i := 1; i < len(bc.Chain); i++ {
		currentBlock := bc.Chain[i]
		prevBlock := bc.Chain[i-1]

		if currentBlock.Hash != currentBlock.CalculateHash() {
			fmt.Printf("Block %d Invalid Hash\n", i)
			return false
		}

		if currentBlock.PrevHash != prevBlock.Hash {
			fmt.Printf("Block %d Invalid PrevHash\n", i)
			return false
		}
	}
	return true
}

// SaveToFile saves the blockchain to a JSON file
func (bc *Blockchain) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(bc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadFromFile loads the blockchain from a JSON file
func LoadFromFile(filename string) (*Blockchain, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var bc Blockchain
	if err := json.Unmarshal(data, &bc); err != nil {
		return nil, err
	}
	return &bc, nil
}
