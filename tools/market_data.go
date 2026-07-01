package tools

import (
	"fmt"
)

func marketDataTool(payload map[string]interface{}) ToolResult {
	symbol, _ := payload["symbol"].(string)

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Market data for %s: Bid 49950.00, Ask 50050.00", symbol),
		Data:    map[string]interface{}{"bid": 49950.00, "ask": 50050.00, "spread": 100.00},
	}
}

func init() {
	RegisterTool("market_data", marketDataTool)
}
