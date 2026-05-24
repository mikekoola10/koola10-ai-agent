package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"koola10/tools"
	"net/http"
	"os"
)

type ActivationRequest struct {
	Connector         string `json:"connector"`
	Provider          string `json:"provider"`
	OAuthClientID     string `json:"oauth_client_id"`
	OAuthClientSecret string `json:"oauth_client_secret"`
}

func ActivateConnector(req ActivationRequest) error {
	// 1. Identity Provisioning
	fmt.Printf("[Orchestrator] Provisioning identity for %s via MachineAuth...\n", req.Connector)
	_, err := tools.GetMachineAuthToken(req.Connector + "-agent")
	if err != nil {
		return fmt.Errorf("identity provisioning failed: %v", err)
	}

	// 2. OAuth Registration
	fmt.Printf("[Orchestrator] Registering %s provider with Nexus Framework...\n", req.Provider)
	nexusURL := os.Getenv("NEXUS_URL")
	if nexusURL == "" {
		nexusURL = "https://koola10-nexus.fly.dev"
	}
	adminKey := os.Getenv("NEXUS_ADMIN_KEY")

	providerData := map[string]string{
		"provider":      req.Provider,
		"client_id":     req.OAuthClientID,
		"client_secret": req.OAuthClientSecret,
	}
	body, _ := json.Marshal(providerData)
	hReq, _ := http.NewRequest("POST", nexusURL+"/api/providers", bytes.NewBuffer(body))
	hReq.Header.Set("Authorization", "Bearer "+adminKey)
	hReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(hReq)
	if err != nil {
		fmt.Printf("[Orchestrator] Nexus registration call failed (expected if not deployed): %v\n", err)
	} else {
		defer resp.Body.Close()
	}

	// 3. Verification
	fmt.Printf("[Orchestrator] Running test query for %s...\n", req.Connector)
	res := tools.RunTool(req.Connector, map[string]interface{}{"action": "test"})
	if !res.Success {
		return fmt.Errorf("connector verification failed: %s", res.Error)
	}

	fmt.Printf("[Orchestrator] Activation of %s complete.\n", req.Connector)
	return nil
}
