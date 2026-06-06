package agents

import (
	"encoding/json"
	"os"
	"sync"
)

type Challenge struct {
	Name      string `json:"name"`
	Completed bool   `json:"completed"`
	Current   int    `json:"current"`
	Target    int    `json:"target"`
}

type GamificationState struct {
	Badges          map[string]bool      `json:"badges"`
	DailyChallenges map[string]*Challenge `json:"daily_challenges"`
	mu              sync.RWMutex
}

var GlobalGamification = &GamificationState{
	Badges:          make(map[string]bool),
	DailyChallenges: make(map[string]*Challenge),
}

const gamificationPath = "/data/gamification.json"

func (g *GamificationState) Load() {
	g.mu.Lock()
	defer g.mu.Unlock()
	data, err := os.ReadFile(gamificationPath)
	if err == nil {
		json.Unmarshal(data, g)
	}
	if g.Badges == nil {
		g.Badges = make(map[string]bool)
	}
	if g.DailyChallenges == nil {
		g.DailyChallenges = make(map[string]*Challenge)
	}

	// Initialize default challenges if missing
	if g.DailyChallenges["Approve 3 trades today"] == nil {
		g.DailyChallenges["Approve 3 trades today"] = &Challenge{Name: "Approve 3 trades today", Target: 3}
	}
	if g.DailyChallenges["Find 5 new leads"] == nil {
		g.DailyChallenges["Find 5 new leads"] = &Challenge{Name: "Find 5 new leads", Target: 5}
	}
}

func (g *GamificationState) Save() {
	g.mu.RLock()
	defer g.mu.RUnlock()
	data, _ := json.Marshal(g)
	os.WriteFile(gamificationPath, data, 0644)
}

func (g *GamificationState) AwardBadge(badge string) {
	g.mu.Lock()
	if !g.Badges[badge] {
		g.Badges[badge] = true
		g.mu.Unlock()
		g.Save()
		return
	}
	g.mu.Unlock()
}

func (g *GamificationState) CompleteChallenge(challengeName string) {
	g.mu.Lock()
	c, ok := g.DailyChallenges[challengeName]
	if !ok {
		g.mu.Unlock()
		return
	}
	c.Current++
	if c.Current >= c.Target {
		c.Completed = true
		g.AwardBadge("Empire Builder") // Example award for any completion
	}
	g.mu.Unlock()
	g.Save()
}
