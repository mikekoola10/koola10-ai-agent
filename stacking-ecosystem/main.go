package main

import (
	"bytes"
	"context"
	_ "embed"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"koola10/agents"
	"koola10/financial"
	"koola10/tools"

	"github.com/redis/go-redis/v9"
)

// --- Structs ---

type ChatRequest struct {
	Prompt  string `json:"prompt"`
}

type ChatResponse struct {
	Response   string `json:"response"`
	TokensUsed int    `json:"tokens_used"`
}

type AuditEntry struct {
	Timestamp string                 `json:"timestamp"`
	Action    string                 `json:"action"`
	Details   map[string]interface{} `json:"details"`
	Hash      string                 `json:"hash"`
}

type UsageLog struct {
	Timestamp  string  `json:"timestamp"`
	TokensUsed int     `json:"tokens_used"`
	Cost       float64 `json:"cost"`
}

type SwarmNode struct {
	ID       string `json:"node_id"`
	Region   string `json:"region"`
	Endpoint string `json:"endpoint"`
	Status   string `json:"status"`
}

// --- Global States ---

var (
	auditMutex   sync.Mutex
	usageMutex   sync.Mutex
	killSwitchMu sync.Mutex

	killSwitchPath = "/data/kill_switch"
	ledgerPath     = "/data/stacking_ledger.json"
	fundPath       = "/data/stacking_fund.json"
	auditPath      = "/data/stacking_audit.jsonl"
	usagePath      = "/data/stacking_usage.jsonl"

	fundManager *financial.StackingFundManager
	globalSwarmManager = agents.NewSwarmManager()

	rlBucket     = 15.0
	rlMaxBucket  = 15.0
	rlRate       = 10.0
	rlLastUpdate = time.Now()
	rlMu         sync.Mutex

	redisClient *redis.Client
	nodeID      string
	region      string

	//go:embed dashboard.html
	dashboardHTML string
)

// --- Middleware ---

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			return
		}
		next(w, r)
	}
}

// --- Main ---

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	region = os.Getenv("FLY_REGION")
	if region == "" { region = "local" }
	nodeID = os.Getenv("NODE_ID")
	if nodeID == "" { h, _ := os.Hostname(); nodeID = h }

	os.MkdirAll("/data", 0755)

	fundManager = financial.NewStackingFundManager(fundPath, ledgerPath)

	globalSwarmManager.AuditLogger = AddAuditEntry
	globalSwarmManager.LedgerLogger = func(vertical, category string, amount float64, description string) {
		// Log cost to the stacking fund
		fundManager.PaySubscription(vertical+":"+category, amount)
	}

	globalSwarmManager.Factories["vault"] = agents.VaultFactory
	globalSwarmManager.Factories["compound"] = agents.CompoundFactory
	globalSwarmManager.Factories["reserve"] = agents.ReserveFactory
	globalSwarmManager.Factories["forge_product"] = agents.ForgeProductFactory

	// Auto-deploy swarms for the Stacking ecosystem
	globalSwarmManager.DeploySwarms("vault", 1)
	globalSwarmManager.DeploySwarms("compound", 1)
	globalSwarmManager.DeploySwarms("reserve", 1)
	globalSwarmManager.DeploySwarms("forge_product", 1)

	if url := os.Getenv("REDIS_URL"); url != "" {
		if opt, err := redis.ParseURL(url); err == nil {
			redisClient = redis.NewClient(opt)
			go startHeartbeat()
		}
	}

	r := chi.NewRouter()

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})

	r.Get("/", corsMiddleware(handleRoot))
	r.Get("/health", corsMiddleware(handleHealth))

	// Stacking Dashboard
	r.Get("/stacking/dashboard", corsMiddleware(handleStackingDashboard))

	// Pillar 1: Enterprise AaaS
	r.Post("/stacking/enterprise/start", corsMiddleware(handleStackingEnterpriseStart))
	r.Post("/stacking/enterprise/onboard", corsMiddleware(handleStackingEnterpriseOnboard))

	// Pillar 2: Autonomous Affiliate Swarm
	r.Post("/stacking/affiliate/start", corsMiddleware(handleStackingAffiliateStart))

	// Pillar 3: Algorithmic Trading Pool
	r.Post("/stacking/trading/start", corsMiddleware(handleStackingTradingStart))
	r.Post("/trading/profit", corsMiddleware(handleTradingProfit))

	// Pillar 4: Digital Product Marketplace
	r.Post("/stacking/products/start", corsMiddleware(handleStackingProductsStart))

	// Pillar 5: Revenue-Share Licensing
	r.Post("/stacking/license/start", corsMiddleware(handleStackingLicenseStart))

	// Infrastructure & Compliance (Isolated)
	r.Get("/financial/status", corsMiddleware(handleFinancialStatus))
	r.Get("/financial/history", corsMiddleware(handleFinancialHistory))
	r.Post("/ai/chat", corsMiddleware(handleAIChat))
	r.Post("/compliance/kill-switch", corsMiddleware(handleComplianceKillSwitch))
	r.Post("/compliance/kill-switch/reset", corsMiddleware(handleComplianceKillSwitchReset))

	r.Post("/tools/execute", corsMiddleware(tools.HandleExecute))

	log.Printf("starting Stacking Fund server on 0.0.0.0:%s", port)
	http.ListenAndServe("0.0.0.0:"+port, r)
}

