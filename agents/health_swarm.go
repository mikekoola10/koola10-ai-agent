package agents

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type InventoryItem struct {
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	Threshold int    `json:"threshold"`
	Category  string `json:"category"`
}

type HealthSchedule struct {
	Name       string `json:"name"`
	Time       string `json:"time"` // "HH:MM"
	Recurrence string `json:"recurrence"` // "daily", "weekly", etc.
	Active     bool   `json:"active"`
}

type VideoLink struct {
	URL         string       `json:"url"`
	Title       string       `json:"title"`
	Summary     string       `json:"summary"`
	Ingredients []Supplement `json:"ingredients,omitempty"`
	DateAdded   time.Time    `json:"date_added"`
}

type HealthJournalEntry struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Mood      string    `json:"mood,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type Supplement struct {
	Name      string `json:"name"`
	Dosage    string `json:"dosage"`
	Frequency string `json:"frequency,omitempty"`
}

type HealthRoutine struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Items     []RoutineItem    `json:"items"`
	CreatedAt time.Time        `json:"created_at"`
}

type RoutineItem struct {
	Time       string `json:"time"`
	Supplement string `json:"supplement"`
	Dosage     string `json:"dosage"`
}

type HealthConflict struct {
	Ingredient string   `json:"ingredient"`
	MaxDose    string   `json:"max_dose"`
	Conflicts  []string `json:"conflicts"` // Names of conflicting supplements/ingredients
}

type HealthAgent struct {
	specialty string
	status    AgentStatus
}

func (a *HealthAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()
	// Health swarm logic can be extended here
	return "Health Swarm (" + a.specialty + "): " + task, nil
}

func (a *HealthAgent) Status() AgentStatus { return a.status }
func (a *HealthAgent) Specialty() string    { return a.specialty }

func HealthSwarmFactory() []SpecialistAgent {
	specialties := []string{
		"Inventory Monitor", "Schedule Coordinator", "Video Librarian", "Health Journaler",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &HealthAgent{specialty: s, status: StatusIdle})
	}
	return agents
}

// Data Storage Helpers

var (
	inventoryPath = "/data/health_inventory.json"
	schedulePath  = "/data/health_schedule.json"
	videosPath    = "/data/health_videos.json"
	journalPath   = "/data/health_journal.json"
	routinePath   = "/data/health_routine.json"
	conflictsPath = "/data/health_conflicts.json"
	healthMu      sync.RWMutex
)

func LoadInventory() ([]InventoryItem, error) {
	healthMu.RLock()
	defer healthMu.RUnlock()
	data, err := os.ReadFile(inventoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []InventoryItem{}, nil
		}
		return nil, err
	}
	var items []InventoryItem
	err = json.Unmarshal(data, &items)
	return items, err
}

func SaveInventory(items []InventoryItem) error {
	healthMu.Lock()
	defer healthMu.Unlock()
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(inventoryPath, data, 0644)
}

func LoadSchedules() ([]HealthSchedule, error) {
	healthMu.RLock()
	defer healthMu.RUnlock()
	data, err := os.ReadFile(schedulePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []HealthSchedule{}, nil
		}
		return nil, err
	}
	var items []HealthSchedule
	err = json.Unmarshal(data, &items)
	return items, err
}

func SaveSchedules(items []HealthSchedule) error {
	healthMu.Lock()
	defer healthMu.Unlock()
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(schedulePath, data, 0644)
}

func LoadVideos() ([]VideoLink, error) {
	healthMu.RLock()
	defer healthMu.RUnlock()
	data, err := os.ReadFile(videosPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []VideoLink{}, nil
		}
		return nil, err
	}
	var items []VideoLink
	err = json.Unmarshal(data, &items)
	return items, err
}

func SaveVideos(items []VideoLink) error {
	healthMu.Lock()
	defer healthMu.Unlock()
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(videosPath, data, 0644)
}

func LoadJournal() ([]HealthJournalEntry, error) {
	healthMu.RLock()
	defer healthMu.RUnlock()
	data, err := os.ReadFile(journalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []HealthJournalEntry{}, nil
		}
		return nil, err
	}
	var items []HealthJournalEntry
	err = json.Unmarshal(data, &items)
	return items, err
}

func SaveJournal(items []HealthJournalEntry) error {
	healthMu.Lock()
	defer healthMu.Unlock()
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(journalPath, data, 0644)
}

func LoadRoutine() (*HealthRoutine, error) {
	healthMu.RLock()
	defer healthMu.RUnlock()
	data, err := os.ReadFile(routinePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &HealthRoutine{}, nil
		}
		return nil, err
	}
	var routine HealthRoutine
	err = json.Unmarshal(data, &routine)
	return &routine, err
}

func SaveRoutine(routine *HealthRoutine) error {
	healthMu.Lock()
	defer healthMu.Unlock()
	data, err := json.MarshalIndent(routine, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(routinePath, data, 0644)
}

func LoadConflicts() ([]HealthConflict, error) {
	healthMu.RLock()
	defer healthMu.RUnlock()
	data, err := os.ReadFile(conflictsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []HealthConflict{}, nil
		}
		return nil, err
	}
	var items []HealthConflict
	err = json.Unmarshal(data, &items)
	return items, err
}

func SaveConflicts(items []HealthConflict) error {
	healthMu.Lock()
	defer healthMu.Unlock()
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(conflictsPath, data, 0644)
}
