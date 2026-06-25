package beatsmith

import (
	"fmt"
)

type ProjectStructure struct {
	Bars         int
	Sections     []string
	Tempo        float64
	Key          string
	EffectsChain map[string][]string
}

type ArrangementAdvisor struct{}

func (a *ArrangementAdvisor) AnalyzeProject(filePath string) (ProjectStructure, error) {
	// Mock deeper FLP parsing
	return ProjectStructure{
		Bars:     64,
		Sections: []string{"Intro", "Verse 1", "Chorus", "Verse 2"},
		Tempo:    140.0,
		Key:      "C Minor",
		EffectsChain: map[string][]string{
			"Kick":   {"Parametric EQ 2", "Soft Tube"},
			"Melody": {"Serum", "RC-20", "ValhallaSupermassive"},
			"Master": {"Fruity Limiter", "Soft Clipper"},
		},
	}, nil
}

func (a *ArrangementAdvisor) SuggestImprovements(structure ProjectStructure) []string {
	suggestions := []string{}
	if len(structure.Sections) < 5 {
		suggestions = append(suggestions, "Add a bridge or breakdown after the second chorus for variety.")
	}
	if structure.Bars < 100 {
		suggestions = append(suggestions, "Consider extending the outro or adding an extra verse to reach standard song length.")
	}
	return suggestions
}

func (a *ArrangementAdvisor) GenerateReport(filePath string) (string, error) {
	ps, err := a.AnalyzeProject(filePath)
	if err != nil {
		return "", err
	}
	suggestions := a.SuggestImprovements(ps)

	report := fmt.Sprintf("Arrangement Report for %s:\n", filePath)
	report += fmt.Sprintf("Tempo: %.1f BPM | Key: %s\n", ps.Tempo, ps.Key)
	report += fmt.Sprintf("Total Bars: %d\n", ps.Bars)
	report += "Current Sections: "
	for i, s := range ps.Sections {
		report += s
		if i < len(ps.Sections)-1 {
			report += ", "
		}
	}
	report += "\n\nEffects Chain Summary:\n"
	for track, fx := range ps.EffectsChain {
		report += fmt.Sprintf("- %s: %v\n", track, fx)
	}

	report += "\nSuggestions:\n"
	for _, s := range suggestions {
		report += "- " + s + "\n"
	}

	return report, nil
}
