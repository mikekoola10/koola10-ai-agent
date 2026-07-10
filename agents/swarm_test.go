package agents

import (
	"testing"
)

func TestSwarmManager(t *testing.T) {
	sm := NewSwarmManager()
	sm.Factories["test"] = func() []SpecialistAgent {
		return []SpecialistAgent{&TradingAgent{specialty: "test", status: StatusIdle}}
	}

	err := sm.DeploySwarms("test", 1)
	if err != nil {
		t.Fatalf("DeploySwarms failed: %v", err)
	}

	status := sm.GetSwarmStatus("test")
	if len(status) != 1 {
		t.Errorf("expected 1 agent, got %d", len(status))
	}

	res, err := sm.DispatchTask("test", "buy BTC")
	if err != nil {
		t.Fatalf("DispatchTask failed: %v", err)
	}
	if res == nil {
		t.Error("expected result, got nil")
	}

	metrics := sm.GetAllSwarmMetrics()
	if _, ok := metrics["test"]; !ok {
		t.Error("expected metrics for test vertical")
	}
}
