package agents

import (
	"log"
	"os"
	"strings"
	"time"
)

type ProactiveAgent struct {
	Broadcaster func(event string, data interface{})
}

func NewProactiveAgent(broadcaster func(string, interface{})) *ProactiveAgent {
	return &ProactiveAgent{
		Broadcaster: broadcaster,
	}
}

func (pa *ProactiveAgent) Start() {
	go pa.runLoop()
}

func (pa *ProactiveAgent) runLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			pa.checkSchedule(now)
		}
	}
}

func (pa *ProactiveAgent) checkSchedule(t time.Time) {
	hour := t.Hour()
	minute := t.Minute()
	weekday := t.Weekday()

	modules := GetCoachingModules()

	// 8:00 AM - Module 1
	if hour == 8 && minute == 0 {
		pa.triggerModule(modules[0], "daily_mental_model")
	}

	// 12:00 PM - Module 2 or 3
	if hour == 12 && minute == 0 {
		// Toggle between 2 and 3
		if t.Day()%2 == 0 {
			pa.triggerModule(modules[1], "micro_lesson")
		} else {
			pa.triggerModule(modules[2], "micro_lesson")
		}
	}

	// 6:00 PM - Module 5 or 7
	if hour == 18 && minute == 0 {
		if t.Day()%2 == 0 {
			pa.triggerModule(modules[4], "reflection")
		} else {
			pa.triggerModule(modules[6], "reflection")
		}
	}

	// Sunday 10:00 AM - Module 4
	if weekday == time.Sunday && hour == 10 && minute == 0 {
		pa.triggerModule(modules[3], "weekly_audit")
	}
}

func (pa *ProactiveAgent) triggerModule(m *CoachingModule, context string) {
	ecosystem := strings.ToUpper(os.Getenv("ECOSYSTEM"))
	if ecosystem == "" {
		ecosystem = "KOOLA10"
	}

	// Filter based on assigned agents/ecosystems
	// Module 4 is Oracle ONLY
	if m.ID == "4" && ecosystem != "ORACLE" {
		return
	}

	content, err := m.Run(context)
	if err != nil {
		log.Printf("Error running coaching module %s: %v", m.ID, err)
		return
	}

	if pa.Broadcaster != nil {
		pa.Broadcaster("proactive_speech", map[string]interface{}{
			"module_id": m.ID,
			"module_name": m.Name,
			"content": content,
			"type": "coaching",
		})
	}
}
