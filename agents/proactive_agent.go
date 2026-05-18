package agents

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"bytes"
	"koola10/tools"
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
	SwarmManager *SwarmManager
	Interval     time.Duration
	DeepSeekKey  string
	DeepSeekBase string
	BroadcastFn  func(alert ProactiveAlert)
	WaitFn       func(alertID string) (string, error)
	AnalyzeFn    func(alert ProactiveAlert) (string, error)
	AnalysisBroadcastFn func(alertID string, analysis string)
	TTSFn        func(text string)
	lastAlerts   map[string]time.Time
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
	_ = state.GetSwarmStatus()
	prs := state.GetOpenPRs()
	violations := state.GetComplianceViolations()
	overdue := state.GetOverdueFollowups()
	grants := state.GetHighROIGrants()

	// 1. Trading drawdown check: if last transaction was a large cost
	if len(txs) > 0 {
		last := txs[0]
		if last.Type == "cost" && last.Amount > (balance * 0.05) {
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

	// 6. Lead responds to outreach (Simulated or from state)
	// 7. Optimizr subscription cancelled (From stripe events in state if available)
	// 8. Daily revenue report ready (Scheduled at 9 AM)
	if time.Now().Hour() == 9 && time.Now().Minute() < 5 {
		pa.TriggerJarvisAlert("Daily Revenue Report", "The daily revenue report is ready for your review.", "Low")
	}

	// General catch-all for GenerateAlert logic if needed
	if len(overdue) > 0 {
		stateJSON, _ := json.Marshal(map[string]interface{}{
			"overdue_followups": overdue,
		})
		pa.GenerateAlert(string(stateJSON))
	}
}

func (pa *ProactiveAgent) TriggerJarvisAlert(title, message, severity string) {
	// Deduplication: Don't repeat the same alert within 30 minutes
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

	// 1. Broadcast Alert to SSE (Dashboard will show notification & Notification API)
	if pa.BroadcastFn != nil {
		pa.BroadcastFn(alert)
	}

	// 2. TTS Prompt
	if pa.TTSFn != nil {
		pa.TTSFn(fmt.Sprintf("Excuse me, CEO. I have an important update regarding %s. %s. Would you like to hear more?", alert.Title, alert.Message))
	} else {
		tools.RunTool("notifications", map[string]interface{}{
			"action": "send_tts_prompt",
			"text":   fmt.Sprintf("Excuse me, CEO. I have an important update regarding %s. %s. Would you like to hear more?", alert.Title, alert.Message),
		})
	}

	// 3. Wait for verbal confirmation
	if pa.WaitFn != nil {
		response, err := pa.WaitFn(alertID)
		if err != nil {
			log.Printf("Proactive Agent: Wait failed for %s: %v", alertID, err)
			return
		}

		confirmationPhrases := []string{"proceed", "yes", "go ahead", "tell me more", "what is it"}
		confirmed := false
		for _, p := range confirmationPhrases {
			if strings.Contains(strings.ToLower(response), p) {
				confirmed = true
				break
			}
		}

		if confirmed {
			log.Printf("Proactive Agent: Confirmation received for %s, starting analysis", alertID)
			pa.RunConversationLoop(alert)
		} else {
			log.Printf("Proactive Agent: Confirmation rejected for %s: %s", alertID, response)
		}
	}
}

func (pa *ProactiveAgent) RunConversationLoop(alert ProactiveAlert) {
	// 4. Start conversation with DeepSeek
	analysis, err := pa.AnalyzeFn(alert)
	if err != nil {
		log.Printf("Proactive Agent: Analysis failed: %v", err)
		return
	}

	// Broadcast analysis to dashboard
	if pa.AnalysisBroadcastFn != nil {
		pa.AnalysisBroadcastFn(alert.ID, analysis)
	}

	// 5. Stream DeepSeek's response back through TTS
	if pa.TTSFn != nil {
		pa.TTSFn(analysis)
	}

	// 6. Ask for permission to handle
	pa.TTSFn("Would you like me to handle this?")

	// 7. Wait for verbal approval
	response, err := pa.WaitFn(alert.ID + "_handle")
	if err == nil {
		approvalPhrases := []string{"yes", "proceed", "go ahead", "do it", "handle it"}
		approved := false
		for _, p := range approvalPhrases {
			if strings.Contains(strings.ToLower(response), p) {
				approved = true
				break
			}
		}

		if approved {
			pa.TTSFn("Understood. Executing recommended action.")
			// Execute recommendation (simplified: log for now or call appropriate tool)
			log.Printf("Proactive Agent: Executing approved action: %s", alert.Recommendation)
		}
	}
}

func (pa *ProactiveAgent) GenerateAlert(systemState string) {
	if pa.DeepSeekKey == "" {
		return
	}

	prompt := fmt.Sprintf("Analyze the following system state and generate a proactive alert if any critical thresholds are crossed (drawdown > 5%%, overdue follow-ups, high-ROI grants, compliance issues). System State: %s. Return a JSON object for the alert with fields: type, title, message, severity (low|medium|high), recommendation. The 'message' MUST explain 'what happened' and 'why it matters'. The 'recommendation' MUST explain 'what action is recommended'.", systemState)

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
			pa.TriggerJarvisAlert(alert.Title, alert.Message, alert.Severity)
		}
	}
}
