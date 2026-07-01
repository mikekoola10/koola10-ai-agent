package agents

import (
	"time"
)

type ProactiveAgent struct {
	SwarmManager *SwarmManager
	BroadcastFunc func(eventType string, data map[string]interface{})
}

func NewProactiveAgent(sm *SwarmManager, broadcast func(string, map[string]interface{})) *ProactiveAgent {
	return &ProactiveAgent{
		SwarmManager: sm,
		BroadcastFunc: broadcast,
	}
}

func (pa *ProactiveAgent) AutoDispatchLoop() {
	ticker := time.NewTicker(30 * time.Minute)
	// Initial run
	pa.checkAndDispatch()

	for range ticker.C {
		pa.checkAndDispatch()
	}
}

func (pa *ProactiveAgent) checkAndDispatch() {
	metrics := pa.SwarmManager.GetAllSwarmMetrics()
	for vertical, m := range metrics {
		vMetrics, ok := m.(map[string]interface{})
		if !ok {
			continue
		}

		idleCount, _ := vMetrics["idle"].(int)
		if idleCount > 0 {
			// Auto-dispatch to all 5 ecosystems for this vertical
			ecosystems := []string{"koola10", "oracle", "sentinel", "nexus", "rebel"}
			for _, eco := range ecosystems {
				task := "Auto-dispatched maintenance task"
				pa.SwarmManager.DispatchTaskWithEcosystem(eco, vertical, task)
			}

			if pa.BroadcastFunc != nil {
				pa.BroadcastFunc("auto_dispatch", map[string]interface{}{
					"vertical":    vertical,
					"idle_agents": idleCount,
					"action":      "dispatched_to_all_ecosystems",
				})
			}
		}
	}
}
