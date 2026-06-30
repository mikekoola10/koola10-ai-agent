package agents
import (
	"fmt"
	"math/rand"
)
type AffiliateAgent struct {
	specialty string
	status    AgentStatus
	vertical  string
}
func (a *AffiliateAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()
	profit := 5.0 + rand.Float64()*20.0
	return map[string]interface{}{
		"message": fmt.Sprintf("Affiliate Result (%s - %s): Generated %.2f profit", a.vertical, a.specialty, profit),
		"profit":  profit,
	}, nil
}
func (a *AffiliateAgent) Status() AgentStatus { return a.status }
func (a *AffiliateAgent) Specialty() string    { return a.specialty }
func AffiliateFactory() []SpecialistAgent { return CreateAffiliateSwarm("affiliate") }
func CreateAffiliateSwarm(vertical string) []SpecialistAgent {
	specialties := []string{"Amazon Associate Optimizer", "ClickBank Niche Hunter"}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &AffiliateAgent{specialty: s, status: StatusIdle, vertical: vertical})
	}
	return agents
}
