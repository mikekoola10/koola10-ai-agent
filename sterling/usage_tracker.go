package sterling

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type UsageRecord struct {
	ServiceName string    `json:"service_name"`
	LastUsed    time.Time `json:"last_used"`
	CallCount   int       `json:"call_count"`
}

type UsageTracker struct {
	mu       sync.Mutex
	records  map[string]*UsageRecord
	filePath string
}

func NewUsageTracker(filePath string) *UsageTracker {
	t := &UsageTracker{
		records:  make(map[string]*UsageRecord),
		filePath: filePath,
	}
	t.load()
	return t
}

func (ut *UsageTracker) RecordUse(serviceName string) {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	if rec, ok := ut.records[serviceName]; ok {
		rec.LastUsed = time.Now()
		rec.CallCount++
	} else {
		ut.records[serviceName] = &UsageRecord{
			ServiceName: serviceName,
			LastUsed:    time.Now(),
			CallCount:   1,
		}
	}
	ut.save()
}

func (ut *UsageTracker) GetIdleServices(days int) []string {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	idle := []string{}
	for name, rec := range ut.records {
		if rec.LastUsed.Before(cutoff) {
			idle = append(idle, name)
		}
	}
	return idle
}

func (ut *UsageTracker) save() {
	data, _ := json.MarshalIndent(ut.records, "", "  ")
	os.WriteFile(ut.filePath, data, 0644)
}

func (ut *UsageTracker) load() {
	data, err := os.ReadFile(ut.filePath)
	if err == nil {
		json.Unmarshal(data, &ut.records)
	}
}
