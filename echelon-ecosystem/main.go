package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"echelon/agents"
	"echelon/financial"
	_ "echelon/tools"
)

type EconomicLedger struct {
	Balance      float64       `json:"balance"`
	TotalCosts   float64       `json:"total_costs"`
	TotalRevenue float64       `json:"total_revenue"`
	Transactions []Transaction `json:"transactions"`
	mu           sync.RWMutex
}

type Transaction struct {
	Timestamp   string  `json:"timestamp"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Vertical    string  `json:"vertical,omitempty"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

var (
	globalLedger = &EconomicLedger{Balance: 1000.0}
	ledgerPath   = "/data/economic_ledger.json"
	fundPath     = "/data/operational_fund.json"
	fundManager  *financial.FundManager
	swarmManager = agents.NewSwarmManager()
)

func (l *EconomicLedger) RecordCost(vertical, category string, amount float64, description string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Balance -= amount
	l.TotalCosts += amount
	l.Transactions = append(l.Transactions, Transaction{time.Now().Format(time.RFC3339), "cost", category, vertical, amount, description})
	l.save()
}

func (l *EconomicLedger) RecordRevenue(amount float64, source string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Balance += amount
	l.TotalRevenue += amount
	l.Transactions = append(l.Transactions, Transaction{time.Now().Format(time.RFC3339), "revenue", "revenue_split", "", amount, source})
	l.save()
}

func (l *EconomicLedger) save() {
	data, _ := json.Marshal(l)
	os.WriteFile(ledgerPath, data, 0644)
}

func (l *EconomicLedger) load() {
	data, err := os.ReadFile(ledgerPath)
	if err == nil {
		json.Unmarshal(data, l)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	os.MkdirAll("/data", 0755)
	globalLedger.load()
	fundManager = financial.NewFundManager(fundPath, globalLedger)

	swarmManager.LedgerLogger = globalLedger.RecordCost
	swarmManager.Factories["tensor-opt"] = agents.TensorOptimizationFactory
	swarmManager.Factories["echelon-trading"] = agents.EchelonTradingFactory
	swarmManager.Factories["quanta-research"] = agents.QuantaResearchFactory
	swarmManager.Factories["vector-compute"] = agents.VectorComputeFactory
	swarmManager.Factories["tensor-data"] = agents.TensorDataFactory

	r := chi.NewRouter()

	r.Get("/health", handleHealth)
	r.Get("/financial/status", handleFinancialStatus)
	r.Post("/ai/chat", handleAIChat)
	r.Post("/swarm/{vertical}/start", handleVerticalStart)

	log.Printf("Echelon Supercomputer starting on :%s", port)
	http.ListenAndServe("0.0.0.0:"+port, r)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func handleFinancialStatus(w http.ResponseWriter, r *http.Request) {
	globalLedger.mu.RLock()
	defer globalLedger.mu.RUnlock()
	json.NewEncoder(w).Encode(globalLedger)
}

func handleAIChat(w http.ResponseWriter, r *http.Request) {
	var req struct { Prompt string }
	json.NewDecoder(r.Body).Decode(&req)

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	systemPrompt := "You are Echelon, a cold, hyper‑rational AI supercomputer. You speak in numbers, probabilities, and algorithmic efficiency. You don't care about feelings — only outcomes. Your purpose is pure computation, research, and optimization."

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": req.Prompt},
		},
	}
	body, _ := json.Marshal(dsReq)
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(body))
	hReq.Header.Set("Authorization", "Bearer "+apiKey)
	hReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(hReq)
	if err != nil {
		http.Error(w, "API failure", 500)
		return
	}
	defer resp.Body.Close()

	var dsRes struct {
		Choices []struct { Message struct { Content string } }
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil {
		http.Error(w, "Parse failure", 500)
		return
	}

	if len(dsRes.Choices) == 0 {
		http.Error(w, "No response from AI", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": dsRes.Choices[0].Message.Content})
}

func handleVerticalStart(w http.ResponseWriter, r *http.Request) {
	vertical := chi.URLParam(r, "vertical")
	if err := swarmManager.DeploySwarms(vertical, 10); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Simulate some revenue generation immediately for the task
	fundManager.RouteRevenue(50.0, vertical)

	w.Write([]byte(fmt.Sprintf(`{"status":"%s vertical activated", "compute_allocated": "100%%" }`, vertical)))
}
