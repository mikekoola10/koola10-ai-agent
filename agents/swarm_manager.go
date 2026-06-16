package agents

import (
	"fmt"
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

	// AGI Capabilities
	Reason(input string) (string, error)
	Plan(goal string) ([]string, error)
	Learn(experience string) error
	Adapt(environment string) error
}

type SwarmManager struct {
	Swarms map[string][]SpecialistAgent
	Mu     sync.RWMutex

	// Callbacks for logging to economic ledger and compliance audit
	AuditLogger func(action string, details map[string]interface{})
	LedgerLogger func(vertical, category string, amount float64, description string)

	// Factory for creating agents for a vertical
	Factories map[string]func() []SpecialistAgent
}

func NewSwarmManager() *SwarmManager {
	return &SwarmManager{
		Swarms:    make(map[string][]SpecialistAgent),
		Factories: make(map[string]func() []SpecialistAgent),
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
	// If count is different from what factory produces, we might need to adjust,
	// but for now we assume factory produces the right set or we scale it.
	// The requirement says 10 agents for each.
	sm.Swarms[vertical] = agents

	if sm.AuditLogger != nil {
		sm.AuditLogger("swarm_deployed", map[string]interface{}{
			"vertical": vertical,
			"count":    len(agents),
		})
	}

	return nil
}

func (sm *SwarmManager) DispatchTask(vertical string, task string) (interface{}, error) {
	sm.Mu.Lock()
	agents, ok := sm.Swarms[vertical]

	if !ok || len(agents) == 0 {
		sm.Mu.Unlock()
		return nil, fmt.Errorf("no swarm deployed for vertical: %s", vertical)
	}

	// Simple dispatch logic: find the first available agent
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

	if sm.AuditLogger != nil {
		sm.AuditLogger("task_executed", map[string]interface{}{
			"vertical":  vertical,
			"specialty": target.Specialty(),
			"task":      task,
			"success":   err == nil,
		})
	}

	if sm.LedgerLogger != nil {
		// Log a nominal cost for agent execution
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

// BaseAGISkills provides default implementations for AGI capabilities
type BaseAGISkills struct {
	StatusVal AgentStatus
}

func (b *BaseAGISkills) Reason(input string) (string, error) {
	return "Reasoning: Analyzing input via Chain-of-Thought...", nil
}

func (b *BaseAGISkills) Plan(goal string) ([]string, error) {
	return []string{"Step 1: Research", "Step 2: Execute", "Step 3: Verify"}, nil
}

func (b *BaseAGISkills) Learn(experience string) error {
	return nil
}

func (b *BaseAGISkills) Adapt(environment string) error {
	return nil
}
