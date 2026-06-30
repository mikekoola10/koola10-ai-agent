package tools

import (
	"fmt"
	"log"
	"os"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
)

func stellarTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)

	secret := os.Getenv("STELLAR_SECRET_KEY")
	public := os.Getenv("STELLAR_PUBLIC_KEY")

	client := horizonclient.DefaultTestNetClient

	switch action {
	case "balance":
		if public == "" {
			return ToolResult{Success: false, Error: "STELLAR_PUBLIC_KEY not set"}
		}
		request := horizonclient.AccountRequest{AccountID: public}
		account, err := client.AccountDetail(request)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to get account: %v", err)}
		}

		var xlmBalance string
		for _, b := range account.Balances {
			if b.Asset.Type == "native" {
				xlmBalance = b.Balance
				break
			}
		}

		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Balance for %s: %s XLM", public, xlmBalance),
			Data:    map[string]interface{}{"balance": xlmBalance, "asset": "XLM"},
		}

	case "send":
		to, _ := payload["to"].(string)
		amount, _ := payload["amount"].(float64)
		if secret == "" {
			return ToolResult{Success: false, Error: "STELLAR_SECRET_KEY not set"}
		}

		sourceKP, err := keypair.ParseFull(secret)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("invalid secret key: %v", err)}
		}

		sourceRequest := horizonclient.AccountRequest{AccountID: sourceKP.Address()}
		sourceAccount, err := client.AccountDetail(sourceRequest)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to get source account: %v", err)}
		}

		tx, err := txnbuild.NewTransaction(
			txnbuild.TransactionParams{
				SourceAccount:        &sourceAccount,
				IncrementSequenceNum: true,
				BaseFee:              txnbuild.MinBaseFee,
				Preconditions: txnbuild.Preconditions{
					TimeBounds: txnbuild.NewInfiniteTimeout(),
				},
				Operations: []txnbuild.Operation{
					&txnbuild.Payment{
						Destination: to,
						Amount:      fmt.Sprintf("%.7f", amount),
						Asset:       txnbuild.NativeAsset{},
					},
				},
			},
		)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to build transaction: %v", err)}
		}

		tx, err = tx.Sign(network.TestNetworkPassphrase, sourceKP)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to sign transaction: %v", err)}
		}

		txe, err := tx.Base64()
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to encode transaction: %v", err)}
		}

		resp, err := client.SubmitTransactionXDR(txe)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to submit transaction: %v", err)}
		}

		log.Printf("[Stellar] Successfully sent %.4f XLM to %s. Hash: %s", amount, to, resp.Hash)
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Successfully sent %.4f XLM to %s. Hash: %s", amount, to, resp.Hash),
			Data:    map[string]interface{}{"tx_hash": resp.Hash},
		}

	default:
		return ToolResult{Success: false, Error: "Invalid stellar action"}
	}
}

func init() {
	RegisterTool("stellar", stellarTool)
}
