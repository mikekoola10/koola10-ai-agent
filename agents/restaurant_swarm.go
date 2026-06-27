package agents

type RestaurantAgent struct {
	specialty string
	status    AgentStatus
}

func (a *RestaurantAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Restaurant Op (" + a.specialty + "): " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *RestaurantAgent) Status() AgentStatus { return a.status }
func (a *RestaurantAgent) Specialty() string    { return a.specialty }

func RestaurantSwarmFactory() []SpecialistAgent {
	specialties := []string{
		"Menu Optimization (Vale)",
		"Kitchen Coordination (Forge)",
		"Supply Management (Nova)",
		"Loyalty Program (Solara)",
		"Table Management (Echo)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &RestaurantAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
