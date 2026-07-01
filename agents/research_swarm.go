package agents

type ResearchAgent struct {
	prompt    string
	specialty string
	status    AgentStatus
}

func (a *ResearchAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Market Intel (" + a.specialty + "): " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *ResearchAgent) Status() AgentStatus { return a.status }
func (a *ResearchAgent) Specialty() string    { return a.specialty }

func ResearchFactory() []SpecialistAgent {
	specialties := []string{
		"News Aggregator (Global)", "News Sentiment Tracker",
		"Patent Search (Tech)", "Patent Search (Biotech)",
		"Competitor Pricing Monitor", "Market Trend Analyzer",
		"Academic Paper Summarizer", "Research Citation Mapper",
		"Market Segment Deep-dive", "Intelligence Brief Generator",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &ResearchAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
func (a *ResearchAgent) SetPrompt(p string) { a.prompt = p }
