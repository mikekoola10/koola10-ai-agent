package agents

import (
	"koola10/mirror"
)

type ContentAgent struct {
	specialty string
	status    AgentStatus
	mirror    *mirror.Mirror
}

func (a *ContentAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusIdle }()

	if a.mirror != nil {
		ctx := a.mirror.GetContext("content")
		_ = ctx.Tone // Use tone for generation
	}

	res := "Content Result (" + a.specialty + "): " + task

	if a.mirror != nil {
		a.mirror.RecordOutcome("content", map[string]interface{}{"task": task, "success": true})
	}

	a.status = StatusCompleted
	return res, nil
}

func (a *ContentAgent) Status() AgentStatus { return a.status }
func (a *ContentAgent) Specialty() string    { return a.specialty }

func ContentFactory(m *mirror.Mirror) func() []SpecialistAgent {
	return func() []SpecialistAgent {
	specialties := []string{
		"Post Generation (Twitter)", "Post Generation (LinkedIn)", "Post Generation (Instagram)",
		"Comment Engagement (Automated)", "Comment Engagement (Filtered)", "Comment Moderation",
		"Content Scheduling (Global)", "Content Scheduling (Targeted)",
		"Performance Analysis (Viral)", "Performance Analysis (Engagement)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
		for _, s := range specialties {
			agents = append(agents, &ContentAgent{specialty: s, status: StatusIdle, mirror: m})
		}
		return agents
	}
}
