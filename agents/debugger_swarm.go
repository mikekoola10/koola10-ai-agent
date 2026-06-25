package agents

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"koola10/tools"
)

type DebuggerAgent struct {
	specialty string
	status    AgentStatus
}

func (a *DebuggerAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	log.Printf("[DebuggerAgent] Running diagnostics: %s", task)
	// For manual dispatch, we just return a success message for now
	return fmt.Sprintf("Completed diagnostics: %s", task), nil
}

func (a *DebuggerAgent) Status() AgentStatus { return a.status }
func (a *DebuggerAgent) Specialty() string    { return a.specialty }

func DebuggerFactory() []SpecialistAgent {
	return []SpecialistAgent{
		&DebuggerAgent{specialty: "System Diagnostics", status: StatusIdle},
		&DebuggerAgent{specialty: "Network Watchdog", status: StatusIdle},
		&DebuggerAgent{specialty: "Autonomous Recovery", status: StatusIdle},
	}
}

func StartDebuggerLoop(sm *SwarmManager) {
	ticker := time.NewTicker(1 * time.Minute)
	log.Println("[DebuggerSwarm] Autonomous monitoring loop started")

	for range ticker.C {
		checkEndpoints()
	}
}

func checkEndpoints() {
	endpoints := []string{
		"http://localhost:8080/health",
		"http://localhost:8080/grants/monitor",
		"http://localhost:8080/economic/ledger/summary",
	}

	for _, ep := range endpoints {
		resp, err := http.Get(ep)
		if err != nil {
			log.Printf("[DebuggerSwarm] Health check failed for %s: %v", ep, err)
			alertOnFailure(fmt.Errorf("endpoint %s unreachable: %v", ep, err))
			if ep == "http://localhost:8080/health" {
				restartApp()
			}
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("[DebuggerSwarm] Endpoint %s returned status %d", ep, resp.StatusCode)
			alertOnFailure(fmt.Errorf("endpoint %s returned status %d", ep, resp.StatusCode))
		}
		resp.Body.Close()
	}
}

func alertOnFailure(err error) {
	tools.RunTool("agentmail", map[string]interface{}{
		"to":      "mikekoola10@agentmail.to",
		"subject": "🚨 Debugger Swarm Alert: System Failure Detected",
		"body":    fmt.Sprintf("The Debugger Swarm detected a system issue:\n\n%v\n\nAttempting autonomous recovery...", err),
	})
}

func restartApp() {
	log.Println("[DebuggerSwarm] Attempting autonomous restart via Fly.io...")
	// flyctl apps restart koola10
	cmd := exec.Command("flyctl", "apps", "restart", "koola10")
	err := cmd.Run()
	if err != nil {
		log.Printf("[DebuggerSwarm] Failed to restart app: %v", err)
		tools.RunTool("agentmail", map[string]interface{}{
			"to":      "mikekoola10@agentmail.to",
			"subject": "❌ Debugger Swarm: Restart Failed",
			"body":    fmt.Sprintf("The Debugger Swarm tried to restart the app but failed:\n\n%v\n\nPlease intervene manually.", err),
		})
		return
	}
	log.Println("[DebuggerSwarm] Restart command issued successfully")
}