// --- Stacking Handlers ---

func handleStackingDashboard(w http.ResponseWriter, r *http.Request) {
	status := fundManager.GetStatus()
	history := fundManager.GetHistory(30)

	revenueByPillar := make(map[string]float64)
	for _, tx := range history {
		if tx.Type == "revenue" {
			if strings.HasPrefix(tx.Description, "Revenue from pillar ") {
				parts := strings.Split(tx.Description, ":")
				pillarPart := strings.TrimPrefix(parts[0], "Revenue from pillar ")
				revenueByPillar[pillarPart] += tx.Amount
			}
		}
	}

	currentMonthlyRevenue := 0.0
	for _, rev := range revenueByPillar {
		currentMonthlyRevenue += rev
	}
	if currentMonthlyRevenue == 0 {
		currentMonthlyRevenue = 5000.0 // Default starting baseline
	}

	targetMonthly := 100000.0

	calcTimeToTarget := func(annualRate float64) string {
		monthlyRate := annualRate / 12.0
		if currentMonthlyRevenue >= targetMonthly {
			return "Already reached"
		}
		months := 0
		projected := currentMonthlyRevenue
		for projected < targetMonthly && months < 120 {
			projected *= (1 + monthlyRate)
			months++
		}
		targetDate := time.Now().AddDate(0, months, 0)
		return targetDate.Format("January 2006")
	}

	res := map[string]interface{}{
		"total_accumulated_wealth": status.TotalEarned,
		"current_balance":          status.Balance,
		"monthly_revenue_per_pillar": revenueByPillar,
		"projections": map[string]interface{}{
			"conservative_8_pct": calcTimeToTarget(0.08),
			"aggressive_25_pct": calcTimeToTarget(0.25),
		},
		"time_to_target_100k": calcTimeToTarget(0.25),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func handleStackingEnterpriseStart(w http.ResponseWriter, r *http.Request) {
	var req struct{ Task string }; json.NewDecoder(r.Body).Decode(&req)
	res, err := globalSwarmManager.DispatchTask("vault", req.Task)
	if err != nil { http.Error(w, err.Error(), 500); return }
	json.NewEncoder(w).Encode(res)
}

func handleStackingEnterpriseOnboard(w http.ResponseWriter, r *http.Request) {
	var req struct{ ClientName string; Plan string }; json.NewDecoder(r.Body).Decode(&req)
	amount := 5000.0; if req.Plan == "premium" { amount = 15000.0 }
	fundManager.RecordRevenue(amount, "Enterprise Onboarding: "+req.ClientName, "enterprise")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "onboarded", "revenue": amount})
}

func handleStackingAffiliateStart(w http.ResponseWriter, r *http.Request) {
	var req struct{ Niche string; Task string }; json.NewDecoder(r.Body).Decode(&req)
	res, err := globalSwarmManager.DispatchTask("compound", req.Task)
	if err != nil { http.Error(w, err.Error(), 500); return }
	revenue := 150.0; fundManager.RecordRevenue(revenue, "Affiliate Conversion: "+req.Niche, "affiliate")
	json.NewEncoder(w).Encode(map[string]interface{}{"agent_result": res, "revenue": revenue})
}

func handleStackingTradingStart(w http.ResponseWriter, r *http.Request) {
	var req struct{ Strategy string; Task string }; json.NewDecoder(r.Body).Decode(&req)
	res, err := globalSwarmManager.DispatchTask("reserve", req.Task)
	if err != nil { http.Error(w, err.Error(), 500); return }
	profit := 500.0; fundManager.RecordRevenue(profit, "Trading Strategy: "+req.Strategy, "trading")
	json.NewEncoder(w).Encode(map[string]interface{}{"agent_result": res, "profit": profit})
}

func handleTradingProfit(w http.ResponseWriter, r *http.Request) {
	var req struct{ Profit float64 }; json.NewDecoder(r.Body).Decode(&req)
	fundManager.RecordRevenue(req.Profit, "External Trading Profit", "trading")
	w.WriteHeader(http.StatusOK)
}

