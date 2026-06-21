package mirror

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type UserPreference struct {
	Category string  `json:"category"`
	Key      string  `json:"key"`
	Value    string  `json:"value"`
	Weight   float64 `json:"weight"` // 0.0 to 1.0
}

type UserHabit struct {
	Action    string    `json:"action"`
	Frequency string    `json:"frequency"` // "daily", "weekly"
	LastSeen  time.Time `json:"last_seen"`
}

type UserProfile struct {
	Values      []string               `json:"values"`
	Preferences []UserPreference       `json:"preferences"`
	Habits      []UserHabit            `json:"habits"`
	Outcomes    []map[string]interface{} `json:"outcomes"`
}

type DecisionContext struct {
	Tone           string  `json:"tone"`
	RiskTolerance  string  `json:"risk_tolerance"`
	EthicsBoundary string  `json:"ethics_boundary"`
	Weight         float64 `json:"weight"`
}

type Mirror struct {
	UserID      string      `json:"user_id"`
	Profile     UserProfile `json:"profile"`
	StoragePath string      `json:"storage_path"`
	mu          sync.RWMutex
}

func NewMirror(userID string, path string) *Mirror {
	m := &Mirror{
		UserID:      userID,
		StoragePath: path,
		Profile: UserProfile{
			Values:      []string{},
			Preferences: []UserPreference{},
			Habits:      []UserHabit{},
		},
	}
	m.Load()
	return m
}

func (m *Mirror) Load() {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := os.ReadFile(m.StoragePath)
	if err == nil {
		json.Unmarshal(data, &m.Profile)
	}
}

func (m *Mirror) Save() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, _ := json.MarshalIndent(m.Profile, "", "  ")
	os.WriteFile(m.StoragePath, data, 0644)
}

func (m *Mirror) UpdatePreference(category, key, value string, weight float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	found := false
	for i, p := range m.Profile.Preferences {
		if p.Category == category && p.Key == key {
			m.Profile.Preferences[i].Value = value
			m.Profile.Preferences[i].Weight = weight
			found = true
			break
		}
	}
	if !found {
		m.Profile.Preferences = append(m.Profile.Preferences, UserPreference{category, key, value, weight})
	}
	go m.Save()
}

func (m *Mirror) GetPreference(category, key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.Profile.Preferences {
		if p.Category == category && p.Key == key {
			return p.Value, true
		}
	}
	return "", false
}

func (m *Mirror) GetContext(category string) DecisionContext {
	m.mu.RLock()
	// Fetch memories to inform context
	memories, _ := m.Recall(category)
	m.mu.RUnlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx := DecisionContext{
		Tone:           "professional",
		RiskTolerance:  "moderate",
		EthicsBoundary: "strict",
		Weight:         1.0,
	}

	// Simple heuristic: if memories contain "aggressive", increase risk tolerance
	for _, mem := range memories {
		if category == "trading" && (contains(mem, "aggressive") || contains(mem, "high risk")) {
			ctx.RiskTolerance = "high"
		}
	}

	for _, p := range m.Profile.Preferences {
		if p.Category == category || p.Category == "general" {
			switch p.Key {
			case "tone":
				ctx.Tone = p.Value
			case "risk_tolerance":
				ctx.RiskTolerance = p.Value
			case "ethics_boundary":
				ctx.EthicsBoundary = p.Value
			}
		}
	}
	return ctx
}

func contains(s, substr string) bool {
	return (len(s) >= len(substr)) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)
}

func (m *Mirror) RecordOutcome(category string, outcome map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	outcome["category"] = category
	outcome["timestamp"] = time.Now().Format(time.RFC3339)
	m.Profile.Outcomes = append(m.Profile.Outcomes, outcome)

	// Persist outcome as a memory as well
	outcomeJSON, _ := json.Marshal(outcome)
	go m.Remember(string(outcomeJSON), "outcome_"+time.Now().Format("20060102150405"))
	go m.Save()
}

func (m *Mirror) Reflect(input string) {
	// In a real implementation, this would use an LLM to extract patterns from text
	// For now, we provide the infrastructure to be called by the orchestrator
}
