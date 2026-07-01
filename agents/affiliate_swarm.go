package agents

type AffiliateAgent struct {
	specialty string
	status    AgentStatus
}

func (a *AffiliateAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Affiliate Result (" + a.specialty + "): " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *AffiliateAgent) Status() AgentStatus { return a.status }
func (a *AffiliateAgent) Specialty() string    { return a.specialty }

func AffiliateFactory() []SpecialistAgent {
	specialties := []string{
		"Amazon Associate Hunter", "Clickbank Offer Finder", "SaaS Affiliate Scout",
		"Content Linker", "Affiliate Dashboard Monitor", "Keyword Traffic Analyst",
		"Review Article Writer", "Coupon Code Aggregator", "Conversion Optimizer", "Affiliate Compliance",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &AffiliateAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
