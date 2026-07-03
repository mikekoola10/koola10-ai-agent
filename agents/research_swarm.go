package agents

type ResearchAgent struct {
	manager *SwarmManager
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *ResearchAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Research Report (" + a.specialty + ") (Prompt: " + a.prompt + "): " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *ResearchAgent) Status() AgentStatus { return a.status }
func (a *ResearchAgent) Specialty() string    { return a.specialty }


func (a *ResearchAgent) SetPrompt(p string)   { a.prompt = p }
func (a *ResearchAgent) GetPrompt() string    { return a.prompt }

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

func (a *ResearchAgent) SetManager(m *SwarmManager) { a.manager = m }

func (a *ResearchAgent) ConfidenceLevel() float64 { return 0.95 }
func (a *ResearchAgent) RequestClarification(ctx string) string { return "" }
