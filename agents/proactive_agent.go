package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"koola10/tools"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	ROI    float64 `json:"roi"`
	Amount float64 `json:"amount"`
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
	SwarmManager        *SwarmManager
	Interval            time.Duration
	DeepSeekKey         string
	DeepSeekBase        string
	BroadcastFn         func(alert ProactiveAlert)
	WaitFn              func(alertID string) (string, error)
	AnalyzeFn           func(alert ProactiveAlert) (string, error)
	AnalysisBroadcastFn func(alertID string, analysis string)
	TTSFn               func(text string)
	lastAlerts          map[string]time.Time
}

type ProactiveAlert struct {
	ID             string `json:"alert_id"`
	Type           string `json:"type"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Severity       string `json:"severity"`
	Recommendation string `json:"recommendation"`
	Timestamp      string `json:"timestamp"`
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
		lastAlerts:   make(map[string]time.Time),
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
	prs := state.GetOpenPRs()
	violations := state.GetComplianceViolations()
	overdue := state.GetOverdueFollowups()
	grants := state.GetHighROIGrants()

	// 1. Trading drawdown check
	if len(txs) > 0 {
		last := txs[0]
		if last.Type == "cost" && last.Amount > (balance*0.05) {
			pa.TriggerJarvisAlert("Trading Drawdown", fmt.Sprintf("Drawdown of $%.2f detected, exceeding 5%% threshold.", last.Amount), "High")
		}
	}

	// 2. High-ROI grant check (>$5K)
	for _, g := range grants {
		if g.Amount > 5000 {
			pa.TriggerJarvisAlert("High-ROI Grant Opportunity", fmt.Sprintf("New grant opportunity '%s' found with projected amount of $%.2f.", g.Title, g.Amount), "High")
		}
	}

	// 3. Operational fund balance low (<$10)
	if balance < 10 {
		pa.TriggerJarvisAlert("Low Operational Balance", fmt.Sprintf("Operational fund balance is critically low at $%.2f.", balance), "High")
	}

	// 4. Night Shift PRs (Batch)
	if len(prs) > 0 && time.Now().Hour()%2 == 0 {
		pa.TriggerJarvisAlert("PRs Need Review", fmt.Sprintf("There are %d open pull requests needing review.", len(prs)), "Low")
	}

	// 5. Compliance Violations
	if len(violations) > 0 {
		pa.TriggerJarvisAlert("Compliance Violation", fmt.Sprintf("Detected %d active compliance violations.", len(violations)), "High")
	}

	// 6. Overdue follow-ups
	if len(overdue) > 0 {
		pa.TriggerJarvisAlert("Overdue Follow-ups", fmt.Sprintf("There are %d overdue follow-ups for grant applications.", len(overdue)), "Medium")
	}
}

func (pa *ProactiveAgent) TriggerJarvisAlert(title, message, severity string) {
	if last, ok := pa.lastAlerts[title]; ok && time.Since(last) < 30*time.Minute {
		return
	}
	pa.lastAlerts[title] = time.Now()

	alertID := fmt.Sprintf("alert_%d", time.Now().Unix())
	alert := ProactiveAlert{
		ID:        alertID,
		Type:      "jarvis_notification",
		Title:     title,
		Message:   message,
		Severity:  severity,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if pa.BroadcastFn != nil {
		pa.BroadcastFn(alert)
	}

	speakText := fmt.Sprintf("George, Sterling has a %s opportunity. Would you like to hear more?", strings.ToLower(alert.Title))
	if strings.Contains(strings.ToLower(alert.Title), "drawdown") || strings.Contains(strings.ToLower(alert.Title), "balance") || strings.Contains(strings.ToLower(alert.Title), "compliance") {
		speakText = fmt.Sprintf("George, I have an important alert regarding %s. %s. Would you like to hear more?", strings.ToLower(alert.Title), alert.Message)
	}

	pa.Speak(speakText)

	if pa.WaitFn != nil {
		response, err := pa.WaitFn(alertID)
		if err != nil {
			log.Printf("Proactive Agent: Wait failed for %s: %v", alertID, err)
			return
		}

		responseLower := strings.ToLower(response)
		if strings.Contains(responseLower, "proceed") || strings.Contains(responseLower, "yes") || strings.Contains(responseLower, "tell me more") {
			pa.RunConversationLoop(alert)
		} else if strings.Contains(responseLower, "handle it") {
			pa.HandleAction(alert)
		}
	}
}

func (pa *ProactiveAgent) RunConversationLoop(alert ProactiveAlert) {
	analysis, err := pa.AnalyzeFn(alert)
	if err != nil {
		log.Printf("Proactive Agent: Analysis failed: %v", err)
		pa.Speak("I am sorry, I encountered an error while analyzing the situation.")
		return
	}

	if pa.AnalysisBroadcastFn != nil {
		pa.AnalysisBroadcastFn(alert.ID, analysis)
	}

	pa.Speak(analysis)
	pa.Speak("Would you like me to handle this?")

	if pa.WaitFn != nil {
		response, err := pa.WaitFn(alert.ID + "_handle")
		if err == nil {
			responseLower := strings.ToLower(response)
			if strings.Contains(responseLower, "yes") || strings.Contains(responseLower, "proceed") || strings.Contains(responseLower, "do it") || strings.Contains(responseLower, "handle it") {
				pa.HandleAction(alert)
			}
		}
	}
}

func (pa *ProactiveAgent) HandleAction(alert ProactiveAlert) {
	pa.Speak("Executing the recommended action.")
	log.Printf("Proactive Agent: Executing approved action for %s", alert.Title)
	// Execution logic would go here
	pa.Speak("Action completed successfully.")
}

func (pa *ProactiveAgent) Speak(text string) {
	if pa.TTSFn != nil {
		pa.TTSFn(text)
	} else {
		tools.RunTool("notifications", map[string]interface{}{
			"action": "send_tts_prompt",
			"text":   text,
		})
	}
}

func (pa *ProactiveAgent) GenerateAlert(systemState string) {
	if pa.DeepSeekKey == "" {
		return
	}

	prompt := fmt.Sprintf("Analyze system state: %s. Generate proactive alert if thresholds crossed. Return JSON: type, title, message, severity, recommendation.", systemState)

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "You are the Koola10 Proactive Advisor."},
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
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err == nil && len(dsRes.Choices) > 0 {
		var alert ProactiveAlert
		if err := json.Unmarshal([]byte(dsRes.Choices[0].Message.Content), &alert); err == nil {
			alert.ID = fmt.Sprintf("alert_%d", time.Now().Unix())
			alert.Timestamp = time.Now().Format(time.RFC3339)
			pa.TriggerJarvisAlert(alert.Title, alert.Message, alert.Severity)
		}
	}
}
