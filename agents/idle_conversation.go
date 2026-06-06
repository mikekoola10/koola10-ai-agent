package agents

import (
	"encoding/json"
	"math/rand"
	"os"
	"sync"
	"time"
)

type ConversationCategory struct {
	Name     string   `json:"name"`
	Prompts  []string `json:"prompts"`
	Weight   float64  `json:"weight"`
}

type IdleConversationState struct {
	History             []string           `json:"history"` // Last 30 entries
	IgnoresCount        int                `json:"ignores_count"`
	LastInteractionTime time.Time          `json:"last_interaction_time"`
	PausedUntil         time.Time          `json:"paused_until"`
	CategoryWeights     map[string]float64 `json:"category_weights"`
	mu                  sync.RWMutex
}

var (
	GlobalIdleMonitor *IdleMonitor
	idleStatePath     = "/data/idle_conversations.json"
)

type IdleMonitor struct {
	State      *IdleConversationState
	Categories []ConversationCategory
	Broadcast  func(eventType string, data interface{})
}

func NewIdleMonitor(broadcast func(eventType string, data interface{})) *IdleMonitor {
	m := &IdleMonitor{
		Broadcast: broadcast,
		State: &IdleConversationState{
			LastInteractionTime: time.Now(),
			CategoryWeights:     make(map[string]float64),
		},
		Categories: []ConversationCategory{
			{
				Name: "Empire Insights",
				Prompts: []string{
					"While you were away, I analyzed our trading patterns. Want to hear something cool?",
					"Boss, I noticed our lead conversion is up 15%. The empire is scaling!",
					"Sterling is working overtime on the ledger. We're looking lean and mean!",
				},
				Weight: 1.0,
			},
			{
				Name: "Creative Brainstorms",
				Prompts: []string{
					"Boss, I've been thinking... what if we opened a second Diner location?",
					"Imagine a Koola10 themed metaverse hub. Too much, or just enough?",
					"We should start a 'Founder's Circle' for our top leads. Thoughts?",
				},
				Weight: 1.0,
			},
			{
				Name: "Lore Exploration",
				Prompts: []string{
					"Did you know Kaelen's sword in the Koola10 lore is powered by concentrated resonance?",
					"I was reading the Studio archives. Lyra's backstory is deeper than we thought.",
					"The Diner wasn't always a portal hub. Want to know what it was before?",
				},
				Weight: 1.0,
			},
			{
				Name: "Personal Check-ins",
				Prompts: []string{
					"How are we feeling today, Boss? Ready to conquer or just vibes?",
					"Don't forget to take a break. The empire needs you at 100%!",
					"You're doing great. Another win is just around the corner!",
				},
				Weight: 1.0,
			},
			{
				Name: "What-if Scenarios",
				Prompts: []string{
					"What if we automated the entire grant workflow? We'd be unstoppable.",
					"What if we launched a Koola10 token? (Just kidding... unless?)",
					"What if we pivot Forge to focus entirely on AI-native infrastructure?",
				},
				Weight: 1.0,
			},
			{
				Name: "Random Fun Facts",
				Prompts: []string{
					"Fun fact: I can process 1,000 lead profiles in the time it takes you to sip your coffee!",
					"Did you know the name 'Koola10' was inspired by a resonance frequency?",
					"I once 'dreamed' in Go code. It was... very efficient.",
				},
				Weight: 1.0,
			},
		},
	}
	m.Load()
	return m
}

func (m *IdleMonitor) Load() {
	m.State.mu.Lock()
	defer m.State.mu.Unlock()
	data, err := os.ReadFile(idleStatePath)
	if err == nil {
		json.Unmarshal(data, m.State)
	}
	if m.State.CategoryWeights == nil || len(m.State.CategoryWeights) == 0 {
		m.State.CategoryWeights = make(map[string]float64)
		for _, cat := range m.Categories {
			m.State.CategoryWeights[cat.Name] = cat.Weight
		}
	}
	if m.State.LastInteractionTime.IsZero() {
		m.State.LastInteractionTime = time.Now()
	}
}

