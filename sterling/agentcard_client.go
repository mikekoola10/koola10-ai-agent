package sterling

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type AgentCardClient struct {
	APIKey string
}

type CardResponse struct {
	ID         string `json:"id"`
	PAN        string `json:"pan"`
	CVV        string `json:"cvv"`
	ExpMonth   string `json:"exp_month"`
	ExpYear    string `json:"exp_year"`
	Memo       string `json:"memo"`
}

func NewAgentCardClient() *AgentCardClient {
	return &AgentCardClient{
		APIKey: os.Getenv("AGENTCARD_API_KEY"),
	}
}

type MCPResponse struct {
	Result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	} `json:"result"`
}

func (ac *AgentCardClient) callMCP(method string, arguments map[string]interface{}) (string, error) {
	url := "https://mcp.agentcard.sh/mcp"
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      method,
			"arguments": arguments,
		},
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+ac.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var strBody string
	reader := bufio.NewReader(resp.Body)

	// Read response. If it's SSE, we'll get "event: ...\ndata: ...\n\n"
	// If it's plain JSON, we'll get "{...}"
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && strBody != "" {
				break
			}
			if err == io.EOF && strBody == "" {
				return "", fmt.Errorf("empty response from MCP")
			}
			return "", err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			strBody = strings.TrimPrefix(line, "data: ")
			break // We got our data event
		}

		if strings.HasPrefix(line, "{") {
			// Probably plain JSON
			strBody = line
			// Read the rest of the line if it was split? ReadString('\n') already got the full line.
			// Just in case it's multi-line JSON, we could keep reading until EOF,
			// but usually these are single-line or we can just try to parse what we have.
			remaining, _ := io.ReadAll(reader)
			strBody += string(remaining)
			break
		}

		// If we see "event: message", we stay in the loop to get the next "data: " line
		if strings.HasPrefix(line, "event: ") {
			continue
		}
	}

	var mcpRes MCPResponse
	if err := json.Unmarshal([]byte(strBody), &mcpRes); err != nil {
		return "", fmt.Errorf("failed to parse MCP response: %v, body: %s", err, strBody)
	}

	if mcpRes.Result.IsError {
		errorMsg := "Unknown MCP error"
		if len(mcpRes.Result.Content) > 0 {
			errorMsg = mcpRes.Result.Content[0].Text
		}
		return "", fmt.Errorf("MCP error: %s", errorMsg)
	}

	if len(mcpRes.Result.Content) == 0 {
		return "", fmt.Errorf("empty MCP response content")
	}

	return mcpRes.Result.Content[0].Text, nil
}

func (ac *AgentCardClient) CreateVirtualCard(memo string, spendLimitCents int, autoDestruct bool) (*CardResponse, error) {
	if ac.APIKey == "" {
		return nil, fmt.Errorf("AGENTCARD_API_KEY not set")
	}

	if spendLimitCents < 100 {
		spendLimitCents = 100
	}

	resText, err := ac.callMCP("create_card", map[string]interface{}{
		"amount_cents": spendLimitCents,
		"description":  memo,
	})
	if err != nil {
		return nil, err
	}

	var temp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(resText), &temp); err == nil && temp.ID != "" {
		return ac.GetCardDetails(temp.ID)
	}

	lines := strings.Split(resText, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Card ID: ") {
			cardID := strings.TrimSpace(strings.TrimPrefix(line, "Card ID: "))
			if cardID != "" {
				return ac.GetCardDetails(cardID)
			}
		}
	}

	return nil, fmt.Errorf("failed to parse card ID from: %s", resText)
}

func (ac *AgentCardClient) GetCardDetails(cardID string) (*CardResponse, error) {
	if ac.APIKey == "" {
		return nil, fmt.Errorf("AGENTCARD_API_KEY not set")
	}

	resText, err := ac.callMCP("get_card_details", map[string]interface{}{
		"card_id": cardID,
	})
	if err != nil {
		return nil, err
	}

	var res struct {
		PAN      string `json:"pan"`
		CVV      string `json:"cvv"`
		Expiry   string `json:"expiry"`
		ExpMonth string `json:"exp_month"`
		ExpYear  string `json:"exp_year"`
	}
	if err := json.Unmarshal([]byte(resText), &res); err == nil && res.PAN != "" {
		month := res.ExpMonth
		year := res.ExpYear
		if month == "" || year == "" {
			parts := strings.Split(res.Expiry, "/")
			if len(parts) == 2 {
				month = parts[0]
				year = "20" + parts[1]
			}
		}
		return &CardResponse{
			ID:       cardID,
			PAN:      res.PAN,
			CVV:      res.CVV,
			ExpMonth: month,
			ExpYear:  year,
		}, nil
	}

	var card CardResponse
	card.ID = cardID
	lines := strings.Split(resText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PAN: ") {
			card.PAN = strings.TrimPrefix(line, "PAN: ")
		} else if strings.HasPrefix(line, "CVV: ") {
			card.CVV = strings.TrimPrefix(line, "CVV: ")
		} else if strings.HasPrefix(line, "Expiry: ") {
			exp := strings.TrimPrefix(line, "Expiry: ")
			parts := strings.Split(exp, "/")
			if len(parts) == 2 {
				card.ExpMonth = parts[0]
				card.ExpYear = "20" + parts[1]
			}
		}
	}

	if card.PAN != "" {
		return &card, nil
	}

	return nil, fmt.Errorf("failed to parse card details: %v, body: %s", err, resText)
}

func (ac *AgentCardClient) BlockCard(cardID string) error {
	if ac.APIKey == "" {
		return fmt.Errorf("AGENTCARD_API_KEY not set")
	}

	_, err := ac.callMCP("close_card", map[string]interface{}{
		"card_id": cardID,
	})
	return err
}
