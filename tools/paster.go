package tools

import (
	"fmt"
	"os"
	"strings"
)

func init() {
	RegisterTool("code_paster", CodePaster)
}

func CodePaster(payload map[string]interface{}) ToolResult {
	filepath, _ := payload["filepath"].(string)
	search, _ := payload["search"].(string)
	replace, _ := payload["replace"].(string)

	if filepath == "" || search == "" || replace == "" {
		return ToolResult{Success: false, Error: "missing parameters (filepath, search, replace)"}
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to read file: %v", err)}
	}

	content := string(data)
	if !strings.Contains(content, search) {
		return ToolResult{Success: false, Error: "search string not found in file"}
	}

	newContent := strings.Replace(content, search, replace, 1)
	err = os.WriteFile(filepath, []byte(newContent), 0644)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to write file: %v", err)}
	}

	return ToolResult{
		Success: true,
		Data:    map[string]string{"status": "applied", "file": filepath},
	}
}
