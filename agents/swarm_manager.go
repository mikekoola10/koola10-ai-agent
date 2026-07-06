package agents

import (
	"time"
	"log"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type AgentStatus string

const (
	StatusIdle      AgentStatus = "idle"
	StatusWorking   AgentStatus = "working"
	StatusError     AgentStatus = "error"
	StatusCompleted AgentStatus = "completed"
)

type SpecialistAgent interface {
	Run(task string) (interface{}, error)
	Status() AgentStatus
	Specialty() string
	SetPrompt(prompt string)
	GetPrompt() string
	SetManager(m *SwarmManager)
	ConfidenceLevel() float64
	RequestClarification(context string) string
}

type TaskForce struct {
	ID        string   `json:"id"`
	Goal      string   `json:"goal"`
	Verticals []string `json:"verticals"`
	Status    string   `json:"status"`
	Insights  []string `json:"insights"`
}

type SwarmManager struct {
	Swarms map[string][]SpecialistAgent
	Mu     sync.RWMutex

	AuditLogger   func(action string, details map[string]interface{})
	LedgerLogger  func(vertical, category string, amount float64, description string)
	RevenueLogger func(amount float64, source string)
	ReflectLogger func(vertical, specialty, task string, result interface{}) string

	Factories map[string]func() []SpecialistAgent

	BasePrompt string
	AGIMode    bool

	LongTermMemory map[string]string
	MemoryPath     string

	TaskForces map[string]*TaskForce
}

func NewSwarmManager() *SwarmManager {
	sm := &SwarmManager{
		Swarms:         make(map[string][]SpecialistAgent),
		Factories:      make(map[string]func() []SpecialistAgent),
		LongTermMemory: make(map[string]string),
		MemoryPath:     "./data/agi_memory.json",
		TaskForces:     make(map[string]*TaskForce),
		AGIMode:        true,
	}
	sm.LoadMemory()
	return sm
}

func (sm *SwarmManager) LoadMemory() {
	data, err := os.ReadFile(sm.MemoryPath)
	if err == nil {
		json.Unmarshal(data, &sm.LongTermMemory)
	}
}

func (sm *SwarmManager) SaveMemory() {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	data, _ := json.MarshalIndent(sm.LongTermMemory, "", "  ")
	os.WriteFile(sm.MemoryPath, data, 0644)
}

func (sm *SwarmManager) SetMemory(key, value string) {
	sm.Mu.Lock()
	sm.LongTermMemory[key] = value
	sm.Mu.Unlock()
	sm.SaveMemory()
}

func (sm *SwarmManager) GetMemory(key string) string {
	sm.Mu.RLock()
	defer sm.Mu.RUnlock()
	return sm.LongTermMemory[key]
}

func (sm *SwarmManager) FormTaskForce(id, goal string, verticals []string) *TaskForce {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	tf := &TaskForce{
		ID:        id,
		Goal:      goal,
		Verticals: verticals,
		Status:    "active",
		Insights:  []string{},
	}
	sm.TaskForces[id] = tf
	if sm.AuditLogger != nil {
		sm.AuditLogger("task_force_formed", map[string]interface{}{
			"id":        id,
			"goal":      goal,
			"verticals": verticals,
		})
	}
	return tf
}

func (sm *SwarmManager) Coordinate(sourceVertical, targetVertical, task string) (interface{}, error) {
	if !sm.IsAGIMode() {
		return nil, fmt.Errorf("swarm coordination requires AGI Mode")
	}

	sm.Mu.RLock()
	if sm.AuditLogger != nil {
		sm.AuditLogger("swarm_coordination", map[string]interface{}{
			"source": sourceVertical,
			"target": targetVertical,
			"task":   task,
		})
	}
	sm.Mu.RUnlock()

	return sm.DispatchTask(targetVertical, task)
}

func (sm *SwarmManager) getEffectivePrompt() string {
	effectivePrompt := sm.BasePrompt
	if sm.AGIMode {
		agiDirectives := "\n\nAGI/ASI DIRECTIVES ACTIVE:\n" +
			"- Operate with General Intelligence: Handle any intellectual task at or beyond human level.\n" +
			"- Practice Recursive Self-Improvement: Analyze performance and suggest improvements.\n" +
			"- Use First-Principles + Antifragility: Break problems down and get stronger from failure.\n" +
			"- Enable Swarm Intelligence: Seamlessly collaborate (Apex coordinates, Spiral creates, Koola10 gamifies).\n" +
			"- Build Persistent Memory: Share insights across sessions.\n" +
			"- Default to 10x/100x Thinking: Seek leverage and exponential outcomes."
		effectivePrompt += agiDirectives
	}
	return effectivePrompt
}

func (sm *SwarmManager) SetGlobalPrompt(prompt string) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	sm.BasePrompt = prompt

	effectivePrompt := sm.getEffectivePrompt()

	for _, swarm := range sm.Swarms {
		for _, agent := range swarm {
			agent.SetPrompt(effectivePrompt)
		}
	}
}

func (sm *SwarmManager) ToggleAGIMode(enabled bool) {
	sm.Mu.Lock()
	sm.AGIMode = enabled
	sm.Mu.Unlock()
	sm.SetGlobalPrompt(sm.BasePrompt)
}

