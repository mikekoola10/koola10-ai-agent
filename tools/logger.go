package tools

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	TraceID   string                 `json:"trace_id"`
	Level     string                 `json:"level"`
	Vertical  string                 `json:"vertical,omitempty"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

var errorLogFile *os.File

func init() {
	var err error
	errorLogFile, err = os.OpenFile("spiral_errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open spiral_errors.log: %v", err)
	}
}

func LogStructured(level, traceID, vertical, message string, details map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		TraceID:   traceID,
		Level:     level,
		Vertical:  vertical,
		Message:   message,
		Details:   details,
	}

	data, _ := json.Marshal(entry)
	fmt.Println(string(data))

	if level == "ERROR" || level == "CRITICAL" {
		if errorLogFile != nil {
			errorLogFile.Write(append(data, '\n'))
		}

		if level == "CRITICAL" {
			RunTool("agentmail", map[string]interface{}{
				"to":      "mikekoola10@agentmail.to",
				"subject": fmt.Sprintf("CRITICAL ALERT: %s", message),
				"body":    fmt.Sprintf("TraceID: %s\nVertical: %s\nMessage: %s\nDetails: %v", traceID, vertical, message, details),
			})
		}
	}
}
