package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func apifyTool(payload map[string]interface{}) ToolResult {
	token := os.Getenv("APIFY_API_TOKEN")
	if token == "" {
		return ToolResult{Success: false, Error: "APIFY_API_TOKEN not set"}
	}

	actorID, _ := payload["actor_id"].(string)
	input := payload["input"]

	if actorID == "" {
		return ToolResult{Success: false, Error: "Missing actor_id"}
	}

	url := fmt.Sprintf("https://api.apify.com/v2/acts/%s/runs?token=%s", actorID, token)
	body, _ := json.Marshal(input)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	return ToolResult{
		Success: true,
		Data:    res,
		Output:  "Apify actor run initiated",
	}
}

func init() {
	RegisterTool("apify", apifyTool)
}
