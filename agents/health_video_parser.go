package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"koola10/tools"
	"net/http"
	"os"
)

type VideoParserAgent struct {
	HealthAgent
}

func (a *VideoParserAgent) Run(url string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	// 1. Scrape content via browser-agent
	extractRes := tools.RunTool("browser", map[string]interface{}{
		"action":      "extract",
		"url":         url,
		"instruction": "Extract the video transcript or description text.",
	})

	if !extractRes.Success {
		return nil, fmt.Errorf("scraping failed: %s", extractRes.Error)
	}

	extractedData, _ := extractRes.Data.(map[string]interface{})
	text, _ := extractedData["data"].(string)

	// 2. Call DeepSeek for structuring
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY not set")
	}

	prompt := fmt.Sprintf("Extract all ingredients, dosages, and supplement names from this text: %s", text)
	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a nutritional analyst. Return structured data in JSON format with an 'ingredients' key containing an array of objects with 'name', 'dosage', and 'frequency'."},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	body, _ := json.Marshal(dsReq)
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(body))
	hReq.Header.Set("Authorization", "Bearer "+apiKey)
	hReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(hReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string
			}
		}
	}
	json.NewDecoder(resp.Body).Decode(&dsRes)

	var result struct {
		Ingredients []Supplement `json:"ingredients"`
	}
	json.Unmarshal([]byte(dsRes.Choices[0].Message.Content), &result)

	// 3. Store result
	video := VideoLink{
		URL:         url,
		Ingredients: result.Ingredients,
	}
	videos, _ := LoadVideos()
	videos = append(videos, video)
	SaveVideos(videos)

	return video, nil
}
