package beatsmith

import (
	"testing"
)

func TestSuggestVSTs(t *testing.T) {
	vm := &VSTManager{}
	profile := ProducerProfile{Genres: []string{"Trap"}}
	suggestions := vm.SuggestVSTs(profile)
	if len(suggestions) == 0 {
		t.Error("Expected suggestions for Trap genre")
	}
	if suggestions[0].Name != "Serum" {
		t.Errorf("Expected Serum, got %s", suggestions[0].Name)
	}
}

func TestMasteringRecommendations(t *testing.T) {
	ma := &MasteringAssistant{}
	settings := ma.RecommendSettings("Lofi")
	if settings.Limiter != "-1.0dB ceiling, soft clip." {
		t.Errorf("Unexpected Lofi mastering settings: %v", settings)
	}
}
