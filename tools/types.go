package tools

type ToolResult struct {
	Success bool        `json:"success"`
	Output  string      `json:"output"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ToolFunc func(payload map[string]interface{}) ToolResult
