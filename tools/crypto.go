package tools

import (
	"fmt"
)

func cryptoTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)
	symbol, _ := payload["symbol"].(string)
	amount, _ := payload["amount"].(float64)

	switch action {
	case "price":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Current price of %s is $50000.00 (Paper Trading)", symbol),
			Data:    map[string]interface{}{"symbol": symbol, "price": 50000.00},
		}
	case "buy":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Paper bought %.4f %s", amount, symbol),
			Data:    map[string]interface{}{"symbol": symbol, "amount": amount, "side": "buy"},
		}
	case "sell":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Paper sold %.4f %s", amount, symbol),
			Data:    map[string]interface{}{"symbol": symbol, "amount": amount, "side": "sell"},
		}
	default:
		return ToolResult{Success: false, Error: "Invalid crypto action"}
	}
}

func init() {
	RegisterTool("crypto", cryptoTool)
}
