package agents

import (
	"sync"
	"time"
)

type SpecialistAgent interface {
	Run(task string) string
	Status() string
	Specialty() string
	GetRevenue() float64
}

type SwarmManager struct {
	Swarms map[string][]SpecialistAgent
	mu     sync.RWMutex
	AuditLogger func(action string, details map[string]interface{})
	LedgerLogger func(category string, amount float64, description string)
}

func NewSwarmManager(audit func(string, map[string]interface{}), ledger func(string, float64, string)) *SwarmManager {
	return &SwarmManager{
		Swarms:       make(map[string][]SpecialistAgent),
		AuditLogger:  audit,
		LedgerLogger: ledger,
	}
}

func (sm *SwarmManager) DeploySwarms(vertical string, count int, factory func(id int) SpecialistAgent) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i := 0; i < count; i++ {
		agent := factory(i)
		sm.Swarms[vertical] = append(sm.Swarms[vertical], agent)
	}

	sm.AuditLogger("swarm_deployed", map[string]interface{}{
		"vertical": vertical,
		"count":    count,
	})
}

func (sm *SwarmManager) DispatchTask(vertical string, task string) string {
	sm.mu.RLock()
	agents, ok := sm.Swarms[vertical]
	sm.mu.RUnlock()

	if !ok || len(agents) == 0 {
		return "No agents available for vertical: " + vertical
	}

	// Simple round-robin or first available (mocking async dispatch)
	// For this simulation, we'll just use the first agent
	agent := agents[time.Now().UnixNano()%int64(len(agents))]
	result := agent.Run(task)

	sm.AuditLogger("task_dispatched", map[string]interface{}{
		"vertical": vertical,
		"task":     task,
		"agent":    agent.Specialty(),
	})

	return result
}

func (sm *SwarmManager) GetAllRevenueMetrics() map[string]float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	metrics := make(map[string]float64)
	for vertical, agents := range sm.Swarms {
		var total float64
		for _, agent := range agents {
			total += agent.GetRevenue()
		}
		metrics[vertical] = total
	}
	return metrics
}

func (sm *SwarmManager) GetStatus() map[string][]map[string]string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	status := make(map[string][]map[string]string)
	for vertical, agents := range sm.Swarms {
		for _, agent := range agents {
			status[vertical] = append(status[vertical], map[string]string{
				"specialty": agent.Specialty(),
				"status":    agent.Status(),
			})
		}
	}
	return status
}
