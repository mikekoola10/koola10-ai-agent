package agents

import (
	"koola10/mirror"
)

type ResearchAgent struct {
	specialty string
	status    AgentStatus
	mirror    *mirror.Mirror
}

func (a *ResearchAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusIdle }()

	if a.mirror != nil {
		ctx := a.mirror.GetContext("vale")
		_ = ctx.RiskTolerance
	}

	res := "Market Intel (" + a.specialty + "): " + task

	if a.mirror != nil {
		a.mirror.RecordOutcome("vale", map[string]interface{}{"task": task, "success": true})
	}

	a.status = StatusCompleted
	return res, nil
}

func (a *ResearchAgent) Status() AgentStatus { return a.status }
func (a *ResearchAgent) Specialty() string    { return a.specialty }

func ResearchFactory(m *mirror.Mirror) func() []SpecialistAgent {
	return func() []SpecialistAgent {
	specialties := []string{
		"News Aggregator (Global)", "News Sentiment Tracker",
		"Patent Search (Tech)", "Patent Search (Biotech)",
		"Competitor Pricing Monitor", "Market Trend Analyzer",
		"Academic Paper Summarizer", "Research Citation Mapper",
		"Market Segment Deep-dive", "Intelligence Brief Generator",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
		for _, s := range specialties {
			agents = append(agents, &ResearchAgent{specialty: s, status: StatusIdle, mirror: m})
		}
		return agents
	}
}
