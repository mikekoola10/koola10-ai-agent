package tools

import (
	"fmt"
	"os"
	"testing"
)

func TestCircleCreateWallet(t *testing.T) {
	client := NewCircleClient()
	wallet, err := client.CreateWallet("Test Wallet")
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}
	if wallet == nil {
		t.Fatal("Wallet is nil")
	}
	fmt.Printf("Created wallet: ID=%s, Address=%s\n", wallet.ID, wallet.Address)

	if os.Getenv("CIRCLE_API_KEY") == "" {
		if wallet.ID != "mock_id" {
			t.Errorf("Expected mock_id, got %s", wallet.ID)
		}
	}
}

func TestCircleGetBalance(t *testing.T) {
	client := NewCircleClient()
	balance, err := client.GetBalance("test_wallet_id")
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}
	fmt.Printf("Balance: %s\n", balance)

	if os.Getenv("CIRCLE_API_KEY") == "" {
		if balance != "0.00" {
			t.Errorf("Expected 0.00, got %s", balance)
		}
	}
}
