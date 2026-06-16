package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	target := os.Getenv("TARGET_URL")
	recovery := os.Getenv("RECOVERY_URL")
	secret := os.Getenv("RECOVERY_WEBHOOK_SECRET")

	if target == "" || recovery == "" || secret == "" {
		log.Fatal("TARGET_URL, RECOVERY_URL, and RECOVERY_WEBHOOK_SECRET must be set")
	}

	log.Printf("[Wizard's Sentry] Monitoring: %s", target)
	failures := 0

	for {
		resp, err := http.Get(target)
		if err != nil || resp.StatusCode != 200 {
			failures++
			log.Printf("[Sentry] Failure %d/3: %v", failures, err)
			if failures >= 3 {
				triggerRecovery(recovery, secret)
				failures = 0
			}
		} else {
			if failures > 0 { log.Printf("[Sentry] System recovered.") }
			failures = 0
			resp.Body.Close()
		}
		time.Sleep(30 * time.Second)
	}
}

func triggerRecovery(url, secret string) {
	log.Printf("[Sentry] TRIGGERING REMOTE RECOVERY...")
	payload := map[string]string{
		"failure_name": "fly_app_down",
		"details":      "Wizard's Sentry detected persistent unreachability",
		"secret":       secret,
	}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("[Sentry] Recovery Trigger Failed: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("[Sentry] Recovery Status: %d", resp.StatusCode)
}
