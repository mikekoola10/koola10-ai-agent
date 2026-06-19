package agents

import (
	"fmt"
	"strings"
	"time"
)

type HealthRoutineAgent struct {
	HealthAgent
}

func (a *HealthRoutineAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	if task == "generate" {
		return a.GenerateRoutine()
	}
	return nil, fmt.Errorf("unknown task")
}

func (a *HealthRoutineAgent) GenerateRoutine() (*HealthRoutine, error) {
	inventory, _ := LoadInventory()
	conflicts, _ := LoadConflicts()

	routine := &HealthRoutine{
		ID:        fmt.Sprintf("routine_%d", time.Now().Unix()),
		Name:      "Daily Health Routine",
		Items:     []RoutineItem{},
		CreatedAt: time.Now(),
	}

	for _, item := range inventory {
		if item.Quantity <= 0 { continue }

		// Conflict Detection
		hasConflict := false
		for _, c := range conflicts {
			if strings.EqualFold(c.Ingredient, item.Name) {
				// Check against existing routine items
				for _, rItem := range routine.Items {
					for _, conflictName := range c.Conflicts {
						if strings.EqualFold(rItem.Supplement, conflictName) {
							hasConflict = true
							break
						}
					}
					if hasConflict { break }
				}
			}
			if hasConflict { break }
		}

		if !hasConflict {
			routine.Items = append(routine.Items, RoutineItem{
				Supplement: item.Name,
				Dosage:     "1 unit", // Default
				Time:       "08:00",  // Default morning
			})
		}
	}

	SaveRoutine(routine)
	return routine, nil
}
