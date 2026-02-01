# Technical System Documentation

## Overview
This document explains the technical details of the **Anomaly Detection** and **Blockchain** features in the Nim Financial Agent. It is written to be easy to understand while providing deep insight into the specific functions and methods used in the code.

---

## 1. Anomaly Detection Subsystem

The Anomaly Detection system is a custom tool (`detect_lil_anomalies`) that simulates a banking environment to find suspicious transactions.

### 1.1 Data Structures

#### Enriched Transaction (`EnrichedTx`)
This is the main object we work with. It combines raw transaction data with real-world currency values.
*   **Code Location**: `main.go`
*   **Structure**: 
    ```go
    type EnrichedTx struct {
        ID         string  // e.g., "tx_1"
        AmountLIL  float64 // Original crypto amount
        AmountFiat float64 // Converted USD amount (from CoinGecko)
        Merchant   string  // e.g., "Coffee Shop"
        // ... timestamp, category, etc.
    }
    ```

#### Anomaly Report (`Anomaly`)
When we find something wrong, we create this object.
*   **Structure**:
    *   `RiskScore`: A number from 0.0 to 1.0 (1.0 is very dangerous).
    *   `ReasonCodes`: Short tags like "large_spend" or "rapid_burst".
    *   `Explanations`: Human-readable text explaining *why* it was flagged.

### 1.2 Key Functions & Methods

#### A. `generateDummyLILTransactions(n int)`
**What it does**: Creates fake history so we have something to analyze.
**Methodology**:
1.  **Normal Loop**: It loops `N-2` times to create standard, boring transactions (small amounts, random merchants).
2.  **The Injection**: It manually appends **2 specific outliers** at the end of the list with huge amounts (15-20 LIL) to test the "Big Spender" rule.
3.  **The Burst**: It modifies a block of 8 transactions to happen 5 minutes apart with tiny amounts, testing the "Machine Gun" rule.

#### B. `fetchLILRateStub(fiat string)`
**What it does**: Gets the real-world price of the token.
**Methodology**:
*   **API Call**: It sends an HTTP GET request to `api.coingecko.com`.
*   **Caching**: To stop us from get banned by CoinGecko for too many requests, it saves the price in a `rateCache` variable.
*   **Logic**: Before calling the API, it checks: `if time.Since(LastUpdated) < 5 minutes`. If true, it uses the saved price instead of calling the internet again.

#### C. `detectAnomaliesOnDummy(...)`
**What it does**: The "Brain" that runs the rules.
**Methodology**:
1.  **Z-Score Logic**: It calculates the *median* (average-ish) spend. If any single transaction is **5x larger** than that median, it marks it as an anomaly.
    *   *Why Median?* Averages can be tricked by one huge number. Medians are safer.
2.  **Sliding Window Logic**: It looks at time. It asks: "Did we see more than 8 small transactions in the last 60 minutes?" 
    *   If yes -> It assumes a bot is testing a stolen credit card.

---

## 2. Blockchain Ledger Subsystem

The blockchain is a digital notebook that is mathematically impossible to edit without breaking the chain.

### 2.1 Data Structures (`blockchain/block.go`)

#### The Block
Each "page" in our notebook is a `Block` struct:
```go
type Block struct {
    Index      int    // Page number (0, 1, 2...)
    Timestamp  string // When it was created
    Data       string // The actual report (JSON text)
    PrevHash   string // The fingerprint of the PREVIOUS block
    Hash       string // The fingerprint of THIS block
    Nonce      int    // The "magic number" used for mining
    Difficulty int    // How many zeros we need (default: 2)
}
```

### 2.2 Key Methods & Algorithms

#### A. `CalculateHash()`
**What it does**: Creates a unique digital fingerprint for the block.
**Methodology**:
1.  It takes all the fields (`Index` + `Timestamp` + `Data` + `PrevHash` + `Nonce`) and glues them together into one long string using `fmt.Sprintf`.
2.  It feeds that string into **SHA-256**, a standard cryptographic algorithm.
3.  **Result**: If you change *one single letter* in the data, this fingerprint changes completely. This is how we detect tampering.

#### B. `MineBlock(difficulty int)`
**What it does**: The "Proof of Work". It forces the computer to do heavy calculation before saving.
**Methodology**:
1.  **The Goal**: Find a hash that starts with specific number of zeros (e.g., "00...").
2.  **The Loop**:
    ```go
    for {
        calculate hash
        if hash starts with "00" {
            STOP! We found it.
        }
        Nonce++ // Try a new number
    }
    ```
3.  **Why?** This makes it expensive to rewrite history. If a hacker wants to change Block 1, they have to re-do the work for Block 1, Block 2, Block 3... which takes too much time.

#### C. `IsChainValid()` (`blockchain/chain.go`)
**What it does**: Audits the entire book to ensure no pages were torn out.
**Methodology**:
*   It loops through every block from start to finish.
*   **Check 1**: Does `Block[i].PrevHash` match `Block[i-1].Hash`? (Are the pages glued together correctly?)
*   **Check 2**: Does re-calculating the hash produce the same result? (Did someone scribble on the page?)
*   If either fails, the chain is "Invalid".

#### D. `SaveToFile()` & `LoadFromFile()`
**What it does**: Ensures data survives a restart.
**Methodology**:
*   Uses Go's `encoding/json` library.
*   **Save**: Dumps the entire `Blockchain` struct into `blockchain.json` immediately after every new block.
*   **Load**: Reads that file when the server starts up. If the file exists, it restores the chain state. If not, it starts a fresh one.
