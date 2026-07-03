package agents

import (
	"encoding/json"
	"fmt"
	"os"
)

type PersonaAgent struct {
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *PersonaAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()
	// Simulated execution
	return map[string]interface{}{
		"status": "success",
		"specialty": a.specialty,
		"result": fmt.Sprintf("Executed task '%s' using %s persona logic.", task, a.specialty),
	}, nil
}

func (a *PersonaAgent) Status() AgentStatus { return a.status }
func (a *PersonaAgent) Specialty() string    { return a.specialty }
func (a *PersonaAgent) SetPrompt(p string)   { a.prompt = p }
func (a *PersonaAgent) GetPrompt() string    { return a.prompt }

func PersonaFactory(vertical string) func() []SpecialistAgent {
	return func() []SpecialistAgent {
		data, err := os.ReadFile("agents/roles.json")
		if err != nil {
			return nil
		}
		var roles map[string][]string
		json.Unmarshal(data, &roles)

		roleList, ok := roles[vertical]
		if !ok {
			return nil
		}

		agents := make([]SpecialistAgent, 0, len(roleList))
		for _, r := range roleList {
			agents = append(agents, &PersonaAgent{specialty: r, status: StatusIdle})
		}
		return agents
	}
}
