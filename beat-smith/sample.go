package beatsmith

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
	// Mock semantic search
	return s.Samples
}
