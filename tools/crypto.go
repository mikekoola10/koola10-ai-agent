package tools

import (
	"fmt"
	"time"
)

func init() {
	RegisterTool("crypto", CryptoTool)
}

func CryptoTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)
	symbol, _ := payload["symbol"].(string)
	amount, _ := payload["amount"].(float64)

	// Mock paper trading logic
	switch action {
	case "buy":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Paper traded: Bought %f of %s", amount, symbol),
			Data: map[string]interface{}{
				"action":    "buy",
				"symbol":    symbol,
				"amount":    amount,
				"price":     50000.0, // Mock price
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}
	case "sell":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Paper traded: Sold %f of %s", amount, symbol),
			Data: map[string]interface{}{
				"action":    "sell",
				"symbol":    symbol,
				"amount":    amount,
				"price":     51000.0, // Mock price
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}
	case "price":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Current price of %s: $50000.00", symbol),
			Data: map[string]interface{}{
				"symbol": symbol,
				"price":  50000.0,
			},
		}
	default:
		return ToolResult{
			Success: false,
			Error:   "Invalid action for crypto tool",
		}
	}
}
