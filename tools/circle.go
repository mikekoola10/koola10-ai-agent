package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// CircleClient wraps the Circle Agent Stack API
type CircleClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// WalletResponse for create/get wallet
type WalletResponse struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Balance string `json:"balance"`
	Name    string `json:"name"`
}

// TransferRequest for sending USDC
type TransferRequest struct {
	FromWallet string `json:"from_wallet"`
	ToAddress  string `json:"to_address"`
	Amount     string `json:"amount"`   // e.g., "10.00"
	Currency   string `json:"currency"` // "USDC"
}

// TransferResponse
type TransferResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// NewCircleClient initializes a client using CIRCLE_API_KEY from env
func NewCircleClient() *CircleClient {
	return &CircleClient{
		APIKey:  os.Getenv("CIRCLE_API_KEY"),
		BaseURL: "https://api.circle.com/v1",
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// doRequest handles auth and JSON parsing
func (c *CircleClient) doRequest(method, path string, body interface{}, result interface{}) error {
	if c.APIKey == "" {
		// Mock handling for tests or missing key
		return nil
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("circle API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// CreateWallet creates a new programmable wallet
func (c *CircleClient) CreateWallet(name string) (*WalletResponse, error) {
	if c.APIKey == "" {
		return &WalletResponse{ID: "mock_id", Address: "mock_address", Name: name}, nil
	}
	var resp struct {
		Data WalletResponse `json:"data"`
	}
	err := c.doRequest("POST", "/wallets", map[string]string{"name": name}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetWallet retrieves wallet details by ID
func (c *CircleClient) GetWallet(id string) (*WalletResponse, error) {
	if c.APIKey == "" {
		return &WalletResponse{ID: id, Balance: "0.00"}, nil
	}
	var resp struct {
		Data WalletResponse `json:"data"`
	}
	err := c.doRequest("GET", "/wallets/"+id, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// SendUSDC sends USDC from a wallet to an external address
func (c *CircleClient) SendUSDC(fromWallet, toAddress, amount string) (*TransferResponse, error) {
	if c.APIKey == "" {
		return &TransferResponse{ID: "mock_tx_id", Status: "pending"}, nil
	}
	req := TransferRequest{
		FromWallet: fromWallet,
		ToAddress:  toAddress,
		Amount:     amount,
		Currency:   "USDC",
	}
	var resp struct {
		Data TransferResponse `json:"data"`
	}
	err := c.doRequest("POST", "/transfers", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetBalance returns the balance of a wallet in USDC
func (c *CircleClient) GetBalance(walletID string) (string, error) {
	wallet, err := c.GetWallet(walletID)
	if err != nil {
		return "", err
	}
	return wallet.Balance, nil
}
