package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func GetNexusToken(provider string) (string, error) {
	nexusURL := os.Getenv("NEXUS_URL")
	if nexusURL == "" {
		nexusURL = "https://koola10-nexus.fly.dev"
	}

	adminKey := os.Getenv("NEXUS_ADMIN_KEY")

	req, _ := http.NewRequest("GET", nexusURL+"/api/tokens/"+provider, nil)
	req.Header.Set("Authorization", "Bearer "+adminKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct { Token string `json:"token"` }
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.Token, nil
}
