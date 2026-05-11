package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type CollaborationEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "decision", "advisor_note", "action", "alert"
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type CollaborationMemory struct {
	Events []CollaborationEvent `json:"events"`
	Path   string               `json:"path"`
	mu     sync.RWMutex
}

func NewCollaborationMemory(path string) *CollaborationMemory {
	cm := &CollaborationMemory{
		Events: []CollaborationEvent{},
		Path:   path,
	}
	cm.Load()
	return cm
}

func (cm *CollaborationMemory) Load() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	data, err := os.ReadFile(cm.Path)
	if err == nil {
		json.Unmarshal(data, &cm.Events)
	}
}

func (cm *CollaborationMemory) Save() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	data, _ := json.Marshal(cm.Events)
	os.WriteFile(cm.Path, data, 0644)
}

func (cm *CollaborationMemory) RecordEvent(eventType string, data map[string]interface{}) {
	event := CollaborationEvent{
		ID:        fmt.Sprintf("ev_%d", time.Now().UnixNano()),
		Type:      eventType,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      data,
	}
	cm.mu.Lock()
	cm.Events = append(cm.Events, event)
	// Keep last 1000 events
	if len(cm.Events) > 1000 {
		cm.Events = cm.Events[len(cm.Events)-1000:]
	}
	cm.mu.Unlock()
	cm.Save()
}

func (cm *CollaborationMemory) GetTimeline(limit int) []CollaborationEvent {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if len(cm.Events) < limit {
		limit = len(cm.Events)
	}
	res := make([]CollaborationEvent, limit)
	for i := 0; i < limit; i++ {
		res[i] = cm.Events[len(cm.Events)-1-i]
	}
	return res
}

func (cm *CollaborationMemory) GetTimelineByType(eventType string, limit int) []CollaborationEvent {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	var res []CollaborationEvent
	for i := len(cm.Events) - 1; i >= 0 && len(res) < limit; i-- {
		if cm.Events[i].Type == eventType {
			res = append(res, cm.Events[i])
		}
	}
	return res
}

func (cm *CollaborationMemory) GetContextDigest() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Get last 5 decisions
	var decisions []string
	decCount := 0
	// Get recent advisor notes
	var notes []string
	noteCount := 0
	// Get recent alerts
	var alerts []string
	alertCount := 0

	for i := len(cm.Events) - 1; i >= 0; i-- {
		e := cm.Events[i]
		switch e.Type {
		case "decision":
			if decCount < 5 {
				dec, _ := json.Marshal(e.Data)
				decisions = append(decisions, string(dec))
				decCount++
			}
		case "advisor_note":
			if noteCount < 3 {
				note, _ := json.Marshal(e.Data)
				notes = append(notes, string(note))
				noteCount++
			}
		case "alert":
			if alertCount < 3 {
				alert, _ := json.Marshal(e.Data)
				alerts = append(alerts, string(alert))
				alertCount++
			}
		}
	}

	return fmt.Sprintf("Recent Decisions: %v\nRecent Advisor Notes: %v\nRecent Alerts: %v", decisions, notes, alerts)
}
