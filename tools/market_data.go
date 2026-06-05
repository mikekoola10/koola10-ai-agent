package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func marketDataTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)

	switch action {
	case "get_stock_price":
		symbol, _ := payload["symbol"].(string)
		if symbol == "" {
			return ToolResult{Success: false, Error: "symbol is required"}
		}
		url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s", symbol)
		resp, err := http.Get(url)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to fetch stock price: %v", err)}
		}
		defer resp.Body.Close()

		var data struct {
			Chart struct {
				Result []struct {
					Meta struct {
						RegularMarketPrice float64 `json:"regularMarketPrice"`
						PreviousClose      float64 `json:"previousClose"`
						Symbol             string  `json:"symbol"`
					} `json:"meta"`
					Indicators struct {
						Quote []struct {
							Volume []int64 `json:"volume"`
						} `json:"quote"`
					} `json:"indicators"`
				} `json:"result"`
			} `json:"chart"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to decode response: %v", err)}
		}

		if len(data.Chart.Result) == 0 {
			return ToolResult{Success: false, Error: "no data found for symbol"}
		}

		res := data.Chart.Result[0]
		price := res.Meta.RegularMarketPrice
		change := price - res.Meta.PreviousClose
		var volume int64
		if len(res.Indicators.Quote) > 0 && len(res.Indicators.Quote[0].Volume) > 0 {
			volume = res.Indicators.Quote[0].Volume[len(res.Indicators.Quote[0].Volume)-1]
		}

		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Price for %s: %.2f (Change: %.2f, Volume: %d)", symbol, price, change, volume),
			Data: map[string]interface{}{
				"symbol": symbol,
				"price":  price,
				"change": change,
				"volume": volume,
			},
		}

	case "get_crypto_price":
		symbol, _ := payload["symbol"].(string) // e.g. BTC
		if symbol == "" {
			return ToolResult{Success: false, Error: "symbol is required"}
		}

		res := RunTool("strike", map[string]interface{}{
			"action": "get_exchange_rates",
		})
		if !res.Success {
			return res
		}

		rates, ok := res.Data.([]interface{})
		if !ok {
			return ToolResult{Success: false, Error: "failed to parse Strike rates"}
		}

		for _, r := range rates {
			rate, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			if rate["currency"] == symbol {
				priceStr, _ := rate["amount"].(string)
				var price float64
				fmt.Sscanf(priceStr, "%f", &price)
				return ToolResult{
					Success: true,
					Output:  fmt.Sprintf("Price for %s: %.2f USD", symbol, price),
					Data: map[string]interface{}{
						"symbol": symbol,
						"price":  price,
					},
				}
			}
		}

		return ToolResult{Success: false, Error: fmt.Sprintf("could not find rate for %s", symbol)}

	case "search_news":
		query, _ := payload["query"].(string)
		apiKey := os.Getenv("NEWSAPI_KEY")
		if apiKey == "" {
			return ToolResult{Success: false, Error: "NEWSAPI_KEY not set"}
		}
		apiUrl := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&apiKey=%s", url.QueryEscape(query), apiKey)
		resp, err := http.Get(apiUrl)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to fetch news: %v", err)}
		}
		defer resp.Body.Close()

		var newsData struct {
			Articles []struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				Url         string `json:"url"`
				PublishedAt string `json:"publishedAt"`
			} `json:"articles"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&newsData); err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to decode news response: %v", err)}
		}

		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Found %d articles for %s", len(newsData.Articles), query),
			Data:    newsData.Articles,
		}

	case "get_ipo_details":
		symbol, _ := payload["symbol"].(string)
		if symbol == "" {
			return ToolResult{Success: false, Error: "symbol is required"}
		}
		// Try to find IPO info via NewsAPI if it's the IPO vertical
		apiKey := os.Getenv("NEWSAPI_KEY")
		query := symbol + " IPO expected price valuation"
		apiUrl := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&apiKey=%s", url.QueryEscape(query), apiKey)
		resp, err := http.Get(apiUrl)

		summary := fmt.Sprintf("No specific IPO data found for %s via public trackers.", symbol)
		if err == nil {
			defer resp.Body.Close()
			var newsData struct {
				Articles []struct {
					Title       string `json:"title"`
					Description string `json:"description"`
				} `json:"articles"`
			}
			if json.NewDecoder(resp.Body).Decode(&newsData) == nil && len(newsData.Articles) > 0 {
				summary = fmt.Sprintf("Recent IPO news for %s: %s", symbol, newsData.Articles[0].Title)
			}
		}

		return ToolResult{
			Success: true,
			Output:  summary,
			Data: map[string]interface{}{
				"symbol":       symbol,
				"status":       "tracking",
				"last_mention": summary,
			},
		}

	default:
		return ToolResult{Success: false, Error: "invalid action"}
	}
}

func init() {
	RegisterTool("market_data", marketDataTool)
}
