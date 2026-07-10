package agents

import (
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
}

type SwarmManager struct {
	Swarms    map[string][]SpecialistAgent
	Divisions map[string][]string // DivisionName -> []Verticals
	Mu        sync.RWMutex

	// Swarm Memory
	Memory map[string]interface{}

	// Callbacks for logging to economic ledger and compliance audit
	AuditLogger   func(action string, details map[string]interface{})
	LedgerLogger  func(vertical, category string, amount float64, description string)
	RevenueLogger func(amount float64, source string)

	// Factory for creating agents for a vertical
	Factories map[string]func() []SpecialistAgent

	BasePrompt string
}

func NewSwarmManager() *SwarmManager {
	sm := &SwarmManager{
		Swarms:    make(map[string][]SpecialistAgent),
		Divisions: make(map[string][]string),
		Memory:    make(map[string]interface{}),
		Factories: make(map[string]func() []SpecialistAgent),
	}
	sm.LoadMemory()

	// Initialize default divisions
	sm.Divisions["apex"] = []string{"management", "strategy"}
	sm.Divisions["spiral"] = []string{"content", "design", "creative"}
	sm.Divisions["koola10"] = []string{"marketing", "gamification", "growth"}

	return sm
}

func (sm *SwarmManager) LoadMemory() {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	data, err := os.ReadFile("/data/swarm_memory.json")
	if err == nil {
		json.Unmarshal(data, &sm.Memory)
	}
}

func (sm *SwarmManager) SaveMemory() {
	sm.Mu.RLock()
	defer sm.Mu.RUnlock()
	data, _ := json.MarshalIndent(sm.Memory, "", "  ")
	os.MkdirAll("/data", 0755)
	os.WriteFile("/data/swarm_memory.json", data, 0644)
}

func (sm *SwarmManager) UpdateMemory(key string, value interface{}) {
	sm.Mu.Lock()
	sm.Memory[key] = value
	sm.Mu.Unlock()
	sm.SaveMemory()
}

func (sm *SwarmManager) SetGlobalPrompt(prompt string) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	sm.BasePrompt = prompt
	for _, swarm := range sm.Swarms {
		for _, agent := range swarm {
			agent.SetPrompt(prompt)
		}
	}
}

func (sm *SwarmManager) DeploySwarms(vertical string, count int) error {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	factory, ok := sm.Factories[vertical]
	if !ok {
		return fmt.Errorf("no factory for vertical: %s", vertical)
	}

	agents := factory()
	for _, agent := range agents {
		agent.SetPrompt(sm.BasePrompt)
	}
	sm.Swarms[vertical] = agents

	if sm.AuditLogger != nil {
		sm.AuditLogger("swarm_deployed", map[string]interface{}{
			"vertical": vertical,
			"count":    len(agents),
		})
	}

	return nil
}

var SwarmTaskCounter func()

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

	if err == nil && SwarmTaskCounter != nil { SwarmTaskCounter() }
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

// QuantumParallelDispatch executes a task across N agents simultaneously and synthesizes the results.
func (sm *SwarmManager) QuantumParallelDispatch(vertical string, task string, instances int) ([]interface{}, error) {
	sm.Mu.RLock()
	agents, ok := sm.Swarms[vertical]
	sm.Mu.RUnlock()

	if !ok || len(agents) < instances {
		return nil, fmt.Errorf("insufficient agents in %s for quantum parallel execution", vertical)
	}

	var wg sync.WaitGroup
	results := make([]interface{}, instances)
	errors := make([]error, instances)

	for i := 0; i < instances; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// For simplicity, we just dispatch to the i-th agent
			// In a real scenario, we'd find idle ones.
			res, err := agents[idx].Run(task)
			results[idx] = res
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	// Check if all failed
	allFailed := true
	for _, err := range errors {
		if err == nil {
			allFailed = false
			break
		}
	}

	if allFailed && instances > 0 {
		return nil, fmt.Errorf("all parallel instances failed: %v", errors[0])
	}

	return results, nil
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
