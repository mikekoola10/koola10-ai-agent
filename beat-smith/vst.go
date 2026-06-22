package beatsmith

import (
	"fmt"
	"strings"
)

type VSTPlugin struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // e.g., Synth, FX
	Version  string `json:"version"`
	Category string `json:"category"`
}

type VSTManager struct {
	VSTPath          string
	InstalledPlugins []VSTPlugin
}

func (v *VSTManager) ScanPlugins() error {
	// Mock scanning VSTPath
	v.InstalledPlugins = []VSTPlugin{
		{Name: "Serum", Type: "Synth", Version: "1.0", Category: "Wavetable"},
		{Name: "FabFilter Pro-Q 3", Type: "FX", Version: "3.5", Category: "EQ"},
	}
	return nil
}

func (v *VSTManager) SuggestVSTs(profile ProducerProfile) []VSTPlugin {
	suggestions := []VSTPlugin{}
	genre := ""
	if len(profile.Genres) > 0 {
		genre = strings.ToLower(profile.Genres[0])
	}

	switch genre {
	case "trap":
		suggestions = append(suggestions, VSTPlugin{Name: "Serum", Type: "Synth", Category: "Bass/Leads"})
		suggestions = append(suggestions, VSTPlugin{Name: "Omnisphere", Type: "Synth", Category: "Pads/Keys"})
	case "lofi":
		suggestions = append(suggestions, VSTPlugin{Name: "RC-20 Retro Color", Type: "FX", Category: "Degradation"})
		suggestions = append(suggestions, VSTPlugin{Name: "Korg M1", Type: "Synth", Category: "Classic Keys"})
	default:
		suggestions = append(suggestions, VSTPlugin{Name: "Vital", Type: "Synth", Category: "General Purpose"})
	}

	return suggestions
}

func (v *VSTManager) CheckForUpdates() []string {
	// Mock logic for checking updates
	alerts := []string{}
	for _, p := range v.InstalledPlugins {
		if p.Name == "Serum" && p.Version == "1.0" {
			alerts = append(alerts, fmt.Sprintf("Update available for %s: v1.1", p.Name))
		}
	}
	return alerts
}
