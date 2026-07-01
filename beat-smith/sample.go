package beatsmith

import (
	"strings"
)

type SampleMetadata struct {
	FilePath string `json:"file_path"`
	Genre    string `json:"genre"`
	Tempo    int    `json:"tempo"`
	Key      string `json:"key"`
	Tags     []string `json:"tags"`
}

type SampleManager struct {
	Samples []SampleMetadata
}

func (s *SampleManager) ScanDirectory(path string) []SampleMetadata {
	// Mock scanning and auto-tagging
	return []SampleMetadata{
		{FilePath: path + "/kick_01.wav", Genre: "Trap", Tempo: 140, Key: "C", Tags: []string{"kick", "hard"}},
		{FilePath: path + "/snare_01.wav", Genre: "Trap", Tempo: 140, Key: "N/A", Tags: []string{"snare", "crisp"}},
	}
}

func (s *SampleManager) SearchSimilar(query string) []SampleMetadata {
	// Mock semantic search logic
	results := []SampleMetadata{}
	queryLower := strings.ToLower(query)
	for _, sample := range s.Samples {
		match := false
		if strings.Contains(strings.ToLower(sample.Genre), queryLower) { match = true }
		for _, tag := range sample.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) { match = true }
		}
		if match {
			results = append(results, sample)
		}
	}
	return results
}

func (s *SampleManager) SemanticSearch(description string) ([]SampleMetadata, error) {
	// In a real implementation, this would call the semantic-agent
	// for vector-based search.
	return s.SearchSimilar(description), nil
}
