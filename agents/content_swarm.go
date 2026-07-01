package agents

type ContentAgent struct {
	prompt    string
	specialty string
	status    AgentStatus
}

func (a *ContentAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Content Result (" + a.specialty + "): " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *ContentAgent) Status() AgentStatus { return a.status }
func (a *ContentAgent) Specialty() string    { return a.specialty }

func ContentFactory() []SpecialistAgent {
	specialties := []string{
		"Post Generation (Twitter)", "Post Generation (LinkedIn)", "Post Generation (Instagram)",
		"Comment Engagement (Automated)", "Comment Engagement (Filtered)", "Comment Moderation",
		"Content Scheduling (Global)", "Content Scheduling (Targeted)",
		"Performance Analysis (Viral)", "Performance Analysis (Engagement)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &ContentAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
func (a *ContentAgent) SetPrompt(p string) { a.prompt = p }
