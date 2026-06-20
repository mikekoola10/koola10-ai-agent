package mirror

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type ValuePoint struct {
	Value     string    `json:"value"`
	Weight    float64   `json:"weight"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserProfile struct {
	Values      map[string]ValuePoint `json:"values"`
	Habits      map[string]float64    `json:"habits"`
	Preferences map[string]string     `json:"preferences"`
	LastUpdated time.Time             `json:"last_updated"`
}

type Mirror struct {
	Profile    UserProfile `json:"profile"`
	Path       string      `json:"path"`
	mu         sync.RWMutex
}

func NewMirror(path string) *Mirror {
	m := &Mirror{
		Path: path,
		Profile: UserProfile{
			Values:      make(map[string]ValuePoint),
			Habits:      make(map[string]float64),
			Preferences: make(map[string]string),
		},
	}
	m.Load()
	return m
}

func (m *Mirror) Load() {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := os.ReadFile(m.Path)
	if err == nil {
		json.Unmarshal(data, &m.Profile)
	}
	if m.Profile.Values == nil {
		m.Profile.Values = make(map[string]ValuePoint)
	}
	if m.Profile.Habits == nil {
		m.Profile.Habits = make(map[string]float64)
	}
	if m.Profile.Preferences == nil {
		m.Profile.Preferences = make(map[string]string)
	}
}

func (m *Mirror) Save() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, _ := json.MarshalIndent(m.Profile, "", "  ")
	os.WriteFile(m.Path, data, 0644)
}

func (m *Mirror) UpdateValue(key string, value string, weight float64) {
	m.mu.Lock()
	m.Profile.Values[key] = ValuePoint{
		Value:     value,
		Weight:    weight,
		UpdatedAt: time.Now(),
	}
	m.Profile.LastUpdated = time.Now()
	m.mu.Unlock()
	m.Save()
}

func (m *Mirror) RecordHabit(key string) {
	m.mu.Lock()
	m.Profile.Habits[key] += 0.1
	if m.Profile.Habits[key] > 1.0 {
		m.Profile.Habits[key] = 1.0
	}
	m.Profile.LastUpdated = time.Now()
	m.mu.Unlock()
	m.Save()
}

func (m *Mirror) PredictConfidence(action string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Logic to predict if the user would take this action
	// For now, a placeholder that checks habit/value alignment
	confidence := 0.5
	if h, ok := m.Profile.Habits[action]; ok {
		confidence = 0.5 + (h * 0.5)
	}
	return confidence
}
