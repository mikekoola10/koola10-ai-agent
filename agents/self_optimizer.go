package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type SelfOptimizer struct {
	DeepSeekKey  string
	DeepSeekBase string
	Swarm        *SwarmManager
}

func NewSelfOptimizer(sm *SwarmManager) *SelfOptimizer {
	base := os.Getenv("DEEPSEEK_BASE_URL")
	if base == "" {
		base = "https://api.deepseek.com"
	}
	return &SelfOptimizer{
		DeepSeekKey:  os.Getenv("DEEPSEEK_API_KEY"),
		DeepSeekBase: base,
		Swarm:        sm,
	}
}

func (so *SelfOptimizer) WeeklyPass() {
	log.Println("Self Optimizer: Starting weekly optimization pass...")
	if so.DeepSeekKey == "" {
		return
	}

	metrics := so.Swarm.GetAllSwarmMetrics()
	metricsJSON, _ := json.Marshal(metrics)

	prompt := fmt.Sprintf("Analyze the weekly performance metrics for the Koola10 agent swarm. Metrics: %s. Model the swarm as a Textual Parameter Graph where each vertical's prompt and workflow are optimizable nodes. Propose systemic prompt and workflow optimizations to improve success rates and ROI. Return JSON with 'proposals' (array of {vertical, optimization, justification}).", string(metricsJSON))

	dsReq := map[string]interface{}{
		"model":           "deepseek-chat",
		"messages":        []map[string]string{{"role": "user", "content": prompt}},
		"response_format": map[string]string{"type": "json_object"},
	}
	body, _ := json.Marshal(dsReq)
	req, _ := http.NewRequest("POST", so.DeepSeekBase+"/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+so.DeepSeekKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
		var dsRes struct{ Choices []struct{ Message struct{ Content string } } }
		if json.NewDecoder(resp.Body).Decode(&dsRes) == nil && len(dsRes.Choices) > 0 {
			var proposals map[string]interface{}
			if json.Unmarshal([]byte(dsRes.Choices[0].Message.Content), &proposals) == nil {
				os.WriteFile("/data/optimizations_"+time.Now().Format("2006-01-02")+".json", []byte(dsRes.Choices[0].Message.Content), 0644)
				log.Printf("Self Optimizer: Weekly proposals generated and saved.")
			}
		}
	}
}
