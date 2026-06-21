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
	Values      []string         `json:"values"`
	Preferences []UserPreference `json:"preferences"`
	Habits      []UserHabit      `json:"habits"`
}

type Mirror struct {
	Profile     UserProfile `json:"profile"`
	StoragePath string      `json:"storage_path"`
	mu          sync.RWMutex
}

func NewMirror(path string) *Mirror {
	m := &Mirror{
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

func (m *Mirror) Reflect(input string) {
	// In a real implementation, this would use an LLM to extract patterns from text
	// For now, we provide the infrastructure to be called by the orchestrator
}
