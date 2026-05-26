package agents

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type ConversationState struct {
	ActiveTopic                 string            `json:"active_topic"`
	LastInteractionTime         time.Time         `json:"last_interaction_time"`
	PendingItems                []string          `json:"pending_items"`
	PersonaContext              string            `json:"persona_context"`
	ConversationHistorySummary string            `json:"conversation_history_summary"`
	UserPreferencesLearned      map[string]string `json:"user_preferences_learned"`
}

type ConversationMemory struct {
	State ConversationState
	Path  string
	mu    sync.RWMutex
}

func NewConversationMemory(path string) *ConversationMemory {
	return &ConversationMemory{
		Path: path,
		State: ConversationState{
			UserPreferencesLearned: make(map[string]string),
		},
	}
}

func (cm *ConversationMemory) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &cm.State)
}

func (cm *ConversationMemory) Save() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	data, err := json.MarshalIndent(cm.State, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.Path, data, 0644)
}

func (cm *ConversationMemory) UpdateInteraction() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.State.LastInteractionTime = time.Now()
}
