# Liminal Vibe Banking Hackathon - Staff Guide

**Audience:** Hackathon organizers, mentors, judges, support staff

This guide helps you support students building AI financial tools with nim-go-sdk and Liminal banking APIs.

---

## Event Overview

### What Students Are Building

Students create **conversational AI financial agents** powered by:
- **Claude AI** (via Anthropic API) - Natural language understanding
- **nim-go-sdk** (Go backend) - WebSocket server with tool execution
- **nim-chat** (React widget) - Beautiful chat interface
- **Liminal Banking APIs** - Real stablecoin banking operations

**Example projects:**
- Savings goal tracker that analyzes spending patterns
- Budget coach that alerts when approaching limits
- Smart savings advisor that recommends optimal deposits
- Financial health score calculator with personalized tips

### Tech Stack

```
┌─────────────────────────────────────────┐
│         Student's Browser               │
│  ┌───────────────────────────────────┐  │
│  │   React App + nim-chat widget     │  │
│  └───────────────┬───────────────────┘  │
└──────────────────┼──────────────────────┘
                   │ WebSocket
┌──────────────────┼──────────────────────┐
│  Student's Laptop│                      │
│  ┌───────────────┴───────────────────┐  │
│  │   nim-go-sdk Backend (Go)         │  │
│  │   - Claude AI integration         │  │
│  │   - 9 Liminal banking tools       │  │
│  │   - Custom analytical tools       │  │
│  └───────────────┬───────────────────┘  │
└──────────────────┼──────────────────────┘
                   │ HTTP
┌──────────────────┼──────────────────────┐
│     Liminal APIs │                      │
│  ┌───────────────┴───────────────────┐  │
│  │   Banking Operations              │  │
│  │   - Balance checks                │  │
│  │   - Transaction history           │  │
│  │   - Money transfers               │  │
│  │   - Savings management            │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

### Timeline (6-Hour Sprint)

- **Hour 0:** Opening ceremony + tech intro (30 min)
- **Hour 1:** Setup, API keys, get "Hello World" working
- **Hour 2-4:** Build core custom tool + functionality
- **Hour 5:** Polish UI/UX and conversational experience
- **Hour 6:** Prep demo and documentation
- **Hour 6.5:** Demos + judging
- **Hour 7:** Awards + closing

---

## Before the Event

### ✅ Pre-Event Checklist

**1 Week Before:**
- [ ] Liminal API environment ready (dev or prod)
- [ ] Email/OTP authentication system tested
- [ ] Test end-to-end flow: Email → OTP → JWT → nim-go-sdk → tools
- [ ] Discord server setup with #hackathon-help channel
- [ ] All staff have access to monitoring dashboards
- [ ] Verify email delivery is working

**1 Day Before:**
- [ ] Test email/OTP login flow on fresh device
- [ ] Verify all 9 Liminal tools working with JWT auth
- [ ] Run through hackathon-starter quickstart
- [ ] Prepare demo for opening ceremony
- [ ] Charge laptops, test projector

**Morning Of:**
- [ ] Verify Liminal APIs are healthy
- [ ] Test email/OTP login flow
- [ ] Post setup instructions in Discord
- [ ] Set up help desk area
- [ ] Test WiFi with multiple connections

---

## Opening Ceremony (30 Minutes)

### Suggested Agenda

**Welcome (5 min)**
- Thanks for coming!
- Today you're building AI financial tools
- Show example: "What's my balance?" demo
- Prizes: 1st ($X), 2nd ($Y), 3rd ($Z)

**Tech Overview (10 min)**
- Show architecture diagram
- Explain: Claude AI + nim-go-sdk + Liminal APIs
- Key concept: "Tools are superpowers for AI"
- You'll use 9 banking tools + build custom ones

**Getting Started (10 min)**
- Live walkthrough: Clone → Add Anthropic key → Run
- Show where to get Anthropic API key (console.anthropic.com)
- Explain login flow: email → OTP → automatic JWT auth
- Quick tour of hackathon-starter code

**Project Ideas (5 min)**
- Spending analyzer, budget tracker, savings coach
- Focus on solving real financial pain points
- Best projects have unique insights, not just data display

**Q&A + Kickoff**
- Support channels: Discord #hackathon-help
- Mentors will circulate
- Let's build! 🚀

### Live Demo Script

```bash
# Show this during opening:

