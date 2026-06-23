package beatsmith

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type LoopGenerator struct {
	LoopsDir string
}

func (l *LoopGenerator) GenerateMIDIPattern(genre string, bpm int) (string, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("DEEPSEEK_API_KEY not set")
	}

	prompt := fmt.Sprintf("Generate a MIDI pattern for a %s beat at %d BPM. Output the MIDI notes as a JSON array of objects with 'note', 'start_time', and 'duration'.", genre, bpm)

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "You are BeatSmith, a MIDI generation expert."},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	dsBody, _ := json.Marshal(dsReq)
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	hReq.Header.Set("Authorization", "Bearer "+apiKey)
	hReq.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(hReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string
			}
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil {
		return "", err
	}

	if len(dsRes.Choices) == 0 {
		return "", fmt.Errorf("no response from DeepSeek")
	}

	pattern := dsRes.Choices[0].Message.Content
	// Save loop to file
	filename := fmt.Sprintf("%s/%s_%dBPM_loop.json", l.LoopsDir, genre, bpm)
	os.MkdirAll(l.LoopsDir, 0755)
	err = os.WriteFile(filename, []byte(pattern), 0644)

	return filename, err
}

func (l *LoopGenerator) GenerateAudioLoop(genre string, bpm int) (string, error) {
	apiKey := os.Getenv("SUNO_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("SUNO_API_KEY not set")
	}

	prompt := fmt.Sprintf("A %s audio loop at %d BPM.", genre, bpm)
	fmt.Printf("[Suno] Calling Suno API with prompt: %s\n", prompt)

	// Mock API call
	// resp, err := http.Post("https://api.suno.ai/v1/generate", ...)

	filename := fmt.Sprintf("%s/%s_%dBPM_audio.wav", l.LoopsDir, genre, bpm)
	os.MkdirAll(l.LoopsDir, 0755)
	err := os.WriteFile(filename, []byte("mock audio data"), 0644)

	return filename, err
}