func (sm *SwarmManager) IsAGIMode() bool {
	sm.Mu.RLock()
	defer sm.Mu.RUnlock()
	return sm.AGIMode
}

func (sm *SwarmManager) DeploySwarms(vertical string, count int) error {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	factory, ok := sm.Factories[vertical]
	if !ok {
		return fmt.Errorf("no factory for vertical: %s", vertical)
	}

	// Quantum Parallelism: Scaling to specified count
	var agents []SpecialistAgent
	effectivePrompt := sm.getEffectivePrompt()

	for len(agents) < count {
		newAgents := factory()
		if len(newAgents) == 0 { break }
		for _, a := range newAgents {
			if len(agents) >= count { break }
			a.SetPrompt(effectivePrompt)
			a.SetManager(sm)
			agents = append(agents, a)
		}
	}

	sm.Swarms[vertical] = agents
	return nil
}

func (sm *SwarmManager) DispatchTask(vertical string, task string) (interface{}, error) {
	sm.Mu.Lock()
	agents, ok := sm.Swarms[vertical]

	if !ok || len(agents) == 0 {
		sm.Mu.Unlock()
		return nil, fmt.Errorf("no swarm deployed for vertical: %s", vertical)
	}

	var target SpecialistAgent
	for _, a := range agents {
		status := a.Status()
		if status == StatusIdle || status == StatusCompleted {
			target = a
			break
		}
	}

	if target == nil {
		sm.Mu.Unlock()
		return nil, fmt.Errorf("all agents in %s swarm are busy", vertical)
	}
	sm.Mu.Unlock()

	result, err := target.Run(task)

	if sm.AGIMode && sm.ReflectLogger != nil {
		sm.ReflectLogger(vertical, target.Specialty(), task, result)
	}

	if err == nil && sm.RevenueLogger != nil {
		if resMap, ok := result.(map[string]interface{}); ok {
			if profit, ok := resMap["profit"].(float64); ok {
				sm.RevenueLogger(profit, vertical)
			}
		}
	}

	if sm.AuditLogger != nil {
		sm.AuditLogger("task_executed", map[string]interface{}{
			"vertical":  vertical,
			"specialty": target.Specialty(),
			"task":      task,
			"success":   err == nil,
		})
	}

	if sm.LedgerLogger != nil {
		sm.LedgerLogger(vertical, "swarm_execution", 0.05, fmt.Sprintf("Executed task in %s: %s", vertical, target.Specialty()))
	}

	return result, err
}

func (sm *SwarmManager) GetSwarmStatus(vertical string) []map[string]string {
	sm.Mu.RLock()
	defer sm.Mu.RUnlock()
	agents := sm.Swarms[vertical]
	res := make([]map[string]string, 0, len(agents))
	for _, a := range agents {
		res = append(res, map[string]string{
			"specialty": a.Specialty(),
			"status":    string(a.Status()),
		})
	}
	return res
}

func (sm *SwarmManager) GetAllSwarmMetrics() map[string]interface{} {
	sm.Mu.RLock()
	defer sm.Mu.RUnlock()

	metrics := make(map[string]interface{})
	for vertical, agents := range sm.Swarms {
		idle := 0
		working := 0
		completed := 0
		errs := 0
		for _, a := range agents {
			switch a.Status() {
			case StatusIdle:
				idle++
			case StatusWorking:
				working++
			case StatusCompleted:
				completed++
			case StatusError:
				errs++
			}
		}
		metrics[vertical] = map[string]interface{}{
			"total":     len(agents),
			"idle":      idle,
			"working":   working,
			"completed": completed,
			"error":     errs,
		}
	}
	return metrics
}

func (sm *SwarmManager) HyperScale(vertical string, baseCount int) error {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	log.Printf("[Hyper-Scale] Initiating mass deployment for %s. Base count: %d", vertical, baseCount)

	// Phase 9: 10x Velocity Expansion
	targetCount := baseCount * 2
	if sm.AGIMode {
		targetCount = baseCount * 5 // 5x scaling in AGI Mode
	}

	factory, ok := sm.Factories[vertical]
	if !ok {
		return fmt.Errorf("no factory for vertical: %s", vertical)
	}

	currentAgents := sm.Swarms[vertical]
	effectivePrompt := sm.getEffectivePrompt()

	newAgents := factory()
	for _, a := range newAgents {
		if len(currentAgents) >= targetCount { break }
		a.SetPrompt(effectivePrompt)
		a.SetManager(sm)
		currentAgents = append(currentAgents, a)
	}

	sm.Swarms[vertical] = currentAgents
	log.Printf("[Hyper-Scale] %s vertical scaled to %d nodes.", vertical, len(currentAgents))
	return nil
}

func (sm *SwarmManager) SummarizeMemory() {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	if len(sm.LongTermMemory) < 1 { return }

	log.Printf("[AGI Memory] Triggering recursive memory summarization...")
	sm.LongTermMemory["consolidated_wisdom"] = fmt.Sprintf("Summarized at %s. AGI Swarm stable.", time.Now().Format(time.RFC3339))
	sm.SaveMemory()
}
