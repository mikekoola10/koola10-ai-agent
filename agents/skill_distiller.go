package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SkillDistiller struct {
	DeepSeekKey  string
	DeepSeekBase string
}

func NewSkillDistiller() *SkillDistiller {
	base := os.Getenv("DEEPSEEK_BASE_URL")
	if base == "" {
		base = "https://api.deepseek.com"
	}
	return &SkillDistiller{
		DeepSeekKey:  os.Getenv("DEEPSEEK_API_KEY"),
		DeepSeekBase: base,
	}
}

func (sd *SkillDistiller) Distill(vertical, trajectory string, success bool) error {
	if sd.DeepSeekKey == "" {
		return fmt.Errorf("DEEPSEEK_API_KEY not set")
	}

	prompt := fmt.Sprintf("Analyze the following execution trajectory for the '%s' vertical. Success: %v. Trajectory: %s. \n\nIf successful, distill it into a reusable skill. If it failed, distill it into a 'lesson-learned'. Return a JSON object with fields: title, description, lessons (array of strings).", vertical, success, trajectory)

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "You are the Koola10 Skill Distiller. Your goal is to extract reusable knowledge from agent executions."},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	body, _ := json.Marshal(dsReq)
	req, _ := http.NewRequest("POST", sd.DeepSeekBase+"/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+sd.DeepSeekKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
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
		return err
	}

	if len(dsRes.Choices) == 0 {
		return fmt.Errorf("no response from DeepSeek")
	}

	var distilled map[string]interface{}
	if err := json.Unmarshal([]byte(dsRes.Choices[0].Message.Content), &distilled); err != nil {
		return err
	}

	title, _ := distilled["title"].(string)
	description, _ := distilled["description"].(string)
	lessons, _ := distilled["lessons"].([]interface{})

	statusStr := "FAILURE"
	if success {
		statusStr = "SUCCESS"
	}

	skillMD := fmt.Sprintf("# %s\n\n**Vertical:** %s\n**Status:** %s\n**Timestamp:** %s\n\n## Description\n%s\n\n## Lessons Learned\n", title, vertical, statusStr, time.Now().Format(time.RFC3339), description)
	for _, l := range lessons {
		skillMD += fmt.Sprintf("- %v\n", l)
	}
	skillMD += fmt.Sprintf("\n## Original Trajectory\n%s\n", trajectory)

	dir := filepath.Join("/data/skills", vertical)
	os.MkdirAll(dir, 0755)
	filename := strings.ReplaceAll(strings.ToLower(title), " ", "_") + ".md"
	if filename == ".md" {
		filename = fmt.Sprintf("skill_%d.md", time.Now().Unix())
	}
	return os.WriteFile(filepath.Join(dir, filename), []byte(skillMD), 0644)
}

func (sd *SkillDistiller) GetRelevantSkills(vertical, task string, limit int) []string {
	dir := filepath.Join("/data/skills", vertical)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var skills []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".md") {
			content, err := os.ReadFile(filepath.Join(dir, f.Name()))
			if err == nil {
				skills = append(skills, string(content))
			}
		}
		if len(skills) >= limit {
			break
		}
	}
	return skills
}
