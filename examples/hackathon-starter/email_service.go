package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// sendAnomalyReportEmail sends the details of detected anomalies to the specified recipients.
// For the hackathon purpose, this uses Gmail SMTP.
// NOTE: This requires 'SMTP_PASSWORD' (App Password) to be set in environment variables for actual sending.
// If not set, it will log the email content to console for verification.
// sendAnomalyReportEmail sends the details of detected anomalies to the specified recipients.
// Returns a status string to be displayed to the user.
func sendAnomalyReportEmail(recipients []string, anomalies []Anomaly, transactions []EnrichedTx) string {
	if len(anomalies) == 0 {
		return "No anomalies detected, so no email was sent."
	}

	// 1. Construct Email Body
	var body strings.Builder
	subject := fmt.Sprintf("Subject: 🚨 Suspicious Activity Alert (%d Anomalies Detected)\n", len(anomalies))

	// Headers
	body.WriteString(subject)
	body.WriteString("To: " + strings.Join(recipients, ", ") + "\n")
	body.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\n\n")

	// Message
	body.WriteString("Nim AI Monitor has detected suspicious activity in your recent transaction batch.\n\n")
	body.WriteString("===== SUMMARY =====\n")
	body.WriteString(fmt.Sprintf("Total Transactions Scanned: %d\n", len(transactions)))
	body.WriteString(fmt.Sprintf("Total Anomalies Found: %d\n", len(anomalies)))
	body.WriteString("===================\n\n")

	body.WriteString("===== DETAILS =====\n")
	for i, a := range anomalies {
		body.WriteString(fmt.Sprintf("[%d] TxID: %s\n", i+1, a.TransactionID))
		body.WriteString(fmt.Sprintf("    Risk Score: %.2f / 1.0\n", a.RiskScore))
		body.WriteString(fmt.Sprintf("    Flags: %s\n", strings.Join(a.ReasonCodes, ", ")))
		if len(a.Explanations) > 0 {
			body.WriteString(fmt.Sprintf("    Analysis: %s\n", strings.Join(a.Explanations, "; ")))
		}
		body.WriteString("\n")
	}
	body.WriteString("===================\n")
	body.WriteString("\nPlease review these transactions immediately via the Liminal Dashboard.\n")
	body.WriteString("- Your AI Financial Assistant\n")

	emailContent := body.String()

	// 2. Check for SMTP Configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Allow configuring sender via ENV, default to a placeholder
	senderEmail := os.Getenv("SENDER_EMAIL")

	senderPassword := os.Getenv("SMTP_PASSWORD")

	// 3. Send Actual Email
	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, recipients, []byte(emailContent))
	if err != nil {
		log.Printf("❌ Failed to send alert email: %v\n", err)
		return fmt.Sprintf("❌ Failed to send email: %v. Check credentials for %s.", err, senderEmail)
	} else {
		log.Printf("✅ Anomaly alert email sent successfully to %v\n", recipients)
		return fmt.Sprintf("✅ Alert email successfully sent to %d recipients (from %s).", len(recipients), senderEmail)
	}
}
