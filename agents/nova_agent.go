package agents

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type Goal struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Progress    int       `json:"progress"` // 0-100
	Active      bool      `json:"active"`
}

type NovaState struct {
	Goals               []Goal            `json:"goals"`
	ActiveProjects      []string          `json:"active_projects"`
	UserPreferences     map[string]int    `json:"user_preferences"` // Style -> Score
	LastProactiveCheck  time.Time         `json:"last_proactive_check"`
	LastCreation        time.Time         `json:"last_creation"`
	LastUserInteraction time.Time         `json:"last_user_interaction"`
	mu                  sync.RWMutex
	path                string
}

func NewNovaAgent(statePath string) *NovaAgent {
	state := &NovaState{
		Goals:           []Goal{},
		ActiveProjects:  []string{},
		UserPreferences: make(map[string]int),
		path:            statePath,
	}
	state.Load()
	return &NovaAgent{
		State: state,
	}
}

type NovaAgent struct {
	State *NovaState
}

func (s *NovaState) Load() {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err == nil {
		json.Unmarshal(data, s)
	}
	if s.UserPreferences == nil {
		s.UserPreferences = make(map[string]int)
	}
}

func (s *NovaState) Save() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, _ := json.Marshal(s)
	os.WriteFile(s.path, data, 0644)
}

func (a *NovaAgent) AddGoal(description string, deadline time.Time) string {
	a.State.mu.Lock()
	id := fmt.Sprintf("%x", time.Now().UnixNano())
	goal := Goal{
		ID:          id,
		Description: description,
		Deadline:    deadline,
		Progress:    0,
		Active:      true,
	}
	a.State.Goals = append(a.State.Goals, goal)
	a.State.mu.Unlock()
	a.State.Save()
	return id
}

func (a *NovaAgent) UpdateGoalProgress(goalID string, progress int) error {
	a.State.mu.Lock()
	defer a.State.mu.Unlock()
	for i, g := range a.State.Goals {
		if g.ID == goalID {
			a.State.Goals[i].Progress = progress
			if progress >= 100 {
				a.State.Goals[i].Active = false
			}
			a.State.mu.Unlock()
			a.State.Save()
			a.State.mu.Lock()
			return nil
		}
	}
	return fmt.Errorf("goal not found")
}

func (a *NovaAgent) ListActiveGoals() []Goal {
	a.State.mu.RLock()
	defer a.State.mu.RUnlock()
	var active []Goal
	for _, g := range a.State.Goals {
		if g.Active {
			active = append(active, g)
		}
	}
	return active
}

func (a *NovaAgent) StartProactiveLoop(broadcast func(event string, data interface{})) {
	ticker := time.NewTicker(3 * time.Hour)
	go func() {
		for range ticker.C {
			a.CheckAndGenerate(broadcast)
		}
	}()
}

func (a *NovaAgent) CheckAndGenerate(broadcast func(event string, data interface{})) {
	a.State.mu.Lock()
	a.State.LastProactiveCheck = time.Now()
	inactiveDuration := time.Since(a.State.LastUserInteraction)
	a.State.mu.Unlock()
	a.State.Save()

	if inactiveDuration >= 2*time.Hour {
		log.Println("[Nova] Inactive for 2+ hours, generating creative content...")
		content := a.GenerateCreativeContent()
		if content != nil {
			a.State.mu.Lock()
			a.State.LastCreation = time.Now()
			a.State.mu.Unlock()
			a.State.Save()
			if broadcast != nil {
				broadcast("nova_creation", content)
			}
		}
	}
}

func (a *NovaAgent) GenerateCreativeContent() map[string]interface{} {
	// Simple simulation of creative generation
	styles := a.GetTopStyles()
	styleStr := "neon-soaked"
	if len(styles) > 0 {
		styleStr = styles[0]
	}

	prompt := fmt.Sprintf("A %s diner scene in the Koola10 universe", styleStr)
	log.Printf("[Nova] Proactively generating image with prompt: %s", prompt)

	// Since we can't easily call tools.RunTool here due to circular dependency or availability,
	// and this is meant to be autonomous, we'll return the intent which the main loop can handle
	// OR we assume we can call an internal generation function.

	return map[string]interface{}{
		"type":      "image",
		"prompt":    prompt,
		"style":     styleStr,
		"timestamp": time.Now().Format(time.RFC3339),
	}
}

func (a *NovaAgent) RecordInteraction(userInput string) {
	a.State.mu.Lock()
	a.State.LastUserInteraction = time.Now()
	// Simple learning: if prompt contains certain keywords, increase style preference
	keywords := map[string]string{
		"neon":  "neon-soaked",
		"dark":  "gritty",
		"space": "cosmic",
		"retro": "vaporwave",
	}
	for kw, style := range keywords {
		if containsIgnoreCase(userInput, kw) {
			a.State.UserPreferences[style]++
		}
	}
	a.State.mu.Unlock()
	a.State.Save()
}

func (a *NovaAgent) GetTopStyles() []string {
	a.State.mu.RLock()
	defer a.State.mu.RUnlock()
	type styleScore struct {
		style string
		score int
	}
	var scores []styleScore
	for s, sc := range a.State.UserPreferences {
		scores = append(scores, styleScore{s, sc})
	}
	// sort scores
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].score < scores[j].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	var top []string
	for i := 0; i < len(scores) && i < 3; i++ {
		top = append(top, scores[i].style)
	}
	return top
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
