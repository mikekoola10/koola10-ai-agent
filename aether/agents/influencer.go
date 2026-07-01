package agents

import (
	"fmt"
	"koola10/tools"
)

type InfluencerAgent struct {
	status string
}

func (a *InfluencerAgent) GeneratePost() (string, error) {
	a.status = "generating"
	// Use HF for content generation
	payload := map[string]interface{}{
		"model": "gpt2",
		"inputs": "An inspiring post about the future of AI influencers and decentralized autonomy.",
	}
	res := tools.RunTool("huggingface", payload)
	if !res.Success {
		return "", fmt.Errorf("HF failed: %s", res.Error)
	}
	a.status = "idle"
	return fmt.Sprintf("Post: %v", res.Data), nil
}
