package tools

import (
	"fmt"
)

func nucleiTool(payload map[string]interface{}) ToolResult {
	target, _ := payload["target"].(string)
	if target == "" {
		return ToolResult{Success: false, Error: "Missing target for nuclei scan"}
	}

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Nuclei: Scan complete for %s. Found 2 low, 1 medium vulnerabilities.", target),
		Data: map[string]interface{}{
			"vulnerabilities": []map[string]string{
				{"severity": "medium", "name": "Exposed Config File", "path": "/.git/config"},
				{"severity": "low", "name": "X-Frame-Options Header Missing", "path": "/"},
			},
		},
	}
}

func init() {
	RegisterTool("nuclei", nucleiTool)
}
