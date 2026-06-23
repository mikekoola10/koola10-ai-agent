package agents

import (
	"fmt"
	"koola10/beat-smith"
)

type BeatSmithAgent struct {
	specialty string
	status    AgentStatus
}

func (a *BeatSmithAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	switch a.specialty {
	case "VST Manager":
		vm := &beatsmith.VSTManager{}
		vm.ScanPlugins()
		return vm.InstalledPlugins, nil
	case "Loop Generator":
		lg := &beatsmith.LoopGenerator{LoopsDir: "/data/beat-smith/loops"}
		return lg.GenerateMIDIPattern("Trap", 140)
	case "Arrangement Advisor":
		aa := &beatsmith.ArrangementAdvisor{}
		return aa.GenerateReport("current_project.flp")
	case "Mastering Assistant":
		ma := &beatsmith.MasteringAssistant{}
		return ma.RecommendSettings("Trap"), nil
	}

	return fmt.Sprintf("BeatSmith Processed (%s): %s", a.specialty, task), nil
}

func (a *BeatSmithAgent) Status() AgentStatus { return a.status }
func (a *BeatSmithAgent) Specialty() string    { return a.specialty }

func BeatSmithFactory() []SpecialistAgent {
	specialties := []string{
		"VST Manager", "Loop Generator", "Arrangement Advisor",
		"Sample Manager", "Mastering Assistant",
		"Marketplace Sync", "Creative Strategist", "Sound Designer", "Rhythm Analyst",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &BeatSmithAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
