package orchestrator

import (
	"encoding/json"
	"log"
	"net/http"
)

type A2AMessage struct {
	FromAgent string                 `json:"from_agent"`
	ToAgent   string                 `json:"to_agent"`
	Action    string                 `json:"action"`
	Payload   map[string]interface{} `json:"payload"`
}

type A2AResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type A2ABridge struct {
	Engine *Engine
}

func NewA2ABridge(engine *Engine) *A2ABridge {
	return &A2ABridge{Engine: engine}
}

func (b *A2ABridge) HandleDiscovery(w http.ResponseWriter, r *http.Request) {
	services := []string{"financial_reporting", "swarm_orchestration", "email_automation", "self_healing"}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id": "koola10",
		"version":  "1.5.0",
		"capabilities": services,
	})
}

func (b *A2ABridge) HandleDelegate(w http.ResponseWriter, r *http.Request) {
	var msg A2AMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "bad request", 400)
		return
	}

	log.Printf("[A2A] Received delegation from %s: %s", msg.FromAgent, msg.Action)

	b.Engine.ReportEvent("a2a_bridge", "delegation_received", "Action: "+msg.Action, map[string]interface{}{
		"from":    msg.FromAgent,
		"payload": msg.Payload,
	})

	// Response back to delegating agent
	json.NewEncoder(w).Encode(A2AResponse{
		Status:  "accepted",
		Message: "Task queued for autonomous execution",
	})
}
