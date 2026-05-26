package agents

import (
	"time"
)

type ProactiveAgent struct {
	ConvMemory *ConversationMemory
	Broadcast  func(typeStr string, content interface{}, agent string)
}

func NewProactiveAgent(memory *ConversationMemory, broadcast func(string, interface{}, string)) *ProactiveAgent {
	return &ProactiveAgent{
		ConvMemory: memory,
		Broadcast:  broadcast,
	}
}

func (pa *ProactiveAgent) StartLoop() {
	go pa.checkTimerLoop()
}

func (pa *ProactiveAgent) checkTimerLoop() {
	ticker := time.NewTicker(15 * time.Minute)
	for range ticker.C {
		pa.checkTriggers()
	}
}

func (pa *ProactiveAgent) checkTriggers() {
	now := time.Now()

	// 1. Idle timer: 4 hours
	pa.ConvMemory.mu.RLock()
	lastInteraction := pa.ConvMemory.State.LastInteractionTime
	pa.ConvMemory.mu.RUnlock()

	if !lastInteraction.IsZero() && now.Sub(lastInteraction) > 4*time.Hour {
		pa.Broadcast("speak", "Hey there, I haven't heard from you in a while. Just checking in to see if you need anything!", "Sable")
		pa.ConvMemory.UpdateInteraction()
		pa.ConvMemory.Save()
		return
	}

	// 2. Scheduled briefings: 9 AM and 6 PM
	hour := now.Hour()
	minute := now.Minute()

	if (hour == 9 && minute < 15) {
		pa.Broadcast("speak", "Good morning! I'm ready for our daily briefing. We have several pending grant opportunities and the trading pool is performing well.", "Sable")
	} else if (hour == 18 && minute < 15) {
		pa.Broadcast("speak", "Good evening. Wrapping up the day's work. The lead generation swarm found 10 new high-quality leads today.", "Vega")
	}
}

func (pa *ProactiveAgent) TriggerEvent(event string, details string, agent string) {
	pa.Broadcast("speak", details, agent)
}
