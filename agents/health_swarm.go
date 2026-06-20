package agents

type HealthAgent struct {
	specialty string
	status    AgentStatus
}

func (a *HealthAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Health Swarm (" + a.specialty + "): " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *HealthAgent) Status() AgentStatus { return a.status }
func (a *HealthAgent) Specialty() string    { return a.specialty }

func HealthSwarmFactory() []SpecialistAgent {
	specialties := []string{
		"Vitals Monitoring", "Nutrition Analysis",
		"Sleep Optimization", "Exercise Routine",
		"Mental Health Check-in", "Biofeedback Analysis",
		"Supplement Management", "Longevity Research",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &HealthAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
