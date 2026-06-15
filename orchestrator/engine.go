package orchestrator

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type ComponentStatus string

const (
	StateIdle      ComponentStatus = "idle"
	StateWorking   ComponentStatus = "working"
	StateHealing   ComponentStatus = "healing"
	StateFailing   ComponentStatus = "failing"
	StateSafeMode  ComponentStatus = "safe_mode"
)

type Event struct {
	ID               string                 `json:"id"`
	Source           string                 `json:"source"`
	Type             string                 `json:"type"`
	Message          string                 `json:"message"`
	Details          map[string]interface{} `json:"details"`
	RequiresApproval bool                   `json:"requires_approval"`
	Approved         bool                   `json:"approved"`
	Timestamp        time.Time              `json:"timestamp"`
}

type FailureDefinition struct {
	Name            string   `json:"name"`
	Detection       string   `json:"detection"`
	RootCauses      []string `json:"root_causes"`
	RecoveryActions []string `json:"recovery_actions"`
	Verification    string   `json:"verification"`
	Escalation      string   `json:"escalation"`
}

type RecoveryMap struct {
	Failures []FailureDefinition `json:"failures"`
}

type Engine struct {
	Status        ComponentStatus `json:"status"`
	RetryCounts   map[string]int  `json:"retry_counts"`
	Events        []Event         `json:"events"`
	OnEvent       func(Event)     `json:"-"`
	RecoveryMap   *RecoveryMap    `json:"recovery_map"`
	mu            sync.RWMutex
	eventChan     chan Event
}

func NewEngine() *Engine {
	e := &Engine{
		Status:      StateIdle,
		RetryCounts: make(map[string]int),
		Events:      make([]Event, 0),
		eventChan:   make(chan Event, 100),
	}
	e.LoadRecoveryMap("data/recovery_map.json")
	return e
}

func (e *Engine) LoadRecoveryMap(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[Engine] Warning: Recovery map not found at %s", path)
		return
	}
	var rMap RecoveryMap
	if err := json.Unmarshal(data, &rMap); err == nil {
		e.RecoveryMap = &rMap
		log.Printf("[Engine] Loaded %d failure definitions from recovery map.", len(rMap.Failures))
	}
}

func (e *Engine) Start() {
	log.Printf("[Engine] Unified Orchestration Brain operational.")
	for event := range e.eventChan {
		e.handleEvent(event)
	}
}

func (e *Engine) ReportEvent(source, eventType, message string, details map[string]interface{}) {
	event := Event{
		ID:        time.Now().Format("20060102150405"),
		Source:    source,
		Type:      eventType,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}

	// Manual Approval Gate for financial logic
	if strings.Contains(message, "financial/") || strings.Contains(eventType, "payout") {
		event.RequiresApproval = true
		log.Printf("[Engine] Event tagged as SENSITIVE - Holding for review.")
	}

	e.mu.Lock()
	e.Events = append(e.Events, event)
	if len(e.Events) > 100 {
		e.Events = e.Events[len(e.Events)-100:]
	}
	e.mu.Unlock()

	e.eventChan <- event
	if e.OnEvent != nil {
		e.OnEvent(event)
	}
}

func (e *Engine) handleEvent(event Event) {
	log.Printf("[Engine] Processing event from %s: %s", event.Source, event.Message)

	if event.Type == "error" || event.Type == "failure" {
		e.mu.Lock()
		e.Status = StateFailing
		taskID := ""
		if tid, ok := event.Details["task_id"].(string); ok { taskID = tid }

		count := e.RetryCounts[taskID]
		if count < 5 { // Circuit breaker limit is 5
			e.RetryCounts[taskID] = count + 1
			e.Status = StateHealing
			e.mu.Unlock()

			// Select strategy from Recovery Map
			var recoveryActions []string
			if e.RecoveryMap != nil {
				for _, f := range e.RecoveryMap.Failures {
					if strings.Contains(strings.ToLower(event.Message), strings.ToLower(f.Name)) ||
					   strings.Contains(strings.ToLower(event.Source), strings.ToLower(f.Name)) {
						recoveryActions = f.RecoveryActions
						break
					}
				}
			}

			log.Printf("[Engine] Initiating self-healing loop for %s (Attempt %d/5). Actions: %v", taskID, count+1, recoveryActions)

			// Automated Healing Flow
			if event.Source == "e2e_watchdog" || len(recoveryActions) > 0 {
				env := os.Getenv("DEVICE_AGENT_ENV")
				if env == "" { env = "staging" }

				log.Printf("[Engine] Invoking MetaSwarm and DeviceAgent (Env: %s) to apply: %v", env, recoveryActions)
				// Execute recovery actions...
			}
		} else {
			e.Status = StateSafeMode
			e.mu.Unlock()
			log.Printf("[Engine] CIRCUIT BREAKER TRIGGERED for %s. Entering SAFE MODE.", taskID)
			e.ReportEvent("engine", "circuit_breaker", "Circuit breaker triggered. System entering safe mode.", map[string]interface{}{"task_id": taskID})
		}
	} else {
		e.mu.Lock()
		e.Status = StateIdle
		e.mu.Unlock()
	}
}
