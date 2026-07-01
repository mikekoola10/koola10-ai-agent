package agents

import (
	"testing"
)

func TestNexusSwarm(t *testing.T) {
	sm := NewSwarmManager()
	sm.Factories["liaison"] = LiaisonFactory
	sm.Factories["harmony"] = HarmonyFactory
	sm.Factories["synapse"] = SynapseFactory

	err := sm.DeploySwarms("liaison", 10)
	if err != nil {
		t.Fatalf("DeploySwarms failed: %v", err)
	}

	status := sm.GetSwarmStatus("liaison")
	if len(status) != 10 {
		t.Errorf("expected 10 agents, got %d", len(status))
	}

	res, err := sm.DispatchTask("liaison", "negotiate partnership")
	if err != nil {
		t.Fatalf("DispatchTask failed: %v", err)
	}
	if res == nil {
		t.Error("expected result, got nil")
	}

	metrics := sm.GetAllSwarmMetrics()
	if _, ok := metrics["liaison"]; !ok {
		t.Error("expected metrics for liaison vertical")
	}
}
