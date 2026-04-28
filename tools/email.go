package tools

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
	"time"
)

type Transaction struct {
	Timestamp   string  `json:"timestamp"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Vertical    string  `json:"vertical,omitempty"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

type AuditEntry struct {
	Timestamp string                 `json:"timestamp"`
	Action    string                 `json:"action"`
	Details   map[string]interface{} `json:"details"`
	Hash      string                 `json:"hash"`
}

func emailTool(payload map[string]interface{}) ToolResult {
	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing action"}
	}

	switch action {
	case "send":
		return sendEmail(payload)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func sendEmail(payload map[string]interface{}) ToolResult {
	to, _ := payload["to"].(string)
	subject, _ := payload["subject"].(string)
	body, _ := payload["body"].(string)

	if to == "" || subject == "" || body == "" {
		return ToolResult{Success: false, Error: "Missing to, subject, or body"}
	}

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	var status string
	if host == "" {
		fmt.Printf("MOCK EMAIL SENT to %s: %s\n", to, subject)
		status = "Email sent successfully (mocked)"
	} else {
		auth := smtp.PlainAuth("", user, pass, host)
		msg := []byte("To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body + "\r\n")
		err := smtp.SendMail(host+":"+port, auth, user, []string{to}, msg)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("Failed to send email: %v", err)}
		}
		status = "Email sent successfully via SMTP"
	}

	// Logging to ledger and audit (data/ is excluded from git but used for verification)
	logToLedger("leadgen", "outreach", 0.05, fmt.Sprintf("Sent email to %s", to))
	logToAudit("email_sent", map[string]interface{}{"to": to, "subject": subject})

	return ToolResult{Success: true, Output: status}
}

func logToLedger(vertical, category string, amount float64, description string) {
	ledgerPath := "data/economic_ledger.json"
	os.MkdirAll("data", 0755)

	type EconomicLedger struct {
		Balance      float64       `json:"balance"`
		TotalCosts   float64       `json:"total_costs"`
		TotalRevenue float64       `json:"total_revenue"`
		Transactions []Transaction `json:"transactions"`
	}

	var ledger EconomicLedger
	data, err := os.ReadFile(ledgerPath)
	if err == nil {
		json.Unmarshal(data, &ledger)
	} else {
		ledger.Balance = 100.0
	}

	tx := Transaction{
		Timestamp:   time.Now().Format(time.RFC3339),
		Type:        "cost",
		Category:    category,
		Vertical:    vertical,
		Amount:      amount,
		Description: description,
	}
	ledger.Transactions = append(ledger.Transactions, tx)
	ledger.Balance -= amount
	ledger.TotalCosts += amount

	updated, _ := json.Marshal(ledger)
	os.WriteFile(ledgerPath, updated, 0644)
}

func logToAudit(action string, details map[string]interface{}) {
	auditPath := "data/audit_chain.jsonl"
	os.MkdirAll("data", 0755)

	lastHash := "0000000000000000000000000000000000000000000000000000000000000000"
	f, err := os.Open(auditPath)
	if err == nil {
		scanner := bufio.NewScanner(f)
		var lastLine string
		for scanner.Scan() {
			lastLine = scanner.Text()
		}
		f.Close()
		if lastLine != "" {
			var e AuditEntry
			if err := json.Unmarshal([]byte(lastLine), &e); err == nil {
				lastHash = e.Hash
			}
		}
	}

	entry := AuditEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Action:    action,
		Details:   details,
	}
	entryJSON, _ := json.Marshal(entry)
	h := sha256.New()
	h.Write([]byte(lastHash + string(entryJSON)))
	entry.Hash = hex.EncodeToString(h.Sum(nil))

	f, _ = os.OpenFile(auditPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	json.NewEncoder(f).Encode(entry)
	f.Close()
}

func init() {
	RegisterTool("email", emailTool)
}
