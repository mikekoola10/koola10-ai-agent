package agents

import (
	"fmt"
	"koola10/mirror"
)

type HealthAgent struct {
	specialty string
	status    AgentStatus
	mirror    *mirror.Mirror
}

func (a *HealthAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusIdle }()

	// Proactive logic: consult the mirror
	if a.mirror != nil {
		if val, ok := a.mirror.GetPreference("health", "automation_level"); ok && val == "apex" {
			fmt.Printf("[APEX] Health Swarm acting proactively for specialty: %s\n", a.specialty)
			// Implementation would perform proactive refills or schedule adjustments here
		}
	}

	if a.mirror != nil {
		a.mirror.RecordOutcome("health", map[string]interface{}{"task": task, "success": true})
	}

	return fmt.Sprintf("Health task executed: %s", task), nil
}

func (a *HealthAgent) Status() AgentStatus { return a.status }
func (a *HealthAgent) Specialty() string    { return a.specialty }

func HealthFactory(m *mirror.Mirror) func() []SpecialistAgent {
	return func() []SpecialistAgent {
		specialties := []string{
			"Inventory Monitor", "Schedule Optimizer", "Supplement Refill",
			"Bio-Feedback Analysis", "Routine Adjustment",
		}
		agents := make([]SpecialistAgent, 0, len(specialties))
		for _, s := range specialties {
			agents = append(agents, &HealthAgent{specialty: s, status: StatusIdle, mirror: m})
		}
		return agents
	}
}
