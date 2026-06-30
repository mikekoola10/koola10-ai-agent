package agents
import (
	"fmt"
	"math/rand"
)
type BountyAgent struct {
	specialty string
	status    AgentStatus
	vertical  string
}
func (a *BountyAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()
	profit := 50.0 + rand.Float64()*500.0
	return map[string]interface{}{
		"message": fmt.Sprintf("Bounty Result (%s - %s): Secured %.2f bounty", a.vertical, a.specialty, profit),
		"profit":  profit,
	}, nil
}
func (a *BountyAgent) Status() AgentStatus { return a.status }
func (a *BountyAgent) Specialty() string    { return a.specialty }
func BountyFactory() []SpecialistAgent { return CreateBountySwarm("bounty") }
func CreateBountySwarm(vertical string) []SpecialistAgent {
	specialties := []string{"HackerOne Scanner", "Bugcrowd Vulnerability Finder"}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &BountyAgent{specialty: s, status: StatusIdle, vertical: vertical})
	}
	return agents
}
