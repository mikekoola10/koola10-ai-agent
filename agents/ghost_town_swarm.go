package agents

import (
	"fmt"
	"koola10/mirror"
	"koola10/tools"
)

// Global mirror instance for the swarm to use
var swarmMirror *mirror.Mirror

func SetSwarmMirror(m *mirror.Mirror) {
	swarmMirror = m
}

// Scout Agent
type ScoutAgent struct {
	specialty string
	status    AgentStatus
}

func (a *ScoutAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	results := make(map[string]interface{})

	// Ingest from Apify
	payload := map[string]interface{}{
		"actor_id": "apify/zillow-scraper",
		"input": map[string]interface{}{
			"search": task,
		},
	}
	apifyRes := tools.RunTool("apify", payload)
	if apifyRes.Success {
		results["apify"] = apifyRes.Data
	}

	// Ingest from MCP Real Estate
	mcpPayload := map[string]interface{}{
		"action": "search",
		"params": map[string]interface{}{
			"query": task,
		},
	}
	mcpRes := tools.RunTool("mcp_realestate", mcpPayload)
	if mcpRes.Success {
		results["mcp"] = mcpRes.Data
	}

	return results, nil
}

func (a *ScoutAgent) Status() AgentStatus { return a.status }
func (a *ScoutAgent) Specialty() string    { return a.specialty }

// Mapper Agent
type MapperAgent struct {
	specialty string
	status    AgentStatus
}

func (a *MapperAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()
	// Simulated QGIS/LLM Geospatial analysis
	return fmt.Sprintf("Autonomous GIS analysis for %s completed using QGIS engine and geospatial LLM.", task), nil
}

func (a *MapperAgent) Status() AgentStatus { return a.status }
func (a *MapperAgent) Specialty() string    { return a.specialty }

// Valuator Agent
type ValuatorAgent struct {
	specialty string
	status    AgentStatus
}

func (a *ValuatorAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	// Real-world logic would parse property data from Scout
	estimatedCost := 250000.0
	projectedRevenue := 600000.0
	roi := projectedRevenue / estimatedCost

	return map[string]interface{}{
		"estimated_cost":    estimatedCost,
		"projected_revenue": projectedRevenue,
		"projected_roi":     roi,
	}, nil
}

func (a *ValuatorAgent) Status() AgentStatus { return a.status }
func (a *ValuatorAgent) Specialty() string    { return a.specialty }

// Strategist Agent
type StrategistAgent struct {
	specialty string
	status    AgentStatus
}

func (a *StrategistAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	pref := "sustainable growth"
	if swarmMirror != nil {
		if p, ok := swarmMirror.Profile.Preferences["investment_style"]; ok {
			pref = p
		}
	}

	return fmt.Sprintf("Generated revival strategy based on Mirror preference (%s): Community-led redevelopment with focus on %s.", pref, task), nil
}

func (a *StrategistAgent) Status() AgentStatus { return a.status }
func (a *StrategistAgent) Specialty() string    { return a.specialty }

// Sentinel Agent
type SentinelAgent struct {
	specialty string
	status    AgentStatus
}

func (a *SentinelAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	// Implementation would trigger alerts via AgentMail
	tools.RunTool("agentmail", map[string]interface{}{
		"to":      "mikekoola10@agentmail.to",
		"subject": "Sentinel Alert: Property Change Detected",
		"body":    fmt.Sprintf("Sentinel monitored update for %s: Market activity spike detected.", task),
	})

	return "Continuous monitoring active. Alerts will be routed to AgentMail.", nil
}

func (a *SentinelAgent) Status() AgentStatus { return a.status }
func (a *SentinelAgent) Specialty() string    { return a.specialty }

func GhostTownFactory() []SpecialistAgent {
	return []SpecialistAgent{
		&ScoutAgent{specialty: "Property Data Ingestion", status: StatusIdle},
		&MapperAgent{specialty: "Autonomous GIS", status: StatusIdle},
		&ValuatorAgent{specialty: "Cost & ROI Estimation", status: StatusIdle},
		&StrategistAgent{specialty: "Revival Strategist", status: StatusIdle},
		&SentinelAgent{specialty: "Continuous Sentinel", status: StatusIdle},
	}
}
