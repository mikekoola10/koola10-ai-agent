package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"time"
)

var (
	sessions        = make(map[string]string) // sessionID -> machineID
	sessionMu       sync.RWMutex
	terminalLogPath = "/data/terminal_audit.jsonl"
	machineIDRegex  = regexp.MustCompile(`[0-9a-f]{14}`)
)

type TerminalEntry struct {
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	Command   string `json:"command"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	ExitCode  int    `json:"exit_code"`
	SessionID string `json:"session_id,omitempty"`
}

func logTerminalCommand(entry TerminalEntry) {
	entry.Timestamp = time.Now().Format(time.RFC3339)
	f, err := os.OpenFile(terminalLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(entry)
}

func terminalTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)

	switch action {
	case "exec":
		return execCommand(payload)
	case "interactive":
		return interactiveCommand(payload)
	case "status":
		return terminalStatus()
	default:
		return ToolResult{Success: false, Error: "Invalid terminal action"}
	}
}

func execCommand(payload map[string]interface{}) ToolResult {
	command, _ := payload["command"].(string)
	sessionID, _ := payload["session_id"].(string)
	timeout, _ := payload["timeout"].(float64)
	if timeout == 0 {
		timeout = 30
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if sessionID != "" {
		sessionMu.RLock()
		machineID, ok := sessions[sessionID]
		sessionMu.RUnlock()
		if !ok {
			return ToolResult{Success: false, Error: "Session not found"}
		}
		// Exec inside existing machine
		cmd = exec.CommandContext(ctx, "flyctl", "machine", "exec", machineID, "--", "sh", "-c", command)
	} else {
		// Run in one-off machine
		cmd = exec.CommandContext(ctx, "flyctl", "machine", "run", "--rm", "-i", "alpine:latest", "--", "sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	res := ToolResult{
		Success: err == nil,
		Output:  string(output),
		Data:    map[string]interface{}{"exit_code": exitCode},
	}
	if err != nil {
		res.Error = err.Error()
	}

	logTerminalCommand(TerminalEntry{
		Action:    "exec",
		Command:   command,
		Output:    string(output),
		Error:     res.Error,
		ExitCode:  exitCode,
		SessionID: sessionID,
	})

	return res
}

func interactiveCommand(payload map[string]interface{}) ToolResult {
	command, _ := payload["command"].(string)

	// Start a long-running machine in detached mode
	cmd := exec.Command("flyctl", "machine", "run", "-d", "alpine:latest", "--", "sleep", "3600")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to start session: %v (Output: %s)", err, string(output))}
	}

	// Extract Machine ID from output
	machineID := machineIDRegex.FindString(string(output))
	if machineID == "" {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to extract machine ID from output: %s", string(output))}
	}

	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	sessionMu.Lock()
	sessions[sessionID] = machineID
	sessionMu.Unlock()

	// If a command was provided initially, run it
	var initialOutput string
	if command != "" {
		execRes := execCommand(map[string]interface{}{
			"command":    command,
			"session_id": sessionID,
		})
		initialOutput = execRes.Output
	}

	logTerminalCommand(TerminalEntry{
		Action:    "interactive",
		Command:   command,
		Output:    initialOutput,
		SessionID: sessionID,
	})

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Started interactive session %s (Machine %s)", sessionID, machineID),
		Data:    map[string]interface{}{"session_id": sessionID, "machine_id": machineID, "initial_output": initialOutput},
	}
}

func terminalStatus() ToolResult {
	f, err := os.Open(terminalLogPath)
	if err != nil {
		return ToolResult{Success: true, Output: "No history found", Data: []TerminalEntry{}}
	}
	defer f.Close()

	var entries []TerminalEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry TerminalEntry
		if err := json.Unmarshal([]byte(scanner.Text()), &entry); err == nil {
			entries = append(entries, entry)
		}
	}

	start := len(entries) - 10
	if start < 0 {
		start = 0
	}
	recent := entries[start:]

	out, _ := json.MarshalIndent(recent, "", "  ")
	return ToolResult{
		Success: true,
		Output:  string(out),
		Data:    recent,
	}
}

func init() {
	RegisterTool("terminal", terminalTool)
}
