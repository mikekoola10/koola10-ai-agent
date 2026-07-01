package sterling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"koola10/tools"
	"log"
	"net/http"
	"os"
)

type PayoutRequest struct {
	Amount     float64 `json:"amount"`
	TargetTag  string  `json:"target_tag"`
	CardNumber string  `json:"card_number"`
	CardExpiry string  `json:"card_expiry"` // MM/YY
	CardCVV    string  `json:"card_cvv"`
}

type PayoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	TxID    string `json:"tx_id,omitempty"`
}

type PayoutSwarm struct {
	agentCardClient *AgentCardClient
	browserAgentURL string
	ledger          Ledger
	vault           *VaultClient
}

func NewPayoutSwarm(ledger Ledger, vault *VaultClient) *PayoutSwarm {
	return &PayoutSwarm{
		agentCardClient: NewAgentCardClient(),
		browserAgentURL: os.Getenv("BROWSER_AGENT_URL"),
		ledger:          ledger,
		vault:           vault,
	}
}

// SendPayout creates a virtual card and instructs the browser agent to send money to Cash App tag.
func (ps *PayoutSwarm) SendPayout(amount float64, targetTag string) error {
	// 1. Create a virtual card with a spend limit equal to the payout amount (in cents)
	spendLimitCents := int(amount * 100)
	card, err := ps.agentCardClient.CreateVirtualCard(
		fmt.Sprintf("Payout to %s", targetTag),
		spendLimitCents,
		true, // auto-destruct after use (one-time card)
	)
	if err != nil {
		return fmt.Errorf("failed to create payout card: %w", err)
	}

	// 2. Prepare request for browser agent
	reqBody := PayoutRequest{
		Amount:     amount,
		TargetTag:  targetTag,
		CardNumber: card.PAN,
		CardExpiry: fmt.Sprintf("%02d/%02d", card.ExpMonth, card.ExpYear%100),
		CardCVV:    card.CVV,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// 3. Call browser agent
	resp, err := http.Post(ps.browserAgentURL+"/cashapp/payout", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("browser agent call failed: %w", err)
	}
	defer resp.Body.Close()

	var payoutResp PayoutResponse
	if err := json.NewDecoder(resp.Body).Decode(&payoutResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !payoutResp.Success {
		return fmt.Errorf("payout failed: %s", payoutResp.Message)
	}

	// 4. Record the expense in the economic ledger (from operations fund)
	txID, err := ps.ledger.RecordTransaction(
		fmt.Sprintf("Payout to %s", targetTag),
		-amount,
		"expense",
		fmt.Sprintf("CashApp payout via AgentCard, txID: %s", payoutResp.TxID),
	)
	if err != nil {
		log.Printf("Warning: ledger record failed: %v", err)
	}

	// 5. Log to vault (Google Sheets)
	err = ps.vault.AddEntry(VaultEntry{
		Description: fmt.Sprintf("Payout to %s", targetTag),
		Amount:      amount,
		Type:        "expense",
		Notes:       fmt.Sprintf("CashApp payout, card ID: %s", card.ID),
	})
	if err != nil {
		log.Printf("Warning: vault log failed: %v", err)
	}

	// 6. Send confirmation email via AgentMail
	tools.RunTool("agentmail", map[string]interface{}{
		"to":      os.Getenv("ADMIN_EMAIL"),
		"subject": fmt.Sprintf("Payout of $%.2f to %s Confirmed", amount, targetTag),
		"body":    fmt.Sprintf("The payout of $%.2f to %s has been successfully processed.\nTransaction ID: %s\nCard ID: %s", amount, targetTag, payoutResp.TxID, card.ID),
	})

	log.Printf("[PayoutSwarm] Successfully sent $%.2f to %s. Ledger TX: %s", amount, targetTag, txID)
	return nil
}