# 1. Clone the repo
git clone https://github.com/becomeliminal/nim-go-sdk.git
cd nim-go-sdk/examples/hackathon-starter

# 2. Add Anthropic API key (already set up on your machine)
# Show .env file with ANTHROPIC_API_KEY

# 3. Start backend
go run main.go
# Point out: "Added 9 Liminal banking tools"

# 4. Start frontend (different terminal)
cd frontend && npm run dev
# Browser opens automatically

# 5. Demo login and conversation
# Click chat bubble → show login screen
# Enter email → show OTP code entry
# Enter code → authenticated!
# Type in chat: "What's my balance?"
# Show AI using get_balance tool
# Type: "Show my recent transactions"
# Show streaming response with transaction data
# Type: "Send $10 to @alice"
# Show confirmation flow with countdown

# 6. Show the code
# Open main.go, point out:
# - Liminal tools setup (line ~100)
# - Custom tool example (line ~110)
# - Easy to add more!

# "That's it! Now go build something awesome!"
```

---

## Supporting Students

### Common Questions & Answers

**Q: "Where do I get my Anthropic API key?"**

**A:** Go to console.anthropic.com, create an account, and generate an API key. Add it to your `.env` file.

**Q: "How do I authenticate with Liminal?"**

**A:** It's automatic! When you first open the chat, you'll see a login screen. Enter your email, get a one-time code, and you're in. No API key needed - JWT authentication is handled automatically.

**Q: "My backend won't start - ANTHROPIC_API_KEY required"**

**A:** You need to create a `.env` file. Copy `.env.example` to `.env` and add your Anthropic key:
```bash
cp .env.example .env
# Then edit .env and add: ANTHROPIC_API_KEY=sk-ant-your-key-here
```

**Q: "The Liminal tools aren't working - Unauthorized"**

**A:** This means you haven't logged in yet:
- Click the chat bubble
- Enter your email address
- Check your email for the OTP code
- Enter the code to authenticate
- The backend will automatically use your JWT token

**Q: "How do I add a custom tool?"**

**A:** Look at the spending analyzer example in `main.go` around line 210. Copy that pattern:
```go
// 1. Create the tool
srv.AddTool(createMyCustomTool(liminalExecutor))

// 2. Implement it
func createMyCustomTool(liminalExecutor core.ToolExecutor) core.Tool {
    return tools.New("my_tool_name").
        Description("What this tool does").
        Schema(tools.ObjectSchema(map[string]interface{}{
            "param": tools.StringProperty("Description"),
        })).
        HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
            // Your logic here
            return result, nil
        }).
        Build()
}
```

**Q: "Can I use Python instead of Go?"**

**A:** The nim-go-sdk is Go-only. If you're not comfortable with Go:
- Go is very readable - you can modify the examples even as a beginner
- We have mentors who can help with Go syntax
- The examples have lots of comments to guide you

**Q: "What should I build?"**

**A:** Think about financial pain points you've experienced:
- "I never know if I'm on track with my budget"
- "I don't save enough because I forget"
- "I wish I understood where my money goes"
- "I want to build an emergency fund but don't know how"

Pick one problem and build an AI tool that solves it.

**Q: "How do I make my AI sound better?"**

**A:** Edit the `hackathonSystemPrompt` in `main.go`. This defines your AI's personality:
- Make it friendly, sassy, professional, encouraging - your choice!
- Give it expertise: "You're a savings expert..."
- Set the tone: formal vs. casual, brief vs. detailed

**Q: "My frontend won't connect to the backend"**

**A:**
1. Make sure backend is running (`go run main.go`)
2. Check it says "WebSocket endpoint: ws://localhost:8080/ws"
3. Make sure frontend is pointing to the right URL (should be automatic)
4. Try refreshing the browser

---

## Troubleshooting Guide

### Quick Diagnostics

When a student has an issue, ask:

1. **"What error message do you see?"**
   - Read it carefully - it usually tells you what's wrong

2. **"Did the backend start successfully?"**
   - Look for: "🚀 Hackathon Starter Server Running"
   - If not, there's an error in the logs above

3. **"Did you add both API keys to .env?"**
   - Check: `cat .env` shows both keys
   - No quotes, no spaces

4. **"Can you curl the health endpoint?"**
   ```bash
   curl http://localhost:8080/health
   # Should return: {"status":"ok"}
   ```

### Issue: Port Already in Use

**Symptom:** `bind: address already in use`

**Fix:**
```bash
# Find what's using port 8080
lsof -i :8080

