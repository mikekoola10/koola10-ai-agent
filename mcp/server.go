package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"koola10/tools"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var (
	resources = []Resource{
		{URI: "k10://economic/ledger", Name: "Economic Ledger", Description: "Real-time financial transactions"},
		{URI: "k10://swarm/status", Name: "Swarm Status", Description: "Live agent fleet status"},
	}
	resMu sync.RWMutex
)

func HandleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simple API key auth
	apiKey := os.Getenv("MCP_API_KEY")
	if apiKey != "" {
		providedKey := r.Header.Get("X-MCP-API-Key")
		if providedKey != apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	var req JSONRPCRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", 400)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON-RPC", 400)
		return
	}

	var result interface{}
	var rpcErr interface{}

	switch req.Method {
	case "tools/list":
		result = tools.ListRegisteredTools()
	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			result = tools.RunTool(params.Name, params.Arguments)
		} else {
			rpcErr = map[string]string{"code": "-32602", "message": "Invalid params"}
		}
	case "resources/list":
		resMu.RLock()
		result = resources
		resMu.RUnlock()
	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			result = readResource(params.URI)
		} else {
			rpcErr = map[string]string{"code": "-32602", "message": "Invalid params"}
		}
	default:
		rpcErr = map[string]string{"code": "-32601", "message": "Method not found"}
	}

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
		Error:   rpcErr,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func readResource(uri string) interface{} {
	switch uri {
	case "k10://economic/ledger":
		return map[string]interface{}{"status": "ready", "source": "economic_ledger.json"}
	case "k10://swarm/status":
		return map[string]interface{}{"status": "active", "agents": 5}
	default:
		return fmt.Errorf("resource not found: %s", uri)
	}
}
