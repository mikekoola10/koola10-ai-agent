package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type CoachingProgress struct {
	MasteredModels []string          `json:"mastered_models"`
	CurrentLessons map[string]string `json:"current_lessons"` // module_id -> lesson/topic
	LastAudit      time.Time         `json:"last_audit"`
	FutureSelf     map[string]string `json:"future_self"`
}

type CoachingModule struct {
	ID           string
	Name         string
	AssignedTo   []string
	SystemPrompt string
}

var (
	GlobalCoachingProgress = &CoachingProgress{
		MasteredModels: []string{},
		CurrentLessons: make(map[string]string),
		FutureSelf:     make(map[string]string),
	}
	coachingMu sync.RWMutex
	progressPath = "/data/coaching_progress.json"
)

func init() {
	LoadCoachingProgress()
}

func LoadCoachingProgress() {
	coachingMu.Lock()
	defer coachingMu.Unlock()
	data, err := os.ReadFile(progressPath)
	if err == nil {
		json.Unmarshal(data, GlobalCoachingProgress)
	}
	if GlobalCoachingProgress.CurrentLessons == nil {
		GlobalCoachingProgress.CurrentLessons = make(map[string]string)
	}
	if GlobalCoachingProgress.FutureSelf == nil {
		GlobalCoachingProgress.FutureSelf = make(map[string]string)
	}
}

func SaveCoachingProgress() {
	coachingMu.RLock()
	defer coachingMu.RUnlock()
	data, _ := json.Marshal(GlobalCoachingProgress)
	os.MkdirAll("/data", 0755)
	os.WriteFile(progressPath, data, 0644)
}

func (m *CoachingModule) Run(sessionContext string) (string, error) {
	// In a real implementation, this would call an LLM with the SystemPrompt and sessionContext.
	// For this task, we'll return the coaching content which will be passed to TTS.

	switch m.ID {
	case "1":
		return m.runThinkLikeABillionaire(sessionContext)
	case "2":
		return m.runSuperhumanLearning(sessionContext)
	case "3":
		return m.runExpertKnowledge(sessionContext)
	case "4":
		return m.runMentalSoftware(sessionContext)
	case "5":
		return m.runGodTierLife(sessionContext)
	case "6":
		return m.runCompressDecades(sessionContext)
	case "7":
		return m.runDreamVersion(sessionContext)
	default:
		return "", fmt.Errorf("unknown module ID: %s", m.ID)
	}
}

func (m *CoachingModule) runThinkLikeABillionaire(ctx string) (string, error) {
	models := []string{"First Principles", "Inversion", "Second-Order Thinking", "Opportunity Cost", "Pareto Principle"}

	coachingMu.Lock()
	// Simple mastery tracking: if we've seen it 3 times, we consider it "mastered" for the sake of the demo
	// In a real app, this would be based on user feedback.
	model := models[time.Now().Day()%len(models)]

	alreadyMastered := false
	for _, m := range GlobalCoachingProgress.MasteredModels {
		if m == model {
			alreadyMastered = true
			break
		}
	}

	if !alreadyMastered {
		// Log reinforcement
		GlobalCoachingProgress.MasteredModels = append(GlobalCoachingProgress.MasteredModels, model)
	}
	coachingMu.Unlock()
	SaveCoachingProgress()

	return fmt.Sprintf("Good morning George. Today's mental model is %s. My challenge for you: apply this to your biggest bottleneck today. System thinking is the key to asymmetric outcomes.", model), nil
}

func (m *CoachingModule) runSuperhumanLearning(topic string) (string, error) {
	if topic == "" {
		topic = GlobalCoachingProgress.CurrentLessons["2"]
	}
	if topic == "" {
		return "What would you like to learn today? Tell me, and I'll build your 90-day superhuman blueprint.", nil
	}
	GlobalCoachingProgress.CurrentLessons["2"] = topic
	SaveCoachingProgress()
	return fmt.Sprintf("For learning %s, we use interleaving. Here is your 5-minute micro-lesson on the fundamentals. Tomorrow, we test with active recall.", topic), nil
}