func handleStackingProductsStart(w http.ResponseWriter, r *http.Request) {
	var req struct{ ProductType string; Task string }; json.NewDecoder(r.Body).Decode(&req)
	res, err := globalSwarmManager.DispatchTask("forge_product", req.Task)
	if err != nil { http.Error(w, err.Error(), 500); return }
	revenue := 49.0; fundManager.RecordRevenue(revenue, "Digital Product Sale: "+req.ProductType, "products")
	json.NewEncoder(w).Encode(map[string]interface{}{"agent_result": res, "revenue": revenue})
}

func handleStackingLicenseStart(w http.ResponseWriter, r *http.Request) {
	var req struct{ Operator string; Task string }; json.NewDecoder(r.Body).Decode(&req)
	res, err := globalSwarmManager.DispatchTask("vault", req.Task)
	if err != nil { http.Error(w, err.Error(), 500); return }
	revShare := 2000.0; fundManager.RecordRevenue(revShare, "Revenue Share: "+req.Operator, "license")
	json.NewEncoder(w).Encode(map[string]interface{}{"agent_result": res, "revenue_share": revShare})
}

// --- Infrastructure Handlers ---

func handleFinancialStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(fundManager.GetStatus())
}

func handleFinancialHistory(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(fundManager.GetHistory(30))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"status":"ok"}`))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboardHTML))
}

func handleAIChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest; json.NewDecoder(r.Body).Decode(&req)
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" { http.Error(w, "no key", 500); return }
	if !rateLimit() { http.Error(w, "limited", 429); return }
	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "You are the Stacking Fund Wealth Engine Orchestrator."},
			{"role": "user", "content": req.Prompt},
		},
	}
	dsBody, _ := json.Marshal(dsReq)
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	hReq.Header.Set("Authorization", "Bearer "+apiKey)
	hReq.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(hReq)
	if err != nil { http.Error(w, "api failed", 500); return }
	defer resp.Body.Close()
	var dsRes struct {
		Choices []struct{ Message struct{ Content string } }
		Usage struct{ TotalTokens int }
	}
	json.NewDecoder(resp.Body).Decode(&dsRes)
	LogUsage(dsRes.Usage.TotalTokens)
	json.NewEncoder(w).Encode(ChatResponse{Response: dsRes.Choices[0].Message.Content, TokensUsed: dsRes.Usage.TotalTokens})
}

func handleComplianceKillSwitch(w http.ResponseWriter, r *http.Request) {
	os.WriteFile(killSwitchPath, []byte("active"), 0644); w.Write([]byte("Active"))
}
func handleComplianceKillSwitchReset(w http.ResponseWriter, r *http.Request) {
	os.Remove(killSwitchPath); w.Write([]byte("Reset"))
}

// --- Helpers ---

func AddAuditEntry(action string, details map[string]interface{}) {
	auditMutex.Lock(); defer auditMutex.Unlock()
	lastHash := "0000000000000000000000000000000000000000000000000000000000000000"
	// Simplified audit chain for the wealth engine
	entry := AuditEntry{time.Now().Format(time.RFC3339), action, details, ""}
	entryJSON, _ := json.Marshal(entry)
	h := sha256.New(); h.Write([]byte(lastHash + string(entryJSON)))
	entry.Hash = hex.EncodeToString(h.Sum(nil))
	if f, err := os.OpenFile(auditPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		json.NewEncoder(f).Encode(entry); f.Close()
	}
}

func LogUsage(tokens int) {
	usageMutex.Lock(); defer usageMutex.Unlock()
	cost := float64(tokens) * 0.000002
	logEntry := UsageLog{time.Now().Format(time.RFC3339), tokens, cost}
	if f, err := os.OpenFile(usagePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		json.NewEncoder(f).Encode(logEntry); f.Close()
	}
}

func rateLimit() bool {
	rlMu.Lock(); defer rlMu.Unlock(); now := time.Now(); elapsed := now.Sub(rlLastUpdate).Seconds()
	rlLastUpdate = now; rlBucket += elapsed * rlRate; if rlBucket > rlMaxBucket { rlBucket = rlMaxBucket }
	if rlBucket >= 1.0 { rlBucket -= 1.0; return true }; return false
}

func startHeartbeat() {
	for {
		ctx := context.Background()
		nodeData := SwarmNode{ID: nodeID, Region: region, Endpoint: "https://koola10-stacking.fly.dev", Status: "healthy"}
		jsonNode, _ := json.Marshal(nodeData)
		redisClient.Set(ctx, "swarm:node:stacking:"+nodeID, jsonNode, 30*time.Second)
		time.Sleep(15 * time.Second)
	}
}

func generateID() string {
	b := make([]byte, 8); rand.Read(b); return hex.EncodeToString(b)
}