# Kill it
kill -9 <PID>

# Or change port in .env
echo "PORT=8081" >> .env
```

### Issue: Go Dependencies Not Found

**Symptom:** `package not found` or `cannot find module`

**Fix:**
```bash
cd /path/to/hackathon-starter
go mod tidy
go run main.go
```

### Issue: npm Install Fails

**Symptom:** Errors during `npm install` in frontend

**Fix:**
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Issue: Confirmation Flow Not Working

**Symptom:** Money transfer doesn't show countdown/confirmation

**Fix:**
- This is expected! Write operations (send_money, deposit_savings, withdraw_savings) require user confirmation
- The UI should show a card with summary and countdown
- Student clicks "Confirm" to approve
- If UI doesn't show, check browser console for errors

---

## Judging Criteria

### What Makes a Winning Project?

**Impact (35%)**
- Solves a real financial problem people face
- Would genuinely improve someone's financial life
- Clear value proposition

**Technical Implementation (25%)**
- Uses Liminal tools effectively
- Custom tool is well-designed
- Code quality and error handling

**AI/UX Experience (25%)**
- Conversational and natural
- AI provides insights, not just data
- Confirmation flows work smoothly
- Delightful to interact with

**Innovation (15%)**
- Unique approach or insight
- Creative use of available tools
- Goes beyond obvious solutions

### Red Flags (Avoid These)

- ❌ Just shows raw data without analysis
- ❌ Generic chatbot that doesn't use banking tools
- ❌ Doesn't actually work / demo fails
- ❌ Too complex - tried to do everything
- ❌ No clear problem being solved

### Green Flags (Look For These)

- ✅ Clear problem statement in demo
- ✅ AI provides actionable insights
- ✅ Custom tool adds real value
- ✅ Polished conversational experience
- ✅ Handles errors gracefully
- ✅ Unique perspective on financial health

---

## Demo Time

### Demo Format (5 min per team)

**Suggested structure:**
1. **Problem** (30 sec) - "Have you ever struggled to save money?"
2. **Solution** (30 sec) - "We built an AI savings coach..."
3. **Live Demo** (3 min) - Show actual conversation
4. **Impact** (30 sec) - "This helps people build emergency funds"
5. **Q&A** (30 sec) - Judges ask questions

### Demo Tips to Share

- **Practice once before presenting**
- **Have a backup recording** in case live demo fails
- **Start with a relatable problem**
- **Show, don't tell** - live interaction is better than slides
- **Highlight the AI** - show how it understands natural language
- **Explain the custom tool** - what unique analysis does it provide?
- **Be enthusiastic!** - energy matters

### Tech Support During Demos

- Have a laptop ready at help desk for emergency debugging
- If a demo fails, stay calm and help them switch to backup
- Make sure projector/screen sharing works before demos start

---

## After the Event

### Immediate (Same Day)

- [ ] Collect feedback forms from students
- [ ] Take photos of winning projects
- [ ] Note any technical issues that occurred
- [ ] Thank students and mentors

### Follow-Up (Within 1 Week)

- [ ] Send winner announcements
- [ ] Share project showcase on social media
- [ ] Process prize payments
- [ ] Review what went well and what didn't

### Clean-Up (Within 1 Week)

- [ ] Revoke hackathon API keys (see Internal Guide)
- [ ] Archive interesting project code
- [ ] Update docs based on lessons learned

### Retrospective

**Discuss with team:**
- What friction did students hit?
- Which parts of docs were confusing?
- What questions came up repeatedly?
- How can we improve next time?

---

## Resources for Mentors

### Quick Links

- **Hackathon Starter README:** `/examples/hackathon-starter/README.md`
- **Internal API Key Guide:** `/labs/docs/hackathon-api-keys.md`
- **nim-go-sdk Docs:** Root README.md
- **Discord:** #hackathon-help channel

### Mentor Tips

**Do:**
- Ask questions to understand their vision first
- Point them to relevant examples in the code
- Help them scope down to something achievable
- Debug together, don't just fix it for them
- Celebrate small wins ("Nice! Balance check works!")

**Don't:**
- Write code for them (guide, don't do)
- Overcomplicate with "enterprise" patterns
- Suggest adding too many features
- Take over their keyboard
- Make them feel bad for asking questions

### Common Code Patterns to Share

**Calling Liminal tools from custom tool:**
```go
txRequest := map[string]interface{}{"limit": 100}
txJSON, _ := json.Marshal(txRequest)
response, err := liminalExecutor.Execute(ctx, "get_transactions", txJSON)
```

**Parsing tool input:**
```go
var params struct {
    Amount string `json:"amount"`
}
json.Unmarshal(input, &params)
```

**Returning structured data:**
```go
return map[string]interface{}{
    "insight": "You spend $50/day on average",
    "trend": "increasing",
    "recommendation": "Consider setting a daily budget",
}, nil
```

---

## Emergency Contacts

- **Platform Issues:** @platform-team in Discord
- **API/Auth Down:** Escalate to backend on-call
- **Email Delivery Issues:** Check email service dashboard
- **Event Coordinator:** [your phone]

---

## Backup Plans

### If Email Delivery is Failing

1. Check email service status immediately
2. Have backup authentication method ready (manual account creation if needed)
3. Announce in Discord with status and ETA
4. Consider extending hackathon time if it affects many teams

### If Liminal APIs are Down

1. Check status dashboard
2. Estimate downtime
3. If >30 min, consider:
   - Extend hackathon time
   - Allow mock data mode
   - Focus judging on concept vs. working demo

### If Claude API is Having Issues

1. Check Anthropic status page
2. Students can switch models in main.go if needed:
   ```go
   Model: "claude-3-5-sonnet-20241022", // Older but stable
   ```

### If WiFi Fails

1. Have mobile hotspot ready
2. Students can tether to phone
3. Coordinate with venue IT immediately

---

## Success Metrics

**Great hackathon if:**
- [ ] >90% of teams get "Hello World" working in Hour 1
- [ ] >80% of teams successfully create custom tool
- [ ] <5 escalations to platform team
- [ ] Students have fun and learn something
- [ ] Winners demo projects that "wow" the judges

**Red flags during event:**
- Multiple teams stuck on same issue → doc/setup problem
- Lots of questions about Go syntax → need more examples
- Confusion about what to build → need clearer project ideas
- API errors → check monitoring dashboard

---

## Contact

Questions about this guide? Ping @hackathon-team in Discord.

Good luck, and thank you for supporting the next generation of fintech builders! 🚀

---

**Last Updated:** 2024-01-30
**For:** Liminal Vibe Banking Hackathon
