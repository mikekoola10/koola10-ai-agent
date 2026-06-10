package tools

import (
	"encoding/json"
	"io"
	"koola10/sterling"
	"net/http"
	"sync"
)

var (
	registry = make(map[string]ToolFunc)
	regMu    sync.RWMutex
)

func RegisterTool(name string, fn ToolFunc) {
	regMu.Lock()
	defer regMu.Unlock()
	registry[name] = fn
}

func init() {
	RegisterTool("create_virtual_card", func(args map[string]interface{}) ToolResult {
		memo, ok := args["memo"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "Missing memo"}
		}
		spendLimit, ok := args["spend_limit"].(float64)
		if !ok {
			return ToolResult{Success: false, Error: "Missing spend_limit"}
		}

		client := sterling.NewPrivacyClient()
		resp, err := client.CreateVirtualCard(memo, int(spendLimit))
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
		return ToolResult{Success: true, Data: resp}
	})
}

func RunTool(name string, payload map[string]interface{}) ToolResult {
	regMu.RLock()
	fn, ok := registry[name]
	regMu.RUnlock()

	if !ok {
		return ToolResult{Success: false, Error: "Tool not found"}
	}
	return fn(payload)
}

func HandleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	toolName := r.URL.Query().Get("tool_name")
	if toolName == "" {
		http.Error(w, "Missing tool_name parameter", http.StatusBadRequest)
		return
	}

	regMu.RLock()
	fn, ok := registry[toolName]
	regMu.RUnlock()

	if !ok {
		http.Error(w, "Tool not found", http.StatusNotFound)
		return
	}

	var payload map[string]interface{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	} else {
		payload = make(map[string]interface{})
	}

	result := fn(payload)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		return
	}
}
