package agents

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type AGIPattern struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	Skill       string    `json:"skill"` // "Reasoning", "Planning", etc.
	Data        string    `json:"data"`
	DiscoveredAt time.Time `json:"discovered_at"`
}

type AGIKnowledgeBase struct {
	Patterns []AGIPattern `json:"patterns"`
	Mu       sync.RWMutex
	path     string
}

func NewAGIKnowledgeBase(path string) *AGIKnowledgeBase {
	kb := &AGIKnowledgeBase{
		path:     path,
		Patterns: []AGIPattern{},
	}
	kb.Load()
	return kb
}

func (kb *AGIKnowledgeBase) Load() {
	kb.Mu.Lock()
	defer kb.Mu.Unlock()
	data, err := os.ReadFile(kb.path)
	if err == nil {
		json.Unmarshal(data, kb)
	}
}

func (kb *AGIKnowledgeBase) Save() {
	kb.Mu.RLock()
	defer kb.Mu.RUnlock()
	data, _ := json.MarshalIndent(kb, "", "  ")
	os.WriteFile(kb.path, data, 0644)
}

func (kb *AGIKnowledgeBase) AddPattern(skill, source, data string) {
	kb.Mu.Lock()
	defer kb.Mu.Unlock()
	kb.Patterns = append(kb.Patterns, AGIPattern{
		ID:           time.Now().Format("20060102150405"),
		Skill:        skill,
		Source:       source,
		Data:         data,
		DiscoveredAt: time.Now(),
	})
	kb.Save()
}
