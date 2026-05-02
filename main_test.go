package main

import (
	"os"
	"testing"
	"time"
)

func TestGetDeepSeekBaseURL(t *testing.T) {
	t.Run("EnvVarOverride", func(t *testing.T) {
		// Reset cache for test
		baseURLMu.Lock()
		lastBaseURLCheck = time.Time{}
		cachedBaseURL = ""
		baseURLMu.Unlock()

		os.Setenv("DEEPSEEK_BASE_URL", "https://custom.api.com/v1")
		defer os.Unsetenv("DEEPSEEK_BASE_URL")

		url := getDeepSeekBaseURL()
		// If Ollama is not running (likely in this environment), it should return the env var
		if url != "https://custom.api.com/v1" && url != "http://localhost:11434/v1" {
			t.Errorf("expected custom URL or Ollama URL, got %s", url)
		}
		if url == "https://custom.api.com/v1" {
			t.Log("Successfully picked up DEEPSEEK_BASE_URL")
		} else {
			t.Log("Ollama detected at localhost:11434")
		}
	})

	t.Run("DefaultValue", func(t *testing.T) {
		// Reset cache for test
		baseURLMu.Lock()
		lastBaseURLCheck = time.Time{}
		cachedBaseURL = ""
		baseURLMu.Unlock()

		os.Unsetenv("DEEPSEEK_BASE_URL")
		url := getDeepSeekBaseURL()
		if url != "https://api.deepseek.com/v1" && url != "http://localhost:11434/v1" {
			t.Errorf("expected default URL or Ollama URL, got %s", url)
		}
		if url == "https://api.deepseek.com/v1" {
			t.Log("Successfully used default DeepSeek URL")
		} else {
			t.Log("Ollama detected at localhost:11434")
		}
	})
}
