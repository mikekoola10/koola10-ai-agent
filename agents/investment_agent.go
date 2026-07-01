package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Recommendation struct {
	Action       string  `json:"action"` // BUY, SELL, HOLD, PASS
	Conviction   int     `json:"conviction"` // 1-10
	PositionSize float64 `json:"position_size_recommendation"`
	RiskFactors  string  `json:"risk_factors"`
	TimeHorizon  string  `json:"time_horizon"`
	Analysis     string  `json:"analysis"`
}

type InvestmentAgent struct{}

func (a *InvestmentAgent) AnalyzeOpportunity(symbol string, details string) (*Recommendation, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY not set")
	}

	systemPrompt := `You are Echelon's Investment Agent, a hyper-rational, quantitative investor. You analyze opportunities using first principles, discounted cash flow, comparable company analysis, and risk assessment. You think in probabilities, not emotions. You never fall for hype. You calculate the expected value of every investment. Return your analysis in JSON format.`

	userPrompt := fmt.Sprintf("Analyze the investment opportunity for %s with the following details: %s. Provide a structured recommendation with action (BUY, SELL, HOLD, PASS), conviction level (1-10), position size recommendation (percentage of portfolio), risk factors, and time horizon.", symbol, details)

	reqBody := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	b, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil {
		return nil, err
	}

	if len(dsRes.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	var rec Recommendation
	if err := json.Unmarshal([]byte(dsRes.Choices[0].Message.Content), &rec); err != nil {
		// Fallback if AI didn't follow JSON perfectly or slightly different keys
		return nil, fmt.Errorf("failed to parse AI recommendation: %v", err)
	}
	rec.Analysis = dsRes.Choices[0].Message.Content

	return &rec, nil
}

func (a *InvestmentAgent) CalculatePositionSize(portfolioBalance float64, riskScore float64, maxAllocation float64) float64 {
	// Kelly Criterion: fraction = (p * (b + 1) - 1) / b
	// where p is probability of win, b is odds (ratio of amount won to amount lost)

	// Use riskScore (1-10) to estimate probability of win p
	// riskScore 1 -> p = 0.9, riskScore 10 -> p = 0.51
	p := 0.9 - (riskScore-1)*(0.39/9)
	if p < 0.5 {
		p = 0.5
	}

	// Assume 2:1 risk/reward for rational quantitative bets
	b := 2.0

	kellyFraction := (p*(b+1) - 1) / b

	if kellyFraction > maxAllocation {
		kellyFraction = maxAllocation
	}

	// instruction says: conservative cap (e.g., 5% of portfolio per position)
	if kellyFraction > 0.05 {
		kellyFraction = 0.05
	}

	if kellyFraction < 0 {
		kellyFraction = 0
	}

	return portfolioBalance * kellyFraction
}
