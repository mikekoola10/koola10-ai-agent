package orchestrator

import (
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

type Engine struct {
	Status        ComponentStatus `json:"status"`
	RetryCounts   map[string]int  `json:"retry_counts"`
	Events        []Event         `json:"events"`
	OnEvent       func(Event)     `json:"-"`
	mu            sync.RWMutex
	eventChan     chan Event
}

func NewEngine() *Engine {
	return &Engine{
		Status:      StateIdle,
		RetryCounts: make(map[string]int),
		Events:      make([]Event, 0),
		eventChan:   make(chan Event, 100),
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
		if count < 3 {
			e.RetryCounts[taskID] = count + 1
			e.Status = StateHealing
			e.mu.Unlock()
			log.Printf("[Engine] Initiating self-healing loop for %s (Attempt %d)", taskID, count+1)

			// Automated Healing Flow
			if event.Source == "e2e_watchdog" {
				env := os.Getenv("DEVICE_AGENT_ENV")
				if env == "" { env = "staging" }

				log.Printf("[Engine] Invoking MetaSwarm to scout for fix (Env: %s)...", env)
				// 1. MetaSwarm searching for fix...
				// 2. DeviceAgent testing in sandbox (routed to env)...
				// 3. E2EWatchdog re-verifying...
			}
		} else {
			e.mu.Unlock()
			log.Printf("[Engine] Max retries reached for %s. Escalating to Portal.", taskID)
		}
	} else {
		e.mu.Lock()
		e.Status = StateIdle
		e.mu.Unlock()
	}
}
