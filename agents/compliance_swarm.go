package agents

type ComplianceAgent struct {
	manager *SwarmManager
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *ComplianceAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Compliance Audit (" + a.specialty + "): " + a.prompt + " | Task: " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *ComplianceAgent) Status() AgentStatus { return a.status }
func (a *ComplianceAgent) Specialty() string    { return a.specialty }
func (a *ComplianceAgent) SetPrompt(p string)   { a.prompt = p }
func (a *ComplianceAgent) GetPrompt() string    { return a.prompt }

func ComplianceFactory() []SpecialistAgent {
	specialties := []string{
		"GDPR Compliance Monitor", "GDPR Risk Assessment",
		"SOC2 Control Mapping", "SOC2 Evidence Collection",
		"HIPAA Privacy Rule Monitor", "HIPAA Security Audit",
		"FINRA Regulatory Scan", "FINRA Reporting automation",
		"Audit Report Generator (Internal)", "Audit Report Generator (External)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &ComplianceAgent{specialty: s, status: StatusIdle})
	}
	return agents
}

func (a *ComplianceAgent) SetManager(m *SwarmManager) { a.manager = m }
