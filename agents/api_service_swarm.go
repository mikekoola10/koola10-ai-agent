package agents

import (
	"koola10/mirror"
)

type APIAgent struct {
	specialty string
	status    AgentStatus
	mirror    *mirror.Mirror
}

func (a *APIAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusIdle }()

	if a.mirror != nil {
		ctx := a.mirror.GetContext("echo")
		_ = ctx.Tone
	}

	res := "API Result for " + a.specialty + ": " + task

	if a.mirror != nil {
		a.mirror.RecordOutcome("echo", map[string]interface{}{"task": task, "success": true})
	}

	a.status = StatusCompleted
	return res, nil
}

func (a *APIAgent) Status() AgentStatus { return a.status }
func (a *APIAgent) Specialty() string    { return a.specialty }

func APIFactory(m *mirror.Mirror) func() []SpecialistAgent {
	return func() []SpecialistAgent {
	specialties := []string{
		"Text-to-Image (Model A)", "Text-to-Image (Model B)",
		"Sentiment Analysis (Fast)", "Sentiment Analysis (Deep)",
		"Code Generation (Go)", "Code Generation (Python)",
		"Translation (EU Languages)", "Translation (Asian Languages)",
		"Summarization (Short)", "Summarization (Long)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
		for _, s := range specialties {
			agents = append(agents, &APIAgent{specialty: s, status: StatusIdle, mirror: m})
		}
		return agents
	}
}
