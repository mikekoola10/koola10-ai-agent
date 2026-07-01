package tools

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func binanceTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)
	symbol, _ := payload["symbol"].(string)
	if symbol == "" {
		symbol = "BTCUSDT"
	}

	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET")

	client := &http.Client{}

	switch action {
	case "get_price":
		resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", symbol))
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		defer resp.Body.Close()

		var data struct {
			Symbol string `json:"symbol"`
			Price  string `json:"price"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return ToolResult{Success: false, Error: "Failed to decode Binance response"}
		}
		price, _ := strconv.ParseFloat(data.Price, 64)

		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Current price of %s is %s", symbol, data.Price),
			Data:    map[string]interface{}{"symbol": data.Symbol, "price": price},
		}

	case "get_klines":
		interval, _ := payload["interval"].(string)
		if interval == "" {
			interval = "1h"
		}
		limit, _ := payload["limit"].(float64)
		if limit == 0 {
			limit = 10
		}

		apiURL := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=%d", symbol, interval, int(limit))
		resp, err := http.Get(apiURL)
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		defer resp.Body.Close()

		var klines [][]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&klines); err != nil {
			return ToolResult{Success: false, Error: "Failed to decode Binance klines"}
		}

		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Retrieved %d klines for %s", len(klines), symbol),
			Data:    map[string]interface{}{"klines": klines},
		}

	case "paper_trade":
		side, _ := payload["side"].(string)
		amount, _ := payload["amount"].(float64)
		if amount == 0 {
			amount = 0.001 // Default small amount
		}
		if side == "" {
			side = "buy"
		}

		priceRes := binanceTool(map[string]interface{}{"action": "get_price", "symbol": symbol})
		if !priceRes.Success {
			return priceRes
		}
		data := priceRes.Data.(map[string]interface{})
		price := data["price"].(float64)

		output := fmt.Sprintf("Paper %s %.4f %s at real market price %.2f", side, amount, symbol, price)

		return ToolResult{
			Success: true,
			Output:  output,
			Data: map[string]interface{}{
				"side":   side,
				"amount": amount,
				"symbol": symbol,
				"price":  price,
			},
		}

	case "trade", "market_buy", "market_sell":
		side, _ := payload["side"].(string)
		if action == "market_buy" {
			side = "BUY"
		} else if action == "market_sell" {
			side = "SELL"
		}
		if side == "" { side = "BUY" }
		side = strings.ToUpper(side)

		quantity, _ := payload["quantity"].(float64)
		if quantity == 0 {
			quantity, _ = payload["amount"].(float64)
		}
		if quantity == 0 { return ToolResult{Success: false, Error: "quantity is required"} }

		timestamp := time.Now().UnixMilli()
		params := url.Values{}
		params.Set("symbol", symbol)
		params.Set("side", side)
		params.Set("type", "MARKET")
		params.Set("quantity", strconv.FormatFloat(quantity, 'f', -1, 64))
		params.Set("timestamp", strconv.FormatInt(timestamp, 10))

		signature := computeHmac256(params.Encode(), secretKey)
		params.Set("signature", signature)

		req, _ := http.NewRequest("POST", "https://api.binance.com/api/v3/order?"+params.Encode(), nil)
		req.Header.Set("X-MBX-APIKEY", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if resp.StatusCode != http.StatusOK {
			return ToolResult{Success: false, Error: fmt.Sprintf("Binance API error: %v", result)}
		}

		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Successfully executed %s %f %s", side, quantity, symbol),
			Data:    result,
		}

	case "get_balance":
		timestamp := time.Now().UnixMilli()
		params := url.Values{}
		params.Set("timestamp", strconv.FormatInt(timestamp, 10))

		signature := computeHmac256(params.Encode(), secretKey)
		params.Set("signature", signature)

		req, _ := http.NewRequest("GET", "https://api.binance.com/api/v3/account?"+params.Encode(), nil)
		req.Header.Set("X-MBX-APIKEY", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return ToolResult{Success: false, Error: fmt.Sprintf("Binance API error: %s", string(body))}
		}

		var accountInfo map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&accountInfo)

		return ToolResult{
			Success: true,
			Output:  "Successfully retrieved account balance from Binance",
			Data:    accountInfo,
		}

	default:
		return ToolResult{Success: false, Error: "Invalid binance action"}
	}
}

func computeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func init() {
	RegisterTool("binance", binanceTool)
}