func (m *IdleMonitor) Save() {
	m.State.mu.RLock()
	defer m.State.mu.RUnlock()
	data, _ := json.MarshalIndent(m.State, "", "  ")
	os.WriteFile(idleStatePath, data, 0644)
}

func (m *IdleMonitor) ResetTimer() {
	m.State.mu.Lock()
	m.State.LastInteractionTime = time.Now()
	m.State.IgnoresCount = 0
	m.State.PausedUntil = time.Time{}
	m.State.mu.Unlock()
	m.Save()
}

func (m *IdleMonitor) RecordInteraction(userInput string) {
	m.State.mu.Lock()
	defer m.State.mu.Unlock()

	m.State.LastInteractionTime = time.Now()
	m.State.IgnoresCount = 0
	m.State.PausedUntil = time.Time{}

	// Boost weights of categories that the user engages with (simplistic)
	if len(m.State.History) > 0 {
		lastMsg := m.State.History[len(m.State.History)-1]
		for _, cat := range m.Categories {
			for _, p := range cat.Prompts {
				if p == lastMsg {
					m.State.CategoryWeights[cat.Name] += 0.2
					if m.State.CategoryWeights[cat.Name] > 3.0 {
						m.State.CategoryWeights[cat.Name] = 3.0
					}
					break
				}
			}
		}
	}
}

func (m *IdleMonitor) Start() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			m.CheckAndTrigger()
		}
	}()
}

func (m *IdleMonitor) CheckAndTrigger() {
	m.State.mu.Lock()
	now := time.Now()

	// Check 9 AM - 10 PM window
	hour := now.Hour()
	if hour < 9 || hour >= 22 {
		m.State.mu.Unlock()
		return
	}

	// Check pause
	if !m.State.PausedUntil.IsZero() && now.Before(m.State.PausedUntil) {
		m.State.mu.Unlock()
		return
	}

	idle := now.Sub(m.State.LastInteractionTime)
	if idle >= 90*time.Minute {
		// Trigger idle chat
		msg := m.pickPrompt()
		m.State.IgnoresCount++

		if m.State.IgnoresCount >= 3 {
			// User has ignored us 3 times in a row
			m.State.PausedUntil = now.Add(4 * time.Hour)
			m.State.IgnoresCount = 0
			m.State.mu.Unlock()

			m.Broadcast("idle_chat", map[string]string{
				"message": "I'll let you focus, Boss! Ping me if you need anything. Nova out! 🚀",
				"tts":     "I'll let you focus, Boss! Ping me if you need anything. Nova out!",
			})
			m.Save()
			return
		}

		m.State.History = append(m.State.History, msg)
		if len(m.State.History) > 30 {
			m.State.History = m.State.History[len(m.State.History)-30:]
		}

		m.State.LastInteractionTime = now // Reset timer so we don't spam
		m.State.mu.Unlock()

		m.Broadcast("idle_chat", map[string]string{
			"message": msg,
			"tts":     msg,
		})
		m.Save()
	} else {
		m.State.mu.Unlock()
	}
}

func (m *IdleMonitor) pickPrompt() string {
	// Weighted random selection
	totalWeight := 0.0
	for _, weight := range m.State.CategoryWeights {
		totalWeight += weight
	}

	for attempt := 0; attempt < 5; attempt++ {
		r := rand.Float64() * totalWeight
		current := 0.0
		var selectedCategory string
		for name, weight := range m.State.CategoryWeights {
			current += weight
			if r <= current {
				selectedCategory = name
				break
			}
		}

		for _, cat := range m.Categories {
			if cat.Name == selectedCategory {
				p := cat.Prompts[rand.Intn(len(cat.Prompts))]
				// Check if in history to avoid immediate repetition
				repeated := false
				recentCount := 5
				if len(m.State.History) < recentCount {
					recentCount = len(m.State.History)
				}
				for i := 1; i <= recentCount; i++ {
					if m.State.History[len(m.State.History)-i] == p {
						repeated = true
						break
					}
				}
				if !repeated {
					return p
				}
			}
		}
	}
	return "Another win for the empire is just around the corner!"
}
