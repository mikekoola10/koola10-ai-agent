package agents

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

func CallDeepSeek(prompt string, system string) (string, int) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return "DeepSeek API key not set", 0
	}

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": prompt},
		},
	}
	dsBody, _ := json.Marshal(dsReq)
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	hReq.Header.Set("Authorization", "Bearer "+apiKey)
	hReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(hReq)
	if err != nil {
		return "API request failed", 0
	}
	defer resp.Body.Close()

	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string
			}
		}
		Usage struct {
			TotalTokens int
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil {
		return "Failed to parse API response", 0
	}

	if len(dsRes.Choices) > 0 {
		return dsRes.Choices[0].Message.Content, dsRes.Usage.TotalTokens
	}
	return "No response from AI", 0
}
