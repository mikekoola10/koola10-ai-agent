package agents

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"bytes"
)

type Transaction struct {
	Timestamp   string  `json:"timestamp"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Vertical    string  `json:"vertical,omitempty"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

type PRInfo struct {
	Repo   string `json:"repo"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Status string `json:"status"`
}

type GrantOpportunity struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	ROI   float64 `json:"roi"`
}

type SystemState interface {
	GetBalance() float64
	GetRecentTransactions(limit int) []Transaction
	GetSwarmStatus() map[string]interface{}
	GetOpenPRs() []PRInfo
	GetComplianceViolations() []string
	GetOverdueFollowups() []string
	GetHighROIGrants() []GrantOpportunity
}

type ProactiveAgent struct {
	SwarmManager *SwarmManager
	Interval     time.Duration
	DeepSeekKey  string
	DeepSeekBase string
	BroadcastFn  func(alert ProactiveAlert)
}

type ProactiveAlert struct {
	ID          string `json:"alert_id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
	Recommendation string `json:"recommendation"`
	Timestamp   string `json:"timestamp"`
}

func NewProactiveAgent(sm *SwarmManager) *ProactiveAgent {
	base := os.Getenv("DEEPSEEK_BASE_URL")
	if base == "" {
		base = "https://api.deepseek.com"
	}
	return &ProactiveAgent{
		SwarmManager: sm,
		Interval:     5 * time.Minute,
		DeepSeekKey:  os.Getenv("DEEPSEEK_API_KEY"),
		DeepSeekBase: base,
	}
}

func (pa *ProactiveAgent) Start(state SystemState) {
	log.Println("Starting Proactive Agent background loop...")
	ticker := time.NewTicker(pa.Interval)
	go func() {
		for {
			pa.CheckThresholds(state)
			<-ticker.C
		}
	}()
}

func (pa *ProactiveAgent) CheckThresholds(state SystemState) {
	log.Println("Proactive Agent: Checking thresholds...")

	balance := state.GetBalance()
	txs := state.GetRecentTransactions(10)
	swarmStatus := state.GetSwarmStatus()
	prs := state.GetOpenPRs()
	violations := state.GetComplianceViolations()
	overdue := state.GetOverdueFollowups()
	grants := state.GetHighROIGrants()

	// Simple drawdown check: if last transaction was a large cost
	drawdownAlert := false
	if len(txs) > 0 {
		last := txs[0]
		if last.Type == "cost" && last.Amount > (balance * 0.05) {
			drawdownAlert = true
		}
	}

	if drawdownAlert || len(violations) > 0 || len(prs) > 0 || len(overdue) > 0 || len(grants) > 0 {
		stateJSON, _ := json.Marshal(map[string]interface{}{
			"balance":               balance,
			"recent_transactions":   txs,
			"swarm_status":          swarmStatus,
			"open_prs":              prs,
			"compliance_violations": violations,
			"overdue_followups":     overdue,
			"high_roi_grants":       grants,
		})
		pa.GenerateAlert(string(stateJSON))
	}
}

func (pa *ProactiveAgent) GenerateAlert(systemState string) {
	if pa.DeepSeekKey == "" {
		return
	}

	prompt := fmt.Sprintf("Analyze the following system state and generate a proactive alert if any critical thresholds are crossed (drawdown > 5%%, overdue follow-ups, high-ROI grants, compliance issues). System State: %s. Return a JSON object for the alert with fields: type, title, message, severity (low|medium|high), recommendation.", systemState)

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "You are the Koola10 Proactive Advisor. Be concise and professional."},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	body, _ := json.Marshal(dsReq)
	req, _ := http.NewRequest("POST", pa.DeepSeekBase+"/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+pa.DeepSeekKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Proactive Agent: DeepSeek call failed: %v", err)
		return
	}
	defer resp.Body.Close()

	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string
			}
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil {
		return
	}

	if len(dsRes.Choices) > 0 {
		var alert ProactiveAlert
		if err := json.Unmarshal([]byte(dsRes.Choices[0].Message.Content), &alert); err == nil {
			alert.ID = fmt.Sprintf("alert_%d", time.Now().Unix())
			alert.Timestamp = time.Now().Format(time.RFC3339)
			if pa.BroadcastFn != nil {
				pa.BroadcastFn(alert)
			}
		}
	}
}