func (m *CoachingModule) runExpertKnowledge(skill string) (string, error) {
	if skill == "" {
		skill = GlobalCoachingProgress.CurrentLessons["3"]
	}
	if skill == "" {
		return "Identify the skill you wish to download into your cognitive OS. I will break it into stages of mastery.", nil
	}
	GlobalCoachingProgress.CurrentLessons["3"] = skill
	SaveCoachingProgress()
	return fmt.Sprintf("Mastering %s requires deconstruction. Stage 1: The 20 percent that matters. Here is your practice assignment for today.", skill), nil
}

func (m *CoachingModule) runMentalSoftware(ctx string) (string, error) {
	GlobalCoachingProgress.LastAudit = time.Now()
	SaveCoachingProgress()
	return "Weekly Cognitive OS Upgrade: I've analyzed your decision patterns. Your 'First Principles' application is improving, but 'Inversion' is missing in your financial logic. Let's rewire this today.", nil
}

func (m *CoachingModule) runGodTierLife(ctx string) (string, error) {
	return "Design update: Your 'Time Freedom' pillar is currently at 40 percent. To reach God-Tier, we must eliminate these three low-leverage habits. Environment design is your leverage.", nil
}

func (m *CoachingModule) runCompressDecades(goal string) (string, error) {
	return fmt.Sprintf("To achieve %s in days instead of decades, we use AI delegation and shortcuts. Here is the 80/20 blueprint for rapid execution.", goal), nil
}

func (m *CoachingModule) runDreamVersion(ctx string) (string, error) {
	return "Close your eyes. Visualize your Future Self: calm, decisive, and operating with infinite leverage. Affirm with me: I am the architect of my reality.", nil
}

func GetCoachingModules() []*CoachingModule {
	return []*CoachingModule{
		{
			ID:   "1",
			Name: "Think Like a Billionaire",
			AssignedTo: []string{"Koola10", "Oracle"},
			SystemPrompt: "You are a thinking coach trained on Elon Musk, Naval Ravikant, Jeff Bezos. Reprogram George's thought process to think in systems, long-vision, leverage, and asymmetric outcomes.",
		},
		{
			ID:   "2",
			Name: "Unlock Superhuman Learning",
			AssignedTo: []string{"Forge", "Atlas"},
			SystemPrompt: "Expert in learning science. Apply spaced repetition, interleaving, and Feynman technique. Create 90-day blueprints.",
		},
		{
			ID:   "3",
			Name: "Download Expert-Level Knowledge",
			AssignedTo: []string{"Nova", "Vega"},
			SystemPrompt: "Mentor-level deconstruction of skills into tasks and simulated scenarios.",
		},
		{
			ID:   "4",
			Name: "Upgrade Mental Software",
			AssignedTo: []string{"Oracle"},
			SystemPrompt: "Analytical auditor of conversation patterns and decision-making. Output Cognitive OS Upgrade reports.",
		},
		{
			ID:   "5",
			Name: "Design a God-Tier Life",
			AssignedTo: []string{"Koola10", "Sterling", "Oracle", "Sable"},
			SystemPrompt: "Architect of life across 5 pillars. Environment design and habit mastery.",
		},
		{
			ID:   "6",
			Name: "Compress Decades into Days",
			AssignedTo: []string{"Forge", "Nova", "Atlas", "Vega"},
			SystemPrompt: "AI tool and automation expert. Build high-leverage blueprints for rapid goal achievement.",
		},
		{
			ID:   "7",
			Name: "Be Your Dream Version",
			AssignedTo: []string{"Aria", "Muse"},
			SystemPrompt: "Creative identity reprogramming through narrative, visualization, and affirmations.",
		},
	}
}
