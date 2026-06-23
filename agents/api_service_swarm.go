package agents

type APIAgent struct {
	prompt    string
	specialty string
	status    AgentStatus
}

func (a *APIAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "API Result for " + a.specialty + ": " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *APIAgent) Status() AgentStatus { return a.status }
func (a *APIAgent) Specialty() string    { return a.specialty }

func APIFactory() []SpecialistAgent {
	specialties := []string{
		"Text-to-Image (Model A)", "Text-to-Image (Model B)",
		"Sentiment Analysis (Fast)", "Sentiment Analysis (Deep)",
		"Code Generation (Go)", "Code Generation (Python)",
		"Translation (EU Languages)", "Translation (Asian Languages)",
		"Summarization (Short)", "Summarization (Long)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &APIAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
func (a *APIAgent) SetPrompt(p string) { a.prompt = p }
