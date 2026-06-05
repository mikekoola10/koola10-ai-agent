package agents

import ()

// Celebrator is an interface to avoid circular dependency with main.go
type Celebrator interface {
	BroadcastSSE(eventType string, data interface{})
}

var GlobalCelebrator Celebrator

func NotifyCelebration(eventType, message string) {
	if GlobalCelebrator != nil {
		GlobalCelebrator.BroadcastSSE(eventType, map[string]string{
			"message": message,
			"tts":     message,
		})
	}

	// Update achievements
	switch eventType {
	case "trade_win":
		GlobalGamification.AwardBadge("First Trade")
		GlobalGamification.CompleteChallenge("Approve 3 trades today")
	case "grant_submitted":
		GlobalGamification.AwardBadge("Grant Guru")
	case "new_lead":
		GlobalGamification.AwardBadge("Lead Magnet")
		GlobalGamification.CompleteChallenge("Find 5 new leads")
	}
}

func BroadcastCelebration(eventType, message string) {
	NotifyCelebration(eventType, message)
}
