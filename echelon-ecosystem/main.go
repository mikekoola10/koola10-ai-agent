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
	"crypto/rand"
	"encoding/hex"

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

type ApprovalRequest struct {
	ID      string                 `json:"approval_id"`
	Action  string                 `json:"action"`
	Details map[string]interface{} `json:"details"`
	Status  string                 `json:"status"`
}

var (
	globalLedger = &EconomicLedger{Balance: 1000.0}
	ledgerPath   = "/data/economic_ledger.json"
	fundPath     = "/data/operational_fund.json"
	portfolioPath = "/data/portfolio.json"
	fundManager  *financial.FundManager
	portfolioManager *financial.PortfolioManager
	swarmManager = agents.NewSwarmManager()

	approvalStore = make(map[string]*ApprovalRequest)
	approvalMu    sync.Mutex
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

func (l *EconomicLedger) RecordRevenueWithVertical(vertical string, amount float64, source string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Balance += amount
	l.TotalRevenue += amount
	l.Transactions = append(l.Transactions, Transaction{time.Now().Format(time.RFC3339), "revenue", "revenue_split", vertical, amount, source})
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

func generateID() string {
	b := make([]byte, 8); rand.Read(b); return hex.EncodeToString(b)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }

	os.MkdirAll("/data", 0755)
	globalLedger.load()
	fundManager = financial.NewFundManager(fundPath, globalLedger)
	portfolioManager = financial.NewPortfolioManager(portfolioPath, globalLedger)

	swarmManager.LedgerLogger = globalLedger.RecordCost
	swarmManager.Factories["tensor-opt"] = agents.TensorOptimizationFactory
	swarmManager.Factories["trading_v2"] = agents.EchelonTradingV2Factory
	swarmManager.Factories["quanta-research"] = agents.QuantaResearchFactory
	swarmManager.Factories["vector-compute"] = agents.VectorComputeFactory
	swarmManager.Factories["tensor-data"] = agents.TensorDataFactory

	// Automated crypto profit sweep (every 24h)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for {
			globalLedger.mu.RLock()
			portfolioManager.SweepProfits(globalLedger.Transactions)
			globalLedger.mu.RUnlock()
			<-ticker.C
		}
	}()

	r := chi.NewRouter()

	r.Get("/health", handleHealth)
	r.Get("/financial/status", handleFinancialStatus)
	r.Post("/ai/chat", handleAIChat)
	r.Post("/swarm/{vertical}/start", handleVerticalStart)
	r.Get("/swarm/status", handleSwarmStatus)

	// Compliance/Approval flow for Echelon
	r.Post("/compliance/approval", handleComplianceApproval)
	r.Post("/compliance/approve", handleComplianceApprove)
	r.Get("/compliance/audit", handleComplianceAudit)

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
	revenue := 50.0
	switch vertical {
	case "tensor-opt": revenue = 499.0
	case "trading_v2": revenue = 120.0
	case "quanta-research": revenue = 299.0
	case "vector-compute": revenue = 15.0
	case "tensor-data": revenue = 99.0
	}

	fundManager.RouteRevenue(revenue, vertical)

	w.Write([]byte(fmt.Sprintf(`{"status":"%s vertical activated", "compute_allocated": "100%%" }`, vertical)))
}

func handleSwarmStatus(w http.ResponseWriter, r *http.Request) {
	swarmManager.Mu.RLock()
	defer swarmManager.Mu.RUnlock()

	status := make(map[string]interface{})
	for v := range swarmManager.Swarms {
		status[v] = swarmManager.GetSwarmStatus(v)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func handleComplianceApproval(w http.ResponseWriter, r *http.Request) {
	var req ApprovalRequest
	json.NewDecoder(r.Body).Decode(&req)
	req.ID = generateID()
	req.Status = "pending"
	approvalMu.Lock()
	approvalStore[req.ID] = &req
	approvalMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

func handleComplianceApprove(w http.ResponseWriter, r *http.Request) {
	var req struct { ApprovalID string }
	json.NewDecoder(r.Body).Decode(&req)
	approvalMu.Lock()
	defer approvalMu.Unlock()
	if ap, ok := approvalStore[req.ApprovalID]; ok {
		ap.Status = "approved"
		json.NewEncoder(w).Encode(ap)
	} else {
		http.Error(w, "not found", 404)
	}
}

func handleComplianceAudit(w http.ResponseWriter, r *http.Request) {
	approvalMu.Lock()
	defer approvalMu.Unlock()
	var res []ApprovalRequest
	for _, v := range approvalStore {
		res = append(res, *v)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
