package mirror

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

func (m *Mirror) Recall(context string) ([]string, error) {
	url := os.Getenv("SEMANTIC_AGENT_URL")
	if url == "" {
		url = "https://koola10-semantic.fly.dev"
	}

	// Search scoped to user via query prefix
	query := m.UserID + ": " + context
	reqBody, _ := json.Marshal(map[string]interface{}{
		"query": query,
		"top_k": 5,
	})

	resp, err := http.Post(url+"/search", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	var memories []string
	for _, r := range results {
		memories = append(memories, r.Text)
	}
	return memories, nil
}

func (m *Mirror) Remember(text string, refID string) error {
	url := os.Getenv("SEMANTIC_AGENT_URL")
	if url == "" {
		url = "https://koola10-semantic.fly.dev"
	}

	// Scope interaction to user
	scopedText := m.UserID + ": " + text
	reqBody, _ := json.Marshal(map[string]string{
		"text":   scopedText,
		"ref_id": refID,
	})

	_, err := http.Post(url+"/index", "application/json", bytes.NewBuffer(reqBody))
	return err
}
