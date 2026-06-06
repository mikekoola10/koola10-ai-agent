package tools

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// CCXTClient wraps the universal exchange library logic
type CCXTClient struct {
	exchange string
	apiKey   string
	secret   string
}

// NewCCXTClient creates a client for any supported exchange
func NewCCXTClient(exchange string) (*CCXTClient, error) {
	apiKey := os.Getenv(fmt.Sprintf("%s_API_KEY", strings.ToUpper(exchange)))
	secret := os.Getenv(fmt.Sprintf("%s_SECRET_KEY", strings.ToUpper(exchange)))

	// For paper trading and public info, we allow missing keys for some actions
	return &CCXTClient{
		exchange: exchange,
		apiKey:   apiKey,
		secret:   secret,
	}, nil
}

// ccxtTool is the registered tool function
func ccxtTool(payload map[string]interface{}) ToolResult {
	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "missing action parameter"}
	}

	exchange, _ := payload["exchange"].(string)
	if exchange == "" {
		exchange = "binance" // default
	}

	client, err := NewCCXTClient(exchange)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	switch action {
	case "get_price":
		symbol, _ := payload["symbol"].(string)
		if symbol == "" {
			symbol = "BTC/USDT"
		}
		price, err := client.fetchTicker(symbol)
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		return ToolResult{Success: true, Data: map[string]interface{}{
			"exchange": exchange,
			"symbol":   symbol,
			"price":    price,
		}}

	case "get_balance":
		if client.apiKey == "" {
			return ToolResult{Success: false, Error: fmt.Sprintf("missing API keys for exchange: %s", exchange)}
		}
		balance, err := client.fetchBalance()
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		return ToolResult{Success: true, Data: balance}

	case "paper_trade":
		symbol, _ := payload["symbol"].(string)
		if symbol == "" {
			symbol = "BTC/USDT"
		}
		side, _ := payload["side"].(string)
		if side == "" {
			side = "buy"
		}
		quantity, _ := payload["quantity"].(float64)
		if quantity == 0 {
			quantity = 0.01
		}

		price, err := client.fetchTicker(symbol)
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}

		// Log paper trade to economic ledger simulation
		log.Printf("[CCXT] Paper trade: %s %s %f %s @ %f", exchange, side, quantity, symbol, price)

		return ToolResult{Success: true, Data: map[string]interface{}{
			"exchange": exchange,
			"symbol":   symbol,
			"side":     side,
			"quantity": quantity,
			"price":    price,
			"mode":     "paper_trade",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}}

	case "list_exchanges":
		// Return all supported exchanges
		exchanges := []string{
			"binance", "coinbase", "kraken", "okx", "bybit",
			"gate", "bitget", "bingx", "huobi", "kucoin",
		}
		return ToolResult{Success: true, Data: map[string]interface{}{
			"supported_exchanges": exchanges,
			"total":               len(exchanges),
		}}

	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("unknown action: %s", action)}
	}
}

// fetchTicker gets the current price for a symbol
func (c *CCXTClient) fetchTicker(symbol string) (float64, error) {
	// In production, this would call the actual CCXT Go library.
	// We'll use a unified approach for simulation/public APIs.

	// Handle symbol formatting (CCXT uses BTC/USDT, APIs often use BTCUSDT)
	cleanSymbol := strings.ReplaceAll(symbol, "/", "")

	var apiURL string
	switch c.exchange {
	case "binance":
		apiURL = fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", cleanSymbol)
	case "kraken":
		apiURL = fmt.Sprintf("https://api.kraken.com/0/public/Ticker?pair=%s", cleanSymbol)
	case "coinbase":
		apiURL = fmt.Sprintf("https://api.exchange.coinbase.com/products/%s/ticker", strings.ReplaceAll(symbol, "/", "-"))
	default:
		// Fallback to binance if unknown exchange for price
		apiURL = fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", cleanSymbol)
	}

	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "Koola10-Agent/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("exchange API error: %d", resp.StatusCode)
	}

	var priceStr string
	switch c.exchange {
	case "binance":
		var result struct{ Price string }
		json.NewDecoder(resp.Body).Decode(&result)
		priceStr = result.Price
	case "kraken":
		var result struct {
			Result map[string]interface{} `json:"result"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		for _, v := range result.Result {
			m := v.(map[string]interface{})
			c := m["c"].([]interface{})
			priceStr = c[0].(string)
			break
		}
	case "coinbase":
		var result struct{ Price string }
		json.NewDecoder(resp.Body).Decode(&result)
		priceStr = result.Price
	default:
		var result struct{ Price string }
		json.NewDecoder(resp.Body).Decode(&result)
		priceStr = result.Price
	}

	return strconv.ParseFloat(priceStr, 64)
}

// fetchBalance gets account balance
func (c *CCXTClient) fetchBalance() (map[string]interface{}, error) {
	// In production, this calls the CCXT Go library with HMAC signatures
	return map[string]interface{}{
		"BTC":  0.0,
		"USDT": 1000.0, // paper trading balance
		"mode": "paper_trading_sim",
	}, nil
}

func init() {
	RegisterTool("ccxt", ccxtTool)
}
