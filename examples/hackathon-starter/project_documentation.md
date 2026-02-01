# Project Feature Documentation
This document provides a detailed, step-by-step explanation of the security and compliance features implemented in our Financial AI Agent ("Nim"). It is designed to be understood by anyone, regardless of technical background.

---

## 🏗️ 1. Anomaly Detection System (`detect_lil_anomalies`)

### What is it?
The Anomaly Detection System is like a security guard that watches over your transactions. It looks for "weird" or suspicious patterns that might indicate fraud or theft. If it finds something wrong, it immediately alerts you.

### How it Works (Step-by-Step)

#### Step 1: generating Transactions (The Simulation)
Since we are in a hackathon environment, we don't have decades of real banking history to analyze. So, the tool **simulates** (makes up) a realistic batch of transactions for you.
- It creates **~30 transactions** to represent your recent spending.
- Most are normal: Buying coffee, groceries, subscriptions (Netflix/Spotify).
- **The "Setup"**: The tool intentionally hides a few "bad" transactions in the mix to test if it can catch them.
    - **Example**: A sudden $2,000 purchase at a luxury store.
    - **Example**: 8 tiny payments of $0.20 in 5 minutes (hackers testing stolen cards).

#### Step 2: Currency Conversion (The Compliance Part)
The system works with a crypto token called **LIL**. However, real-world finance uses Dollars (USD) or Pounds (GBP).
- The tool connects to the internet (CoinGecko) to get the **LIVE price** of the token.
- It converts every single transaction into your local currency (e.g., USD) so it knows exactly how much money is moving in "real world" value.

#### Step 3: The Detection Engine (The Brain)
The AI analyzes every single transaction using specific rules to catch the bad guys:
1.  **The "Big Spender" Rule**: It calculates your "average" spending. If a transaction is **5x bigger** than your normal average (e.g., normally you spend $50, but suddenly spend $500), it flags it as High Risk.
2.  **The "Machine Gun" Rule (Burst Detection)**: If it sees many small transactions happening very fast (e.g., 8 transactions in under an hour), it knows a human can't shop that fast. It flags this as a "Card Testing Attack".

#### Step 4: Alerting (The Email)
Once the analysis is done, if any "High Risk" transactions are found, the system immediately sends an email report to the security team (or you).
- **Recipients**: It attempts to email `jainamvaria1010@gmail.com` and `pks850pks8311@gmail.com`.
- **Content**: The email lists exactly which transactions were flagged, what the Risk Score is (0 to 1), and *why* it was flagged (e.g., "This payment is 10x larger than typical").

---

## 🔗 2. Blockchain Ledger (The Permanent Record)

### What is it?
A blockchain is a digital notebook that **cannot be erased**. Once you write something in it, it is there forever. We use this to record every single security check we perform. This proves to auditors that we are actually doing our job and not hiding anything.

### How it Works (in Plain English)

#### 1. Blocks and Chains
Imagine a notebook where each page is numbered.
- **Page 1 (Genesis Block)**: The first page. It just says "Start".
- **Page 2**: Contains records of the first anomaly scan.
- **Page 3**: Contains records of the second scan.
- **The Chain**: On the top of Page 3, we write a special code that summarizes *everything* on Page 2. If someone tears out Page 2 or changes a number on it, the code on Page 3 won't match anymore, and we know someone tampered with the book. This is what makes it "secure".

#### 2. "Mining" (Proof of Work)
Before we can add a page to the book, the computer has to solve a difficult math puzzle. This takes a split second.
- **Why?** In the real world, this prevents hackers from spamming the book with millions of fake pages instantly. It proves "work" was done to record the data.
- **In our app**: You might see a log saying `Mined Block #2 in 5ms`. That's the computer solving the puzzle to lock the page in.

#### 3. Automatic Recording
You don't need to do anything manually.
- Every time the **Anomaly Detector** runs, the system automatically takes the full report (transactions + alerts) and seals it into a new Block.
- It then adds this Block to the Chain.

#### 4. Persistence (Saving to Disk)
- The system saves the entire notebook to a file on your computer called `blockchain.json`.
- **Why?** If you restart the computer, the notebook doesn't disappear. It loads the `blockchain.json` file when it wakes up, so your history is preserved forever.

### How to See it?
You can ask the AI: *"Show me the blockchain ledger"*.
It will show you:
- How many blocks (pages) are in the book.
- If the book is valid (hasn't been tampered with).
- The unique "Hash" codes for the recent pages.

---

## Summary of the Flow
1.  **Run Tool**: You ask the AI to "Check for suspicious activity".
2.  **Simulate**: The AI mocks up 30 transactions (some normal, some attacks).
3.  **Detect**: The Rules Engine catches the attacks.
4.  **Email**: The AI emails the full report to the admins.
5.  **Record**: The AI seals the entire report into a Blockchain Block and saves it to your hard drive.
