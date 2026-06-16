package agents

import (
	"sync"
	"time"
)

type BlackboardEntry struct {
	Source    string      `json:"source"`
	Content   interface{} `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
}

type Blackboard struct {
	Data map[string][]BlackboardEntry
	mu   sync.RWMutex
}

func NewBlackboard() *Blackboard {
	return &Blackboard{
		Data: make(map[string][]BlackboardEntry),
	}
}

// BroadcastFunc is a callback for broadcasting blackboard updates
var BroadcastFunc func(data map[string][]BlackboardEntry)

func (b *Blackboard) Post(key string, source string, content interface{}) {
	b.mu.Lock()
	b.Data[key] = append(b.Data[key], BlackboardEntry{
		Source:    source,
		Content:   content,
		Timestamp: time.Now(),
	})
	allData := make(map[string][]BlackboardEntry)
	for k, v := range b.Data {
		allData[k] = v
	}
	b.mu.Unlock()

	if BroadcastFunc != nil {
		BroadcastFunc(allData)
	}
}

func (b *Blackboard) Get(key string) []BlackboardEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Data[key]
}

func (b *Blackboard) GetAll() map[string][]BlackboardEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	// Create a copy to avoid race conditions on the map itself
	copy := make(map[string][]BlackboardEntry)
	for k, v := range b.Data {
		copy[k] = v
	}
	return copy
}
