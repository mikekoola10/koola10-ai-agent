package beatsmith

import (
	"encoding/json"
	"os"
	"sync"
)

type ProducerProfile struct {
	Genres       []string `json:"genres"`
	BPMRange     [2]int   `json:"bpm_range"`
	FavoriteVSTs []string `json:"favorite_vsts"`
	Style        string   `json:"style"`
}

type Mirror struct {
	Profile ProducerProfile
	Path    string
	mu      sync.RWMutex
}

func NewMirror(path string) *Mirror {
	m := &Mirror{Path: path}
	m.Load()
	return m
}

func (m *Mirror) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := os.ReadFile(m.Path)
	if err != nil {
		// Initialize with defaults if file doesn't exist
		m.Profile = ProducerProfile{
			Genres:   []string{"Trap", "Lofi"},
			BPMRange: [2]int{80, 160},
		}
		return err
	}
	return json.Unmarshal(data, &m.Profile)
}

func (m *Mirror) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, err := json.MarshalIndent(m.Profile, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.Path, data, 0644)
}

func (m *Mirror) UpdateProfile(p ProducerProfile) error {
	m.mu.Lock()
	m.Profile = p
	m.mu.Unlock()
	return m.Save()
}
