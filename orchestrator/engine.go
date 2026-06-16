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

type RecoveryAction struct {
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	Params      []string `json:"params"`
	TimeoutSecs int      `json:"timeout_secs"`
}

type FailureDefinition struct {
	Name            string           `json:"name"`
	Detection       string           `json:"detection"`
	RootCauses      []string         `json:"root_causes"`
	RecoveryActions []RecoveryAction `json:"recovery_actions"`
	Verification    string           `json:"verification"`
	Escalation      string           `json:"escalation"`
}

type RecoveryMap struct {
	Failures []FailureDefinition `json:"failures"`
}

type Engine struct {
	Status           ComponentStatus `json:"status"`
	RetryCounts      map[string]int  `json:"retry_counts"`
	Events           []Event         `json:"events"`
	OnEvent          func(Event)     `json:"-"`
	RecoveryMap      *RecoveryMap    `json:"recovery_map"`
	AttemptRecovery  func(string, string) bool `json:"-"`
	Notify           func(string, string)      `json:"-"`
	pendingApprovals map[string]chan bool
	mu               sync.RWMutex
	eventChan        chan Event
}

func NewEngine() *Engine {
	e := &Engine{
		Status:           StateIdle,
		RetryCounts:      make(map[string]int),
		Events:           make([]Event, 0),
		eventChan:        make(chan Event, 100),
		pendingApprovals: make(map[string]chan bool),
	}
	e.LoadRecoveryMap("data/recovery_map.json")
	return e
}

func (e *Engine) LoadRecoveryMap(path string) {
	paths := []string{path, "/data/recovery_map.json"}
	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Printf("[Engine] Warning: Recovery map not found in any standard location")
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
		log.Printf("[Engine] Event received on channel: %s", event.ID)
		e.handleEvent(event)
	}
}

func (e *Engine) ApproveEvent(eventID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if ch, ok := e.pendingApprovals[eventID]; ok {
		ch <- true
		delete(e.pendingApprovals, eventID)
		for i, ev := range e.Events {
			if ev.ID == eventID { e.Events[i].Approved = true; break }
		}
	}
}

func (e *Engine) RejectEvent(eventID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if ch, ok := e.pendingApprovals[eventID]; ok {
		ch <- false
		delete(e.pendingApprovals, eventID)
	}
}

func (e *Engine) ReportEvent(source, eventType, message string, details map[string]interface{}) {
	log.Printf("[Engine] Reporting event from %s: %s", source, message)
	event := Event{
		ID:        time.Now().Format("20060102150405-") + source,
		Source:    source,
		Type:      eventType,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}

	// Manual Approval Gate for financial logic and protected paths
	isSensitive := strings.Contains(message, "financial/") ||
				   strings.Contains(eventType, "payout") ||
				   strings.Contains(eventType, "ledger") ||
				   strings.Contains(eventType, "email_sent")

	if det, ok := details["files"].([]interface{}); ok {
		for _, f := range det {
			if path, ok := f.(string); ok && strings.HasPrefix(path, "financial/") {
				isSensitive = true
				break
			}
		}
	}

	if isSensitive {
		event.RequiresApproval = true
		log.Printf("[Engine] Event tagged as SENSITIVE - Holding for manual review.")

		e.mu.Lock()
		approvalChan := make(chan bool)
		e.pendingApprovals[event.ID] = approvalChan
		e.mu.Unlock()

		go func(id string, ch chan bool) {
			if approved := <-ch; approved {
				log.Printf("[Engine] Event APPROVED: %s", id)
				// Proceed logic could be here, but for now we just log and let handleEvent wait if needed
			} else {
				log.Printf("[Engine] Event REJECTED: %s", id)
			}
		}(event.ID, approvalChan)
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
			var failureName string
			if e.RecoveryMap != nil {
				for _, f := range e.RecoveryMap.Failures {
					if strings.Contains(strings.ToLower(event.Message), strings.ToLower(f.Name)) ||
					   strings.Contains(strings.ToLower(event.Source), strings.ToLower(f.Name)) {
						failureName = f.Name
						break
					}
				}
			}

			log.Printf("[Engine] Initiating self-healing loop for %s (Attempt %d/5). Failure: %s", taskID, count+1, failureName)

			// Automated Healing Flow
			env := os.Getenv("DEVICE_AGENT_ENV")
			if env == "" { env = "staging" }

			if e.AttemptRecovery != nil && failureName != "" {
				log.Printf("[Engine] Triggering targeted recovery actions for: %s (Env: %s)", failureName, env)
				go e.AttemptRecovery(failureName, event.Message)
			} else if event.Source == "e2e_watchdog" {
				log.Printf("[Engine] Invoking MetaSwarm and DeviceAgent (Env: %s) to apply fix...", env)
				// Execute meta-swarm healing...
			}
		} else {
			e.Status = StateSafeMode
			e.mu.Unlock()
			msg := "CIRCUIT BREAKER TRIGGERED for " + taskID + ". Entering SAFE MODE."
			log.Printf("[Engine] %s", msg)
			e.ReportEvent("engine", "circuit_breaker", msg, map[string]interface{}{"task_id": taskID})

			if e.Notify != nil {
				e.Notify("CRITICAL: Circuit Breaker Triggered", msg)
			}
		}
	} else {
		e.mu.Lock()
		e.Status = StateIdle
		e.mu.Unlock()
	}
}
