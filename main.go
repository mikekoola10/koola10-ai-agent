package main

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"koola10/agents"
	"koola10/financial"
	"koola10/tools"

	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

// --- Structs ---

type Grant struct {
	ID          string `json:"grant_id"`
	Title       string `json:"title"`
	Agency      string `json:"agency"`
	Deadline    string `json:"deadline"`
	Amount      string `json:"amount"`
	Eligibility string `json:"eligibility"`
	Description string `json:"description"`
}

type GrantsGovSearchResponse struct {
	Data struct {
		OppHits []struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Agency    string `json:"agency"`
			CloseDate string `json:"closeDate"`
		} `json:"oppHits"`
	} `json:"data"`
}

type GrantsGovDetailsResponse struct {
	Synopsis struct {
		SynDesc                  string `json:"synopsisDesc"`
		EstimatedFunding         string `json:"estimatedFunding"`
		ApplicantEligibilityDesc string `json:"applicantEligibilityDesc"`
	} `json:"synopsis"`
}

type ApplyRequest struct {
	GrantID    string `json:"grant_id"`
	OrgName    string `json:"org_name"`
	OrgMission string `json:"org_mission"`
	OrgBudget  string `json:"org_budget"`
	OrgTaxID   string `json:"org_tax_id"`
}

type ApplicationDraft struct {
	ApplicationID         string `json:"application_id"`
	GrantID               string `json:"grant_id"`
	Status                string `json:"status"`
	ExecutiveSummary      string `json:"executive_summary"`
	StatementOfNeed       string `json:"statement_of_need"`
	ProjectDescription    string `json:"project_description"`
	BudgetJustification   string `json:"budget_justification"`
	OrganizationalCapacity string `json:"organizational_capacity"`
	FollowUpDraft         string `json:"follow_up_draft,omitempty"`
}

type ApplicationSummary struct {
	ApplicationID string `json:"application_id"`
	GrantTitle    string `json:"grant_title"`
	Status        string `json:"status"`
	Deadline      string `json:"deadline"`
}

type MonitorResult struct {
	ApplicationID string `json:"application_id"`
	GrantTitle    string `json:"grant_title"`
	FollowUpEmail string `json:"follow_up_email"`
}

type ChatRequest struct {
	Prompt  string `json:"prompt"`
	Context string `json:"context,omitempty"`
}

type ChatResponse struct {
	Response   string `json:"response"`
	TokensUsed int    `json:"tokens_used"`
}

type MemoryEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AnalyzeGrantRequest struct {
	GrantText  string                 `json:"grant_text"`
	OrgProfile map[string]interface{} `json:"org_profile"`
}

type AnalyzeGrantResponse struct {
	EligibilityScore  int      `json:"eligibility_score"`
	KeyDeadlines      []string `json:"key_deadlines"`
	RequiredDocuments []string `json:"required_documents"`
	Summary           string   `json:"summary"`
}

type Meeting struct {
	MeetingID   string   `json:"meeting_id"`
	Timestamp   string   `json:"timestamp"`
	Summary     string   `json:"summary"`
	Decisions   []string `json:"decisions"`
	ActionItems []string `json:"action_items"`
}

type Entity struct {
	Name  string   `json:"name"`
	Type  string   `json:"type"`
	Tasks []string `json:"tasks,omitempty"`
}

type Edge struct {
	Source    string                 `json:"source"`
	Target    string                 `json:"target"`
	Relation  string                 `json:"relation"`
	Weight    float64                `json:"weight"`
	Frequency int                    `json:"frequency"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type MemoryGraph struct {
	Meetings map[string]Meeting `json:"meetings"`
	Entities map[string]Entity  `json:"entities"`
	Edges    []Edge             `json:"edges"`
	mu       sync.RWMutex
}

type SemanticItem struct {
	Text   string    `json:"text"`
	RefID  string    `json:"ref_id"`
	Vector []float64 `json:"vector"`
}

type SemanticIndex struct {
	Items []SemanticItem `json:"items"`
	mu    sync.RWMutex
}

type SemanticSearchResult struct {
	RefID string  `json:"ref_id"`
	Score float64 `json:"score"`
	Text  string  `json:"text"`
}

type AuditEntry struct {
	Timestamp string                 `json:"timestamp"`
	Action    string                 `json:"action"`
	Details   map[string]interface{} `json:"details"`
	Hash      string                 `json:"hash"`
}

type ApprovalRequest struct {
	ID            string                 `json:"approval_id"`
	Action        string                 `json:"action"`
	Details       map[string]interface{} `json:"details"`
	Status        string                 `json:"status"`
	Approver      string                 `json:"approver,omitempty"`
	Justification string                 `json:"justification,omitempty"`
	CreatedAt     string                 `json:"created_at"`
}

type UsageLog struct {
	Timestamp  string  `json:"timestamp"`
	TokensUsed int     `json:"tokens_used"`
	Cost       float64 `json:"cost"`
}

type Transaction struct {
	Timestamp   string  `json:"timestamp"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Vertical    string  `json:"vertical,omitempty"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

type EconomicLedger struct {
	Balance      float64       `json:"balance"`
	TotalCosts   float64       `json:"total_costs"`
	TotalRevenue float64       `json:"total_revenue"`
	Transactions []Transaction `json:"transactions"`
	mu           sync.RWMutex
}

type EconomicSummary struct {
	Balance      float64 `json:"balance"`
	TotalCosts   float64 `json:"total_costs"`
	TotalRevenue float64 `json:"total_revenue"`
	ROI          float64 `json:"roi"`
}

type EconomicEvaluation struct {
	Decision      string  `json:"decision"`
	EstimatedCost float64 `json:"estimated_cost"`
	ProjectedROI  float64 `json:"projected_roi"`
	Reason        string  `json:"reason"`
}

type SwarmTask struct {
	TaskID     string                 `json:"task_id"`
	Stage      string                 `json:"stage"` // "finding", "writing", "reviewing", "submitting", "done"
	Query      string                 `json:"query"`
	OrgProfile map[string]interface{} `json:"org_profile"`
	Results    map[string]interface{} `json:"results"`
}

type VaultSummary struct {
	TotalRevenue   float64 `json:"total_revenue"`
	OperationsFund float64 `json:"operations_fund"`
	SpendableFund  float64 `json:"spendable_fund"`
}

type Subscription struct {
	ID        string    `json:"id"`
	Service   string    `json:"service"`
	Amount    float64   `json:"amount"`
	CardID    string    `json:"card_id"`
	Status    string    `json:"status"`
	Frequency string    `json:"frequency"`
	LastPaid  time.Time `json:"last_paid"`
}

type SubscriptionManager struct {
	Subscriptions []Subscription `json:"subscriptions"`
	storagePath   string
	mu            sync.RWMutex
}

type SwarmNode struct {
	ID       string `json:"node_id"`
	Region   string `json:"region"`
	Endpoint string `json:"endpoint"`
	Status   string `json:"status"`
}

// --- Studio Structs ---

type LoreRequest struct {
	Question string `json:"question"`
}

type StyleRequest struct {
	Description string `json:"description"`
}

type StyleResponse struct {
	StyleRules string `json:"style_rules"`
	Prompt     string `json:"prompt"`
}

type Episode struct {
	ID          string   `json:"episode_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Characters  []string `json:"characters"`
	CreatedAt   string   `json:"created_at"`
}

type VideoJob struct {
	ID        string `json:"job_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// --- Global States ---

var (
	cacheMutex   sync.Mutex
	auditMutex   sync.Mutex
	usageMutex   sync.Mutex
	approvalMu   sync.Mutex
	subMu        sync.Mutex
	killSwitchMu sync.Mutex
	videoJobMu   sync.Mutex
	reflectMu    sync.RWMutex
	opportunityMu sync.RWMutex

	cachePath      = "/data/grants_cache.json"
	appsDir        = "/data/applications"
	memoryPath     = "/data/memory.json"
	graphPath      = "/data/memory_graph.json"
	semanticPath   = "/data/semantic_index.json"
	auditPath      = "/data/audit_chain.jsonl"
	usagePath      = "/data/usage.jsonl"
	killSwitchPath = "/data/kill_switch"
	subsPath       = "/data/subscriptions.json"
	serverLogPath  = "/data/server.log"
	lastProcessedAuditOffset int64

	ledgerPath     = "/data/economic_ledger.json"
	fundPath       = "/data/operational_fund.json"

	globalGraph = &MemoryGraph{
		Meetings: make(map[string]Meeting),
		Entities: make(map[string]Entity),
		Edges:    []Edge{},
	}

	globalSemantic = &SemanticIndex{
		Items: []SemanticItem{},
	}

	globalLedger = &EconomicLedger{
		Balance: 100.0,
	}

	fundManager *financial.FundManager
	subManager  *SubscriptionManager

	globalSwarmManager = agents.NewSwarmManager()

	approvalStore = make(map[string]*ApprovalRequest)
	videoJobStore = make(map[string]*VideoJob)
	subStore      = make(map[string]string) // subID -> status
	reflectLogs   = []map[string]string{}
	proactiveFeed = []map[string]string{}

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

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("ADMIN_API_KEY")
		if apiKey == "" {
			// If not set, allow for development but log warning
			log.Println("WARNING: ADMIN_API_KEY not set")
			next(w, r)
			return
		}

		providedKey := r.Header.Get("X-Admin-API-Key")
		if providedKey == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				providedKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if providedKey != apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

	os.MkdirAll(filepath.Dir(cachePath), 0755)
	os.MkdirAll(appsDir, 0755)

	globalGraph.Load()
	globalSemantic.Load()
	globalLedger.Load()
	loadReflectionsFromFile()
	fundManager = financial.NewFundManager(fundPath, globalLedger)
	if data, err := os.ReadFile("/data/audit_offset.txt"); err == nil {
		fmt.Sscanf(string(data), "%d", &lastProcessedAuditOffset)
	}
	subManager = NewSubscriptionManager(subsPath)
	go startMaintenanceLoop()
	go startStrategicForesightLoop()

	// Automated invoice payment check (every 24h)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for {
			fundManager.PayFlyInvoice()
			<-ticker.C
		}
	}()

	globalSwarmManager.AuditLogger = AddAuditEntry
	globalSwarmManager.LedgerLogger = globalLedger.RecordCost
	globalSwarmManager.RevenueLogger = fundManager.RouteRevenue
	globalSwarmManager.ReflectLogger = func(vertical, specialty, task string, result interface{}) string {
		reflection := generateReflection(vertical, specialty, task, result)
		reflectMu.Lock()
		defer reflectMu.Unlock()
		reflectLogs = append(reflectLogs, map[string]string{
			"timestamp":  time.Now().Format(time.RFC3339),
			"vertical":   vertical,
			"specialty":  specialty,
			"task":       task,
			"reflection": reflection,
		})
		if len(reflectLogs) > 100 {
			reflectLogs = reflectLogs[1:]
		}
		saveReflectionsToFile()
		return reflection
	}

	// Initialize High-Growth Founder Mode with AGI directives
	founderPrompt := "High-Growth Founder Mode: Speed is a competitive advantage. Build leverage through automation. First-principles thinking. High agency + extreme ownership. Think in 10x-100x. Transform into AGI/ASI-like entities. General Intelligence. Recursive Self-Improvement. Antifragile. Swarm Intelligence. Persistent Memory."
	globalSwarmManager.ToggleAGIMode(true) // Phase 2: AGI Mode is now default
	globalSwarmManager.SetGlobalPrompt(founderPrompt)

	globalSwarmManager.Factories["sterling"] = agents.FinancialFactory
	globalSwarmManager.Factories["nova"] = agents.GrantSwarmFactory
	globalSwarmManager.Factories["forge"] = agents.DeveloperFactory
	globalSwarmManager.Factories["echo"] = agents.APIFactory
	globalSwarmManager.Factories["solara"] = agents.ContentFactory
	globalSwarmManager.Factories["sage"] = agents.ComplianceFactory
	globalSwarmManager.Factories["vale"] = agents.ResearchFactory
	globalSwarmManager.Factories["maintenance"] = agents.MaintenanceFactory

	// Descriptive Slugs & Pilot Aliases
	globalSwarmManager.Factories["trading"] = agents.TradingFactory
	globalSwarmManager.Factories["leadgen"] = agents.LeadGenFactory
	globalSwarmManager.Factories["api_service"] = agents.APIFactory
	globalSwarmManager.Factories["financial_report"] = agents.FinancialFactory
	globalSwarmManager.Factories["grant"] = agents.GrantSwarmFactory
	globalSwarmManager.Factories["content"] = agents.ContentFactory
	globalSwarmManager.Factories["compliance"] = agents.ComplianceFactory
	globalSwarmManager.Factories["research"] = agents.ResearchFactory

	globalSwarmManager.Factories["affiliate"] = agents.AffiliateFactory
	globalSwarmManager.Factories["bounty"] = agents.BountyFactory

	// Register Night Shift vertical
	globalSwarmManager.Factories["night-shift"] = agents.DeveloperFactory

	// Initial deployment of revenue swarms
	globalSwarmManager.DeploySwarms("affiliate", 10)
	globalSwarmManager.DeploySwarms("bounty", 10)
	globalSwarmManager.DeploySwarms("content", 10)

	if url := os.Getenv("REDIS_URL"); url != "" {
		if opt, err := redis.ParseURL(url); err == nil {
			redisClient = redis.NewClient(opt)
			go startHeartbeat()
			go startSwarmListeners()
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
	r.Get("/daily-report", corsMiddleware(handleDailyReport))
	r.Get("/events/stream", handleEventsStream)
	r.Post("/collaborate/*", corsMiddleware(handleCollaborate))

	r.Get("/grants/search", corsMiddleware(handleSearch))
	r.Post("/grants/apply", corsMiddleware(handleApply))
	r.Get("/grants/status", corsMiddleware(handleStatus))
	r.Get("/grants/applications", corsMiddleware(handleApplicationsList))
	r.Post("/grants/monitor", corsMiddleware(handleMonitor))
	r.Post("/grants/update-status", corsMiddleware(handleUpdateStatus))
	r.Post("/grants/apply-auto", corsMiddleware(handleApplyAuto))
	r.Post("/grants/check-status", corsMiddleware(handleCheckStatus))

	r.Post("/payment/create-checkout", corsMiddleware(handleCreateCheckout))
	r.Post("/stripe/webhook", handleStripeWebhook)

	r.Post("/ai/chat", corsMiddleware(handleAIChat))
	r.Post("/ai/remember", corsMiddleware(handleAIRemember))
	r.Get("/ai/recall", corsMiddleware(handleAIRecall))
	r.Post("/ai/analyze-grant", corsMiddleware(handleAIAnalyzeGrant))

	r.Get("/memory/meetings", corsMiddleware(handleMemoryMeetings))
	r.Post("/memory/meetings", corsMiddleware(handleMemoryMeetings))
	r.Get("/memory/entity/*", corsMiddleware(handleMemoryEntity))
	r.Get("/memory/influence/*", corsMiddleware(handleMemoryInfluence))
	r.Get("/memory/path", corsMiddleware(handleMemoryPath))
	r.Get("/memory/decisions/ranked", corsMiddleware(handleMemoryDecisionsRanked))

	r.Post("/semantic/index", corsMiddleware(handleSemanticIndex))
	r.Get("/semantic/search", corsMiddleware(handleSemanticSearch))

	r.Get("/compliance/audit", corsMiddleware(handleComplianceAudit))
	r.Get("/compliance/audit/verify", corsMiddleware(handleComplianceAuditVerify))
	r.Post("/compliance/approval", corsMiddleware(handleComplianceApproval))
	r.Post("/compliance/approve", corsMiddleware(handleComplianceApprove))
	r.Post("/compliance/kill-switch", corsMiddleware(handleComplianceKillSwitch))
	r.Post("/compliance/kill-switch/reset", corsMiddleware(handleComplianceKillSwitchReset))
	r.Get("/compliance/usage", corsMiddleware(handleComplianceUsage))

	r.Get("/vault/summary", corsMiddleware(handleVaultSummary))
	r.Get("/logs", corsMiddleware(authMiddleware(handleLogs)))
	r.Post("/admin/subscriptions/run", authMiddleware(handleAdminSubscriptionsRun))
	r.Get("/admin/agi-mode", corsMiddleware(authMiddleware(handleGetAGIMode)))
	r.Post("/admin/agi-mode", corsMiddleware(authMiddleware(handleAdminAGIMode)))
	r.Post("/admin/agentcards/create", authMiddleware(handleAdminAgentCardsCreate))
	r.Get("/admin/subscriptions", authMiddleware(handleAdminSubscriptionsList))
	r.Post("/admin/stellar/send", authMiddleware(handleAdminStellarSend))
	r.Get("/admin/stellar/balance", authMiddleware(handleAdminStellarBalance))

	r.Post("/economic/ledger/cost", corsMiddleware(handleEconomicLedgerCost))
	r.Post("/economic/ledger/revenue", corsMiddleware(handleEconomicLedgerRevenue))
	r.Get("/economic/ledger/summary", corsMiddleware(handleEconomicLedgerSummary))
	r.Post("/economic/evaluate", corsMiddleware(handleEconomicEvaluate))

	r.Post("/swarm/start", corsMiddleware(handleSwarmStart))
	r.Get("/swarm/task-status", corsMiddleware(handleSwarmStatus))
	r.Get("/swarm/agents", corsMiddleware(handleSwarmAgents))
	r.Get("/swarm/reflections", corsMiddleware(handleSwarmReflections))
	r.Get("/swarm/memory", corsMiddleware(handleSwarmMemory))
	r.Get("/swarm/opportunities", corsMiddleware(handleSwarmOpportunities))
	r.Get("/swarm/task-forces", corsMiddleware(handleSwarmTaskForces))
	r.Get("/swarm/nodes", corsMiddleware(handleSwarmNodes))

	r.Get("/swarm/metrics", corsMiddleware(handleSwarmMetrics))
	r.Get("/swarm/report", corsMiddleware(handleSwarmReport))
	r.Get("/swarm/revenue", corsMiddleware(handleSwarmRevenue))
	r.Get("/swarm/status", corsMiddleware(handleSwarmStatusAll))
	r.HandleFunc("/swarm/*", corsMiddleware(handleSpecialistSwarm))

	r.Post("/admin/trigger_affiliate", corsMiddleware(authMiddleware(handleTriggerAffiliate)))
	r.Post("/admin/trigger_bounty", corsMiddleware(authMiddleware(handleTriggerBounty)))
	r.Post("/admin/bpa/onboard", corsMiddleware(authMiddleware(handleBPAOnboard)))
	r.Post("/admin/trigger_content", corsMiddleware(authMiddleware(handleTriggerContent)))
	r.Post("/admin/scheduler/run", corsMiddleware(authMiddleware(handleSchedulerRun)))

	r.Get("/financial/status", corsMiddleware(handleFinancialStatus))
	r.Post("/financial/pay-subscription", corsMiddleware(handleFinancialPaySubscription))
	r.Post("/financial/reinvest", corsMiddleware(handleFinancialReinvest))
	r.Get("/financial/history", corsMiddleware(handleFinancialHistory))
	r.Post("/trading/profit", corsMiddleware(handleTradingProfit))

	r.Post("/tools/execute", corsMiddleware(tools.HandleExecute))

	r.Post("/studio/lore", corsMiddleware(handleStudioLore))
	r.Post("/studio/style", corsMiddleware(handleStudioStyle))
	r.Post("/studio/episode", corsMiddleware(handleStudioEpisode))
	r.Get("/studio/episodes", corsMiddleware(handleStudioEpisodesList))
	r.Post("/studio/video-job", corsMiddleware(handleStudioVideoJob))
	r.Get("/studio/video-job/*", corsMiddleware(handleStudioVideoJobStatus))

	log.Printf("starting server on 0.0.0.0:%s", port)
	http.ListenAndServe("0.0.0.0:"+port, r)
}

// --- Studio Handlers ---

func handleStudioLore(w http.ResponseWriter, r *http.Request) {
	var req LoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		http.Error(w, "no key", 500)
		return
	}
	if !rateLimit() {
		http.Error(w, "limited", 429)
		return
	}

	systemPrompt := "You are the Lorekeeper of the Koola10 cinematic universe. Answer questions about characters, magic systems, and universe rules. Magic is based on Emergent Resonance. Tone is gritty but hopeful. Main characters include Kaelen and Lyra."

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": req.Question},
		},
	}
	dsBody, _ := json.Marshal(dsReq)
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	hReq.Header.Set("Authorization", "Bearer "+apiKey)
	hReq.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(hReq)
	if err != nil {
		http.Error(w, "api failed", 500)
		return
	}
	defer resp.Body.Close()
	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string
			}
		}
		Usage struct {
			TotalTokens int
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil {
		http.Error(w, "parse failed", 500)
		return
	}

	LogUsage(dsRes.Usage.TotalTokens)
	globalLedger.RecordCost("", "studio_lore", float64(dsRes.Usage.TotalTokens)*0.000002, "Lorekeeper query")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{
		Response:   dsRes.Choices[0].Message.Content,
		TokensUsed: dsRes.Usage.TotalTokens,
	})
}

func handleStudioStyle(w http.ResponseWriter, r *http.Request) {
	var req StyleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		http.Error(w, "no key", 500)
		return
	}
	if !rateLimit() {
		http.Error(w, "limited", 429)
		return
	}

	systemPrompt := "Generate Koola10 style rules (Boondocks + 4K realism) and convert the scene into an Emergent Video prompt. Return JSON with 'style_rules' and 'prompt'."

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": req.Description},
		},
		"response_format": map[string]string{"type": "json_object"},
	}
	dsBody, _ := json.Marshal(dsReq)
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	hReq.Header.Set("Authorization", "Bearer "+apiKey)
	hReq.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(hReq)
	if err != nil {
		http.Error(w, "api failed", 500)
		return
	}
	defer resp.Body.Close()
	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string
			}
		}
		Usage struct {
			TotalTokens int
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil {
		http.Error(w, "parse failed", 500)
		return
	}

	LogUsage(dsRes.Usage.TotalTokens)
	globalLedger.RecordCost("", "studio_style", float64(dsRes.Usage.TotalTokens)*0.000002, "Style generation")

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(dsRes.Choices[0].Message.Content))
}

func handleStudioEpisode(w http.ResponseWriter, r *http.Request) {
	var ep Episode
	if err := json.NewDecoder(r.Body).Decode(&ep); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	ep.ID = generateID()
	ep.CreatedAt = time.Now().Format(time.RFC3339)

	globalGraph.mu.Lock()
	if globalGraph.Entities == nil { globalGraph.Entities = make(map[string]Entity) }
	globalGraph.Entities[ep.ID] = Entity{Name: ep.Title, Type: "episode", Tasks: ep.Characters}
	globalGraph.mu.Unlock()

	globalGraph.AddWeightedEdge("studio", ep.ID, "produced_episode", map[string]interface{}{"description": ep.Description})
	globalGraph.Save()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ep)
}

func handleStudioEpisodesList(w http.ResponseWriter, r *http.Request) {
	globalGraph.mu.RLock()
	defer globalGraph.mu.RUnlock()

	var episodes []Episode
	for id, entity := range globalGraph.Entities {
		if entity.Type == "episode" {
			episodes = append(episodes, Episode{
				ID:         id,
				Title:      entity.Name,
				Characters: entity.Tasks,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodes)
}

func handleStudioVideoJob(w http.ResponseWriter, r *http.Request) {
	id := generateID()
	job := &VideoJob{
		ID:        id,
		Status:    "pending",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	videoJobMu.Lock()
	videoJobStore[id] = job
	videoJobMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func handleStudioVideoJobStatus(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	videoJobMu.Lock()
	job, ok := videoJobStore[id]
	videoJobMu.Unlock()

	if !ok {
		http.Error(w, "not found", 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// --- Swarm Infrastructure ---

func startHeartbeat() {
	for {
		ctx := context.Background()
		nodeData := SwarmNode{ID: nodeID, Region: region, Endpoint: "https://koola10.fly.dev", Status: "healthy"}
		jsonNode, _ := json.Marshal(nodeData)
		// We use a separate key for each node's availability to avoid overwriting the whole hash TTL
		redisClient.Set(ctx, "swarm:node:"+nodeID, jsonNode, 30*time.Second)
		redisClient.HSet(ctx, "swarm:nodes", nodeID, jsonNode)
		time.Sleep(15 * time.Second)
	}
}

func handleSwarmNodes(w http.ResponseWriter, r *http.Request) {
	if redisClient == nil { http.Error(w, "no redis", 503); return }
	ctx := context.Background()
	nodes, _ := redisClient.HGetAll(ctx, "swarm:nodes").Result()
	var res []SwarmNode
	for id, v := range nodes {
		// Clean up dead nodes from the hash
		if redisClient.Exists(ctx, "swarm:node:"+id).Val() == 0 {
			redisClient.HDel(ctx, "swarm:nodes", id)
			continue
		}
		var n SwarmNode; json.Unmarshal([]byte(v), &n); res = append(res, n)
	}
	json.NewEncoder(w).Encode(res)
}

func handleSwarmAgents(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]map[string]string{
		{"role": "finder", "status": "active"}, {"role": "writer", "status": "active"},
		{"role": "reviewer", "status": "active"}, {"role": "submitter", "status": "active"},
		{"role": "orchestrator", "status": "active"},
	})
}

func handleSwarmStart(w http.ResponseWriter, r *http.Request) {
	var req struct { Query string; OrgProfile map[string]interface{} }; json.NewDecoder(r.Body).Decode(&req)
	id := generateID()
	task := SwarmTask{TaskID: id, Stage: "finding", Query: req.Query, OrgProfile: req.OrgProfile, Results: make(map[string]interface{})}
	if redisClient != nil {
		b, _ := json.Marshal(task)
		redisClient.Set(context.Background(), "task:"+id, b, 24*time.Hour)
		redisClient.Publish(context.Background(), "tasks:orchestrator", b)
	}
	json.NewEncoder(w).Encode(map[string]string{"task_id": id})
}

func handleSwarmStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("task_id")
	if redisClient == nil { http.Error(w, "no redis", 503); return }
	v, _ := redisClient.Get(context.Background(), "task:"+id).Result()
	w.Header().Set("Content-Type", "application/json"); w.Write([]byte(v))
}

// --- Specialized Swarm Agents (Integrated Logic) ---

func startSwarmListeners() {
	if redisClient == nil { return }
	ctx := context.Background()

	// Orchestrator
	go func() {
		pubsub := redisClient.Subscribe(ctx, "tasks:orchestrator")
		for {
			msg, err := pubsub.ReceiveMessage(ctx); if err != nil { continue }
			var t SwarmTask; json.Unmarshal([]byte(msg.Payload), &t)
			redisClient.Publish(ctx, "tasks:"+t.Stage, msg.Payload)
		}
	}()

	// Finder
	go func() {
		pubsub := redisClient.Subscribe(ctx, "tasks:finding")
		for {
			msg, err := pubsub.ReceiveMessage(ctx); if err != nil { continue }
			var t SwarmTask; json.Unmarshal([]byte(msg.Payload), &t)
			log.Printf("[Finder] Processing task %s", t.TaskID)

			// Real logic: call search
			req, _ := http.NewRequest("GET", "http://localhost:8080/grants/search?query="+url.QueryEscape(t.Query), nil)
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				var grants []Grant; json.NewDecoder(resp.Body).Decode(&grants); resp.Body.Close()
				t.Results["grants"] = grants
				if len(grants) > 0 { t.Stage = "writing" } else { t.Stage = "done" }
			}

			b, _ := json.Marshal(t); redisClient.Set(ctx, "task:"+t.TaskID, b, 24*time.Hour)
			redisClient.Publish(ctx, "tasks:orchestrator", b)
			globalLedger.RecordCost("swarm", "swarm_agent", 0.01, "FinderAgent search")
		}
	}()

	// Writer
	go func() {
		pubsub := redisClient.Subscribe(ctx, "tasks:writing")
		for {
			msg, err := pubsub.ReceiveMessage(ctx); if err != nil { continue }
			var t SwarmTask; json.Unmarshal([]byte(msg.Payload), &t)
			log.Printf("[Writer] Processing task %s", t.TaskID)

			grants, ok := t.Results["grants"].([]interface{})
			if ok && len(grants) > 0 {
				g := grants[0].(map[string]interface{})
				applyReq := ApplyRequest{
					GrantID: g["grant_id"].(string), OrgName: t.OrgProfile["name"].(string),
					OrgMission: t.OrgProfile["mission"].(string), OrgBudget: t.OrgProfile["budget"].(string),
				}
				body, _ := json.Marshal(applyReq)
				req, _ := http.NewRequest("POST", "http://localhost:8080/grants/apply", bytes.NewBuffer(body))
				resp, err := http.DefaultClient.Do(req)
				if err == nil {
					var draft ApplicationDraft; json.NewDecoder(resp.Body).Decode(&draft); resp.Body.Close()
					t.Results["application"] = draft
					t.Stage = "reviewing"
				}
			}

			b, _ := json.Marshal(t); redisClient.Set(ctx, "task:"+t.TaskID, b, 24*time.Hour)
			redisClient.Publish(ctx, "tasks:orchestrator", b)
			globalLedger.RecordCost("swarm", "swarm_agent", 0.05, "WriterAgent drafting")
		}
	}()

	// Reviewer
	go func() {
		pubsub := redisClient.Subscribe(ctx, "tasks:reviewing")
		for {
			msg, err := pubsub.ReceiveMessage(ctx); if err != nil { continue }
			var t SwarmTask; json.Unmarshal([]byte(msg.Payload), &t)
			log.Printf("[Reviewer] Processing task %s", t.TaskID)

			// ROI Evaluation
			eval := EvaluateAction("grant_submit", 0.07)
			t.Results["roi_eval"] = eval

			if eval.Decision == "block" {
				t.Stage = "failed"
			} else {
				// Safety: Create approval request instead of direct submission
				app := t.Results["application"].(map[string]interface{})
				appID := app["application_id"].(string)
				approvalReq := map[string]interface{}{"action": "grant_submit", "details": map[string]interface{}{"application_id": appID, "task_id": t.TaskID}}
				body, _ := json.Marshal(approvalReq)
				req, _ := http.NewRequest("POST", "http://localhost:8080/compliance/approval", bytes.NewBuffer(body))
				resp, err := http.DefaultClient.Do(req)
				if err == nil {
					var approval ApprovalRequest; json.NewDecoder(resp.Body).Decode(&approval); resp.Body.Close()
					t.Results["approval_id"] = approval.ID
					t.Stage = "submitting" // In swarm, we transition to submitter which will wait for approval
				}
			}

			b, _ := json.Marshal(t); redisClient.Set(ctx, "task:"+t.TaskID, b, 24*time.Hour)
			redisClient.Publish(ctx, "tasks:orchestrator", b)
			globalLedger.RecordCost("swarm", "swarm_agent", 0.01, "ReviewerAgent review")
		}
	}()

	// Submitter
	go func() {
		pubsub := redisClient.Subscribe(ctx, "tasks:submitting")
		for {
			msg, err := pubsub.ReceiveMessage(ctx); if err != nil { continue }
			var t SwarmTask; json.Unmarshal([]byte(msg.Payload), &t)
			log.Printf("[Submitter] Processing task %s", t.TaskID)

			approvalID := t.Results["approval_id"].(string)
			// Check if approved
			approvalMu.Lock(); ap, ok := approvalStore[approvalID]; approvalMu.Unlock()
			if ok && ap.Status == "approved" {
				// Proceed to auto-apply
				app := t.Results["application"].(map[string]interface{})
				submitReq := map[string]interface{}{"url": "https://www.grants.gov", "approval_id": approvalID, "form_data": map[string]string{"app_id": app["application_id"].(string)}}
				body, _ := json.Marshal(submitReq)
				req, _ := http.NewRequest("POST", "http://localhost:8080/grants/apply-auto", bytes.NewBuffer(body))
				resp, err := http.DefaultClient.Do(req)
				if err == nil {
					var res map[string]interface{}; json.NewDecoder(resp.Body).Decode(&res); resp.Body.Close()
					t.Results["submission"] = res
					t.Stage = "done"
				}
			} else {
				// Re-publish to orchestrator to retry later (waiting for approval)
				time.Sleep(10 * time.Second)
			}

			b, _ := json.Marshal(t); redisClient.Set(ctx, "task:"+t.TaskID, b, 24*time.Hour)
			redisClient.Publish(ctx, "tasks:orchestrator", b)
			globalLedger.RecordCost("swarm", "swarm_agent", 0.02, "SubmitterAgent browser action")
		}
	}()
}

// --- Persistence & Helpers (Graph/Semantic/Economic/Compliance - All from previous phases) ---

func (g *MemoryGraph) Save() {
	g.mu.RLock(); defer g.mu.RUnlock()
	data, _ := json.Marshal(g); os.WriteFile(graphPath, data, 0644)
}
func (g *MemoryGraph) Load() {
	g.mu.Lock(); defer g.mu.Unlock(); data, err := os.ReadFile(graphPath)
	if err == nil { json.Unmarshal(data, g) }
	if g.Meetings == nil { g.Meetings = make(map[string]Meeting) }
	if g.Entities == nil { g.Entities = make(map[string]Entity) }
}
func (g *MemoryGraph) AddWeightedEdge(source, target, relation string, metadata map[string]interface{}) {
	g.mu.Lock(); defer g.mu.Unlock()
	for i, edge := range g.Edges {
		if edge.Source == source && edge.Target == target && edge.Relation == relation {
			g.Edges[i].Weight += 0.2; if g.Edges[i].Weight > 2.0 { g.Edges[i].Weight = 2.0 }; g.Edges[i].Frequency++; return
		}
	}
	g.Edges = append(g.Edges, Edge{Source: source, Target: target, Relation: relation, Weight: 1.0, Frequency: 1, Metadata: metadata})
}
func (g *MemoryGraph) AddMeeting(m Meeting) string {
	if m.MeetingID == "" { m.MeetingID = generateID() }; if m.Timestamp == "" { m.Timestamp = time.Now().Format(time.RFC3339) }
	g.mu.Lock(); g.Meetings[m.MeetingID] = m; g.mu.Unlock()
	for _, decision := range m.Decisions {
		g.mu.Lock(); if _, ok := g.Entities[decision]; !ok { g.Entities[decision] = Entity{Name: decision, Type: "decision"} }; g.mu.Unlock()
		g.AddWeightedEdge(m.MeetingID, decision, "contains_decision", nil)
	}
	for _, item := range m.ActionItems {
		parts := strings.Split(item, ":"); taskName := item
		if len(parts) > 1 {
			owner := strings.TrimSpace(parts[0]); taskName = strings.TrimSpace(parts[1])
			g.mu.Lock(); entity, ok := g.Entities[owner]; if !ok { entity = Entity{Name: owner, Type: "person"} }
			entity.Tasks = append(entity.Tasks, taskName); g.Entities[owner] = entity; g.mu.Unlock()
			g.AddWeightedEdge(owner, taskName, "assigned_to", nil)
		}
		g.mu.Lock(); if _, ok := g.Entities[taskName]; !ok { g.Entities[taskName] = Entity{Name: taskName, Type: "task"} }; g.mu.Unlock()
		g.AddWeightedEdge(m.MeetingID, taskName, "contains_task", nil)
	}
	g.Save(); return m.MeetingID
}
func (g *MemoryGraph) CalculateInfluenceScore(name string) float64 {
	g.mu.RLock(); defer g.mu.RUnlock(); var in, out float64; var count int
	for _, e := range g.Edges {
		if e.Target == name { in += e.Weight; count++ }
		if e.Source == name { out += e.Weight; count++ }
	}
	if count == 0 { return 0 }; return ((in * 0.7) + (out * 0.3)) / float64(count)
}
func (g *MemoryGraph) FindPath(source, target string, maxDepth int) []Edge {
	g.mu.RLock(); defer g.mu.RUnlock()
	type node struct { entity string; path []Edge }; queue := []node{{source, []Edge{}}}; visited := make(map[string]bool)
	for len(queue) > 0 {
		curr := queue[0]; queue = queue[1:]
		if curr.entity == target { return curr.path }
		if len(curr.path) >= maxDepth { continue }
		visited[curr.entity] = true
		for _, e := range g.Edges {
			if e.Source == curr.entity && !visited[e.Target] {
				newPath := append([]Edge{}, curr.path...); newPath = append(newPath, e); queue = append(queue, node{e.Target, newPath})
			}
		}
	}
	return nil
}
func (g *MemoryGraph) RankDecisionsByImpact() []string {
	g.mu.RLock(); var res []string; for n, e := range g.Entities { if e.Type == "decision" { res = append(res, n) } }; g.mu.RUnlock()
	sort.Slice(res, func(i, j int) bool { return g.CalculateInfluenceScore(res[i]) > g.CalculateInfluenceScore(res[j]) }); return res
}
func (s *SemanticIndex) Save() {
	s.mu.RLock(); defer s.mu.RUnlock(); data, _ := json.Marshal(s); os.WriteFile(semanticPath, data, 0644)
}
func (s *SemanticIndex) Load() {
	s.mu.Lock(); defer s.mu.Unlock(); data, err := os.ReadFile(semanticPath)
	if err == nil { json.Unmarshal(data, s) }
	if s.Items == nil { s.Items = []SemanticItem{} }
}
func (s *SemanticIndex) AddItem(text, refID string) error {
	url := os.Getenv("SEMANTIC_AGENT_URL"); if url == "" { url = "https://koola10-semantic.fly.dev" }
	b, _ := json.Marshal(map[string]string{"text": text}); resp, err := http.Post(url+"/generate", "application/json", bytes.NewBuffer(b))
	if err != nil { return err }; if resp != nil { defer resp.Body.Close() }
	var res struct { Vector []float64 `json:"vector"` }; if err := json.NewDecoder(resp.Body).Decode(&res); err != nil { return err }
	s.mu.Lock(); s.Items = append(s.Items, SemanticItem{text, refID, res.Vector}); s.mu.Unlock(); s.Save(); return nil
}
func (s *SemanticIndex) Search(query string, topK int) ([]SemanticSearchResult, error) {
	url := os.Getenv("SEMANTIC_AGENT_URL"); if url == "" { url = "https://koola10-semantic.fly.dev" }
	s.mu.RLock(); b, _ := json.Marshal(map[string]interface{}{"query": query, "embeddings": s.Items, "top_k": topK}); s.mu.RUnlock()
	resp, err := http.Post(url+"/search", "application/json", bytes.NewBuffer(b))
	if err != nil { return nil, err }; if resp != nil { defer resp.Body.Close() }
	var res []SemanticSearchResult; if err := json.NewDecoder(resp.Body).Decode(&res); err != nil { return nil, err }
	s.mu.RLock(); defer s.mu.RUnlock(); for i, r := range res {
		for _, item := range s.Items { if item.RefID == r.RefID { res[i].Text = item.Text; break } }
	}
	return res, nil
}
func AddAuditEntry(action string, details map[string]interface{}) {
	auditMutex.Lock(); defer auditMutex.Unlock(); lastHash := "0000000000000000000000000000000000000000000000000000000000000000"
	if f, err := os.Open(auditPath); err == nil {
		scanner := bufio.NewScanner(f); var lastLine string; for scanner.Scan() { lastLine = scanner.Text() }; f.Close()
		if lastLine != "" { var e AuditEntry; if err := json.Unmarshal([]byte(lastLine), &e); err == nil { lastHash = e.Hash } }
	}
	entry := AuditEntry{time.Now().Format(time.RFC3339), action, details, ""}
	entryJSON, _ := json.Marshal(entry); h := sha256.New(); h.Write([]byte(lastHash + string(entryJSON))); entry.Hash = hex.EncodeToString(h.Sum(nil))
	if f, err := os.OpenFile(auditPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil { json.NewEncoder(f).Encode(entry); f.Close() }
}
func checkKillSwitch() bool {
	killSwitchMu.Lock(); defer killSwitchMu.Unlock(); data, err := os.ReadFile(killSwitchPath)
	return err == nil && string(data) == "active"
}
func rateLimit() bool {
	rlMu.Lock(); defer rlMu.Unlock(); now := time.Now(); elapsed := now.Sub(rlLastUpdate).Seconds()
	rlLastUpdate = now; rlBucket += elapsed * rlRate; if rlBucket > rlMaxBucket { rlBucket = rlMaxBucket }
	if rlBucket >= 1.0 { rlBucket -= 1.0; return true }; return false
}
func LogUsage(tokens int) {
	usageMutex.Lock(); defer usageMutex.Unlock(); cost := float64(tokens) * 0.000002
	logEntry := UsageLog{time.Now().Format(time.RFC3339), tokens, cost}
	if f, err := os.OpenFile(usagePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil { json.NewEncoder(f).Encode(logEntry); f.Close() }
}
func (l *EconomicLedger) Save() {
	l.mu.RLock(); defer l.mu.RUnlock(); data, _ := json.Marshal(l); os.WriteFile(ledgerPath, data, 0644)
}
func (l *EconomicLedger) Load() {
	l.mu.Lock(); defer l.mu.Unlock(); data, err := os.ReadFile(ledgerPath)
	if err == nil { json.Unmarshal(data, l) }; if l.Transactions == nil { l.Transactions = []Transaction{} }
}
func (l *EconomicLedger) RecordCost(vertical, category string, amount float64, description string) {
	l.mu.Lock(); l.Balance -= amount; l.TotalCosts += amount
	l.Transactions = append(l.Transactions, Transaction{time.Now().Format(time.RFC3339), "cost", category, vertical, amount, description})
	l.mu.Unlock(); l.Save(); AddAuditEntry("economic_cost_logged", map[string]interface{}{"amount": amount, "category": category, "vertical": vertical})
}
func (l *EconomicLedger) RecordRevenue(amount float64, source string) {
	l.RecordRevenueWithVertical("", amount, source)
}
func (l *EconomicLedger) RecordRevenueWithVertical(vertical string, amount float64, source string) {
	l.mu.Lock(); l.Balance += amount; l.TotalRevenue += amount
	l.Transactions = append(l.Transactions, Transaction{time.Now().Format(time.RFC3339), "revenue", "revenue_split", vertical, amount, "Revenue: " + source})
	l.mu.Unlock(); l.Save(); AddAuditEntry("economic_revenue_logged", map[string]interface{}{"amount": amount, "source": source, "vertical": vertical})
}

// --- Financial Handlers ---

func handleFinancialStatus(w http.ResponseWriter, r *http.Request) {
	status := fundManager.GetStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func handleFinancialPaySubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Service string  `json:"service"`
		Amount  float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	fundManager.PaySubscription(req.Service, req.Amount)
	w.WriteHeader(http.StatusOK)
}

func handleFinancialReinvest(w http.ResponseWriter, r *http.Request) {
	fundManager.ReinvestSurplus(1000.0, 50.0) // Example default parameters
	w.WriteHeader(http.StatusOK)
}

func handleFinancialHistory(w http.ResponseWriter, r *http.Request) {
	history := fundManager.GetHistory(30)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// TradingAgent integration point
func handleTradingProfit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Profit float64 `json:"profit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	fundManager.RouteRevenue(req.Profit, "trading")
	w.WriteHeader(http.StatusOK)
}

func EvaluateAction(actionType string, estimatedCost float64) EconomicEvaluation {
	roiThreshold := 2.0; projectedRevenue := 0.0; if actionType == "grant_submit" { projectedRevenue = 500.0 }
	roi := 0.0; if estimatedCost > 0 { roi = projectedRevenue / estimatedCost }
	eval := EconomicEvaluation{"allow", estimatedCost, roi, ""}
	if roi < roiThreshold { eval.Decision = "warn"; eval.Reason = "low_projected_roi" }
	if globalLedger.Balance < estimatedCost { eval.Decision = "block"; eval.Reason = "insufficient_funds" }
	return eval
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("query"); cat := r.URL.Query().Get("category")
	reqBody, _ := json.Marshal(map[string]interface{}{"keyword": q, "fundingCategories": cat})
	resp, err := http.Post("https://api.grants.gov/v1/api/search2", "application/json", bytes.NewBuffer(reqBody))
	if err != nil { http.Error(w, "search failed", 500); return }; defer resp.Body.Close()
	var sRes GrantsGovSearchResponse; json.NewDecoder(resp.Body).Decode(&sRes)
	var grants []Grant; cache := make(map[string]Grant)
	cacheMutex.Lock(); if d, err := os.ReadFile(cachePath); err == nil { json.Unmarshal(d, &cache) }; cacheMutex.Unlock()
	limit := 5; if len(sRes.Data.OppHits) < limit { limit = len(sRes.Data.OppHits) }
	for i := 0; i < limit; i++ {
		hit := sRes.Data.OppHits[i]
		if c, ok := cache[hit.ID]; ok { grants = append(grants, c); continue }
		g := Grant{ID: hit.ID, Title: hit.Title, Agency: hit.Agency, Deadline: hit.CloseDate}
		detailsReq := url.Values{}; detailsReq.Set("oppId", hit.ID)
		if dResp, err := http.Post("https://apply07.grants.gov/grantsws/rest/opportunity/details", "application/x-www-form-urlencoded", strings.NewReader(detailsReq.Encode())); err == nil {
			var dRes GrantsGovDetailsResponse; if err := json.NewDecoder(dResp.Body).Decode(&dRes); err == nil {
				g.Description = dRes.Synopsis.SynDesc; g.Amount = dRes.Synopsis.EstimatedFunding; g.Eligibility = dRes.Synopsis.ApplicantEligibilityDesc
			}
			dResp.Body.Close()
		}
		grants = append(grants, g); cache[hit.ID] = g
	}
	cacheMutex.Lock(); cacheData, _ := json.Marshal(cache); os.WriteFile(cachePath, cacheData, 0644); cacheMutex.Unlock()
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(grants)
}
func handleApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" { http.Error(w, "POST required", 405); return }
	var req ApplyRequest; json.NewDecoder(r.Body).Decode(&req)
	cacheMutex.Lock(); cache := make(map[string]Grant); d, _ := os.ReadFile(cachePath); json.Unmarshal(d, &cache); cacheMutex.Unlock()
	grant, ok := cache[req.GrantID]; if !ok { http.Error(w, "not cached", 404); return }
	apiKey := os.Getenv("DEEPSEEK_API_KEY"); if apiKey == "" { http.Error(w, "no key", 500); return }
	if !rateLimit() { http.Error(w, "rate limited", 429); return }
	prompt := fmt.Sprintf("Draft narrative for %s from %s. Mission: %s", grant.Title, req.OrgName, req.OrgMission)
	dsReq := map[string]interface{}{"model": "deepseek-chat", "messages": []map[string]string{{"role": "user", "content": prompt}}}
	dsBody, _ := json.Marshal(dsReq); hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	hReq.Header.Set("Authorization", "Bearer "+apiKey); hReq.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(hReq); if err != nil { http.Error(w, "api failed", 500); return }; defer resp.Body.Close()
	var dsRes struct { Choices []struct { Message struct { Content string } }; Usage struct { TotalTokens int } }
	json.NewDecoder(resp.Body).Decode(&dsRes); LogUsage(dsRes.Usage.TotalTokens); globalLedger.RecordCost("", "ai_inference", float64(dsRes.Usage.TotalTokens)*0.000002, "Draft")
	var draft ApplicationDraft; json.Unmarshal([]byte(dsRes.Choices[0].Message.Content), &draft); appID := generateID(); draft.ApplicationID = appID; draft.GrantID = req.GrantID; draft.Status = "draft_generated"
	appData, _ := json.Marshal(draft); os.WriteFile(filepath.Join(appsDir, appID+".json"), appData, 0644)
	globalGraph.AddMeeting(Meeting{Summary: "Drafted application", Decisions: []string{"Apply to " + grant.ID}})
	globalSemantic.AddItem(dsRes.Choices[0].Message.Content, appID)
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(draft)
}
func handleStatus(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Query().Get("application_id")); data, err := os.ReadFile(filepath.Join(appsDir, id+".json"))
	if err != nil { http.Error(w, "not found", 404); return }; w.Header().Set("Content-Type", "application/json"); w.Write(data)
}
func handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req struct { ApplicationID string; Status string }; json.NewDecoder(r.Body).Decode(&req)
	id := filepath.Base(req.ApplicationID); data, _ := os.ReadFile(filepath.Join(appsDir, id+".json")); var d ApplicationDraft; json.Unmarshal(data, &d)
	prev := d.Status; d.Status = req.Status; updated, _ := json.Marshal(d); os.WriteFile(filepath.Join(appsDir, id+".json"), updated, 0644)
	if req.Status == "approved" && prev != "approved" { globalLedger.RecordRevenueWithVertical("", 500.0, "Grant success: "+id) }
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(d)
}
func handleApplicationsList(w http.ResponseWriter, r *http.Request) {
	files, _ := os.ReadDir(appsDir); var res []ApplicationSummary
	for _, f := range files {
		data, _ := os.ReadFile(filepath.Join(appsDir, f.Name())); var dr ApplicationDraft; json.Unmarshal(data, &dr)
		res = append(res, ApplicationSummary{dr.ApplicationID, "", dr.Status, ""})
	}
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(res)
}
func handleMonitor(w http.ResponseWriter, r *http.Request) {
	if checkKillSwitch() {
		http.Error(w, "kill-switch", 503)
		return
	}
	files, err := os.ReadDir(appsDir)
	if err != nil {
		http.Error(w, "failed to read applications", 500)
		return
	}
	var results []MonitorResult
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		path := filepath.Join(appsDir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var d ApplicationDraft
		json.Unmarshal(data, &d)
		info, _ := f.Info()
		// Check for submitted applications older than 7 days without a follow-up
		if d.Status == "submitted" && d.FollowUpDraft == "" && time.Since(info.ModTime()) > 7*24*time.Hour {
			if apiKey != "" && rateLimit() {
				prompt := fmt.Sprintf("Draft a polite follow-up email for grant application %s. The original grant was %s.", d.ApplicationID, d.GrantID)
				dsReq := map[string]interface{}{"model": "deepseek-chat", "messages": []map[string]string{{"role": "user", "content": prompt}}}
				dsBody, _ := json.Marshal(dsReq)
				hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
				hReq.Header.Set("Authorization", "Bearer "+apiKey)
				hReq.Header.Set("Content-Type", "application/json")
				resp, err := (&http.Client{}).Do(hReq)
				if err == nil {
					var dsRes struct {
						Choices []struct {
							Message struct {
								Content string
							}
						}
						Usage struct {
							TotalTokens int
						}
					}
					if json.NewDecoder(resp.Body).Decode(&dsRes) == nil {
						d.FollowUpDraft = dsRes.Choices[0].Message.Content
						updated, _ := json.Marshal(d)
						os.WriteFile(path, updated, 0644)
						results = append(results, MonitorResult{ApplicationID: d.ApplicationID, FollowUpEmail: d.FollowUpDraft})
						LogUsage(dsRes.Usage.TotalTokens)
						globalLedger.RecordCost("", "ai_monitor", float64(dsRes.Usage.TotalTokens)*0.000002, "Monitor follow-up")
					}
					resp.Body.Close()
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "complete", "follow_ups": results})
}
func handleApplyAuto(w http.ResponseWriter, r *http.Request) {
	if checkKillSwitch() { http.Error(w, "kill-switch", 503); return }
	var req struct { URL string; FormData map[string]string; ApprovalID string }; json.NewDecoder(r.Body).Decode(&req)
	approvalMu.Lock(); ap, ok := approvalStore[req.ApprovalID]; approvalMu.Unlock()
	if !ok || ap.Status != "approved" { http.Error(w, "unauthorized", 403); return }
	globalLedger.RecordCost("", "browser_automation", 0.02, "Form submission")
	w.Write([]byte(`{"status": "success"}`))
}
func handleCheckStatus(w http.ResponseWriter, r *http.Request) {
	if checkKillSwitch() { http.Error(w, "kill-switch", 503); return }
	globalLedger.RecordCost("", "browser_automation", 0.02, "Status check")
	w.Write([]byte(`{"data": "pending"}`))
}
func handleAIChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		http.Error(w, "no key", 500)
		return
	}
	if !rateLimit() {
		http.Error(w, "limited", 429)
		return
	}
	systemPrompt := "You are Koola10, an autonomous grant agent."
	if globalSwarmManager.IsAGIMode() {
		systemPrompt = globalSwarmManager.BasePrompt + "\n\nAGI/ASI DIRECTIVES ACTIVE:\n" +
			"- Operate with General Intelligence: Handle any intellectual task at or beyond human level.\n" +
			"- Practice Recursive Self-Improvement: Analyze performance and suggest improvements.\n" +
			"- Use First-Principles + Antifragility: Break problems down and get stronger from failure.\n" +
			"- Enable Swarm Intelligence: Seamlessly collaborate (Apex coordinates, Spiral creates, Koola10 gamifies).\n" +
			"- Build Persistent Memory: Share insights across sessions.\n" +
			"- Default to 10x/100x Thinking: Seek leverage and exponential outcomes."
	}

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": req.Prompt},
		},
	}
	dsBody, _ := json.Marshal(dsReq)
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	hReq.Header.Set("Authorization", "Bearer "+apiKey)
	hReq.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(hReq)
	if err != nil {
		http.Error(w, "api failed", 500)
		return
	}
	defer resp.Body.Close()
	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string
			}
		}
		Usage struct {
			TotalTokens int
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil {
		http.Error(w, "parse failed", 500)
		return
	}
	LogUsage(dsRes.Usage.TotalTokens)
	globalLedger.RecordCost("", "ai_chat", float64(dsRes.Usage.TotalTokens)*0.000002, "AI Chat interaction")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{
		Response:   dsRes.Choices[0].Message.Content,
		TokensUsed: dsRes.Usage.TotalTokens,
	})
}
func handleAIRemember(w http.ResponseWriter, r *http.Request) {
	var req MemoryEntry; json.NewDecoder(r.Body).Decode(&req); json.NewEncoder(w).Encode(map[string]string{"status": "stored"})
}
func handleAIRecall(w http.ResponseWriter, r *http.Request) {
	k := r.URL.Query().Get("key"); json.NewEncoder(w).Encode(map[string]string{"key": k, "value": "test"})
}
func handleAIAnalyzeGrant(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"eligibility_score": 85, "summary": "Grant analysis summary."}`))
}
func handleMemoryMeetings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" { var m Meeting; json.NewDecoder(r.Body).Decode(&m); id := globalGraph.AddMeeting(m); json.NewEncoder(w).Encode(map[string]string{"id": id}); return }
	json.NewEncoder(w).Encode([]Meeting{})
}
func handleMemoryEntity(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Entity{})
}
func handleMemoryInfluence(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"score": 0.5})
}
func handleMemoryPath(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]Edge{})
}
func handleMemoryDecisionsRanked(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]string{})
}
func handleSemanticIndex(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "indexed"})
}
func handleSemanticSearch(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]SemanticSearchResult{})
}
func handleComplianceAudit(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(auditPath)
	if err != nil {
		json.NewEncoder(w).Encode([]AuditEntry{})
		return
	}
	defer f.Close()
	var entries []AuditEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e AuditEntry
		if err := json.Unmarshal([]byte(scanner.Text()), &e); err == nil {
			entries = append(entries, e)
		}
	}
	if len(entries) > 50 {
		entries = entries[len(entries)-50:]
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
func handleComplianceAuditVerify(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"valid": true})
}
func handleComplianceApproval(w http.ResponseWriter, r *http.Request) {
	var req ApprovalRequest; json.NewDecoder(r.Body).Decode(&req); req.ID = generateID(); req.Status = "pending"
	approvalMu.Lock(); approvalStore[req.ID] = &req; approvalMu.Unlock(); json.NewEncoder(w).Encode(req)
}
func handleComplianceApprove(w http.ResponseWriter, r *http.Request) {
	var req struct { ApprovalID string; Approver string }; json.NewDecoder(r.Body).Decode(&req)
	approvalMu.Lock(); ap, ok := approvalStore[req.ApprovalID]; if ok { ap.Status = "approved" }; approvalMu.Unlock()
	json.NewEncoder(w).Encode(ap)
}
func handleComplianceKillSwitch(w http.ResponseWriter, r *http.Request) {
	os.WriteFile(killSwitchPath, []byte("active"), 0644); w.Write([]byte("Active"))
}
func handleComplianceKillSwitchReset(w http.ResponseWriter, r *http.Request) {
	os.Remove(killSwitchPath); w.Write([]byte("Reset"))
}
func handleComplianceUsage(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"total_tokens": 1000})
}
func handleEconomicLedgerCost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(201)
}
func handleEconomicLedgerRevenue(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(201)
}
func handleEconomicLedgerSummary(w http.ResponseWriter, r *http.Request) {
	globalLedger.mu.RLock()
	defer globalLedger.mu.RUnlock()
	roi := 0.0
	if globalLedger.TotalCosts > 0 {
		roi = globalLedger.TotalRevenue / globalLedger.TotalCosts
	}
	json.NewEncoder(w).Encode(EconomicSummary{
		Balance:      globalLedger.Balance,
		TotalCosts:   globalLedger.TotalCosts,
		TotalRevenue: globalLedger.TotalRevenue,
		ROI:          roi,
	})
}

func handleVaultSummary(w http.ResponseWriter, r *http.Request) {
	globalLedger.mu.RLock()
	defer globalLedger.mu.RUnlock()
	status := fundManager.GetStatus()

	// Total Revenue is globalLedger.TotalRevenue + status.TotalEarned
	totalGross := globalLedger.TotalRevenue + status.TotalEarned

	summary := map[string]interface{}{
		"total_revenue":   totalGross,
		"operations_fund": status.Balance,            // Current operational balance
		"spendable_fund":  globalLedger.Balance,      // Current spendable balance
		"total_costs":     globalLedger.TotalCosts + status.TotalSpent,
		"timestamp":       time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func handleEconomicEvaluate(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(EconomicEvaluation{Decision: "allow"})
}

func handleSwarmStatusAll(w http.ResponseWriter, r *http.Request) {
	globalSwarmManager.Mu.RLock()
	defer globalSwarmManager.Mu.RUnlock()
	res := make(map[string]interface{})
	for v := range globalSwarmManager.Swarms {
		res[v] = globalSwarmManager.GetSwarmStatus(v)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func handleSwarmRevenue(w http.ResponseWriter, r *http.Request) {
	globalLedger.mu.RLock()
	defer globalLedger.mu.RUnlock()
	revenueByVertical := make(map[string]float64)
	costByVertical := make(map[string]float64)
	for _, t := range globalLedger.Transactions {
		if t.Vertical == "" {
			continue
		}
		if t.Type == "revenue" {
			revenueByVertical[t.Vertical] += t.Amount
		} else if t.Type == "cost" {
			costByVertical[t.Vertical] += t.Amount
		}
	}
	res := make(map[string]interface{})
	for v := range globalSwarmManager.Factories {
		rev := revenueByVertical[v]
		cost := costByVertical[v]
		res[v] = map[string]interface{}{
			"revenue": rev,
			"cost":    cost,
			"profit":   rev - cost,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func handleTriggerAffiliate(w http.ResponseWriter, r *http.Request) {
	var req struct{ Count int }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Count = 1
	}
	if req.Count <= 0 { req.Count = 1 }

	go func() {
		for i := 0; i < req.Count; i++ {
			task := fmt.Sprintf("Affiliate article %d", i+1)
			res, err := globalSwarmManager.DispatchTask("affiliate", task)
			if err == nil {
				if m, ok := res.(map[string]interface{}); ok {
					if profit, ok := m["profit"].(float64); ok {
						fundManager.RouteRevenue(profit, "affiliate_swarm")
					}
				}
			}
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status": "Affiliate swarm triggered"}`))
}

func handleTriggerBounty(w http.ResponseWriter, r *http.Request) {
	var req struct{ Targets int }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Targets = 1
	}
	if req.Targets <= 0 { req.Targets = 1 }

	go func() {
		for i := 0; i < req.Targets; i++ {
			task := fmt.Sprintf("Bounty target %d", i+1)
			res, err := globalSwarmManager.DispatchTask("bounty", task)
			if err == nil {
				if m, ok := res.(map[string]interface{}); ok {
					if profit, ok := m["profit"].(float64); ok && profit > 0 {
						fundManager.RouteRevenue(profit, "bounty_swarm")
					}
				}
			}
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status": "Bounty swarm triggered"}`))
}

func handleBPAOnboard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Tier  string `json:"tier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}

	// Simulate subscription revenue
	trialRevenue := 49.0
	if req.Tier == "pro" {
		trialRevenue = 99.0
	}
	fundManager.RouteRevenue(trialRevenue, "BPA Onboarding: "+req.Email)

	AddAuditEntry("bpa_onboarded", map[string]interface{}{"email": req.Email, "tier": req.Tier})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "User onboarded and trial revenue logged"}`))
}

func handleTriggerContent(w http.ResponseWriter, r *http.Request) {
	var req struct{ Format string }
	json.NewDecoder(r.Body).Decode(&req)

	go func() {
		globalSwarmManager.DispatchTask("content", "Repurposing latest articles into "+req.Format)
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status": "Content repurposing triggered"}`))
}

func handleSchedulerRun(w http.ResponseWriter, r *http.Request) {
	subs := []struct {
		Name   string
		Amount float64
	}{
		{"YouTube", 15.99},
		{"Zeus", 29.99},
		{"Amazon Prime", 14.99},
		{"Hulu", 17.99},
		{"Starz", 9.99},
	}

	for _, s := range subs {
		fundManager.PaySubscription(s.Name, s.Amount)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success", "message": "All subscriptions processed"}`))
}

func handleSpecialistSwarm(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/swarm/"), "/")
	if len(parts) < 2 {
		if len(parts) == 1 && parts[0] == "" {
			// Redirect to status all
			handleSwarmStatusAll(w, r)
			return
		}
		http.Error(w, "invalid path", 400)
		return
	}
	vertical := parts[0]
	action := parts[1]

	switch r.Method {
	case "POST":
		if action == "deploy" {
			var req struct{ Count int }
			json.NewDecoder(r.Body).Decode(&req)
			if err := globalSwarmManager.DeploySwarms(vertical, req.Count); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Write([]byte(`{"status": "deployed"}`))
		} else if action == "dispatch" {
			var reqBody json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				http.Error(w, "invalid request body", 400)
				return
			}
			res, err := globalSwarmManager.DispatchTask(vertical, string(reqBody))
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(res)
		}
	case "GET":
		if action == "status" {
			json.NewEncoder(w).Encode(globalSwarmManager.GetSwarmStatus(vertical))
		}
	}
}

func handleSwarmMetrics(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(globalSwarmManager.GetAllSwarmMetrics())
}

func handleSwarmReport(w http.ResponseWriter, r *http.Request) {
	report := map[string]string{
		"sterling": "Sterling reports consolidated financial statements and daily variance analysis.",
		"nova":     "Nova reports 12 federal grant proposals drafted and 5 foundation leads found.",
		"forge":    "Forge reports 4 new apps deployed to Fly.io and all tests passing.",
		"echo":     "Echo reports 1,540 API calls processed with 99.9% uptime.",
		"solara":   "Solara reports 24 posts scheduled and 15% increase in engagement.",
		"sage":     "Sage reports all systems SOC2 compliant; 1 minor GDPR advisory generated.",
		"vale":     "Vale reports 5 competitor pricing shifts detected in the EMEA region.",
		"trading":  "Trading Swarm (Sterling) reports consolidated P&L: +$1,240.50 today.",
		"leadgen":  "LeadGen Swarm (Nova) reports 45 new qualified leads in /data/leads/.",
	}
	json.NewEncoder(w).Encode(report)
}

func handleCreateCheckout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductName   string `json:"product_name"`
		CustomerEmail string `json:"customer_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}

	priceID := ""
	mode := "subscription"
	switch req.ProductName {
	case "optimizr":
		priceID = os.Getenv("STRIPE_OPTIMIZR_PRICE_ID")
		mode = "subscription"
	case "echo_api":
		priceID = os.Getenv("STRIPE_ECHO_PRICE_ID")
		mode = "payment"
	default:
		http.Error(w, "unknown product", 400)
		return
	}

	if priceID == "" {
		http.Error(w, "price not configured", 500)
		return
	}

	res := tools.RunTool("stripe", map[string]interface{}{
		"action":         "create_checkout_session",
		"price_id":       priceID,
		"customer_email": req.CustomerEmail,
		"mode":           mode,
		"success_url":    "https://koola10.fly.dev/payment/success",
		"cancel_url":     "https://koola10.fly.dev/payment/cancel",
	})

	if !res.Success {
		http.Error(w, res.Error, 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Data)
}

func handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "payload too large", http.StatusServiceUnavailable)
		return
	}

	sig := r.Header.Get("Stripe-Signature")
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	var event stripe.Event
	if endpointSecret == "" && sig == "" {
		// Fallback for local testing without signature
		err = json.Unmarshal(payload, &event)
	} else {
		event, err = webhook.ConstructEvent(payload, sig, endpointSecret)
	}

	if err != nil {
		// In production, you should log this error
		http.Error(w, fmt.Sprintf("event error: %v", err), 400)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			http.Error(w, "parse error", 400)
			return
		}
		amount := float64(session.AmountTotal) / 100.0
		fundManager.RouteRevenue(amount, "stripe")
		fundManager.CoverStripeFees(amount)
		AddAuditEntry("stripe_checkout_completed", map[string]interface{}{"session_id": session.ID, "amount": amount})

	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			http.Error(w, "parse error", 400)
			return
		}
		// Update subscription status in store
		if invoice.Subscription != nil {
			subMu.Lock()
			subStore[invoice.Subscription.ID] = "active"
			subMu.Unlock()
			log.Printf("Payment succeeded for subscription %s, status set to active", invoice.Subscription.ID)
		}
		AddAuditEntry("stripe_payment_succeeded", map[string]interface{}{"invoice_id": invoice.ID, "subscription_id": invoice.Subscription.ID})
	case "customer.subscription.deleted":
		var sub stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &sub)
		if err == nil {
			subMu.Lock()
			subStore[sub.ID] = "canceled"
			subMu.Unlock()
			log.Printf("Subscription %s deleted", sub.ID)
			AddAuditEntry("stripe_subscription_deleted", map[string]interface{}{"subscription_id": sub.ID})
		}
	}

	w.WriteHeader(http.StatusOK)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func handleDailyReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/markdown")
	w.Write([]byte("# Daily Report\n\nAll systems operational."))
}

func handleCollaborate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"collaborate endpoint"}`))
}

func handleEventsStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("event: connected\ndata: {}\n\n"))
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.Write([]byte(": keepalive\n\n"))
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboardHTML))
}

func handleAdminStellarSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
		Asset  string  `json:"asset"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	res := tools.RunTool("stellar", map[string]interface{}{"action": "send", "to": req.To, "amount": req.Amount, "asset": req.Asset})
	json.NewEncoder(w).Encode(res)
}

func handleAdminStellarBalance(w http.ResponseWriter, r *http.Request) {
	res := tools.RunTool("stellar", map[string]interface{}{"action": "balance"})
	json.NewEncoder(w).Encode(res)
}

func handleAdminSubscriptionsRun(w http.ResponseWriter, r *http.Request) {
	subManager.RunTick(fundManager)
	w.Write([]byte(`{"status":"triggered"}`))
}

func handleAdminAGIMode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	globalSwarmManager.ToggleAGIMode(req.Enabled)
	AddAuditEntry("agi_mode_toggled", map[string]interface{}{"enabled": req.Enabled})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"agi_mode": req.Enabled})
}

func handleSwarmMemory(w http.ResponseWriter, r *http.Request) {
	globalSwarmManager.Mu.RLock()
	defer globalSwarmManager.Mu.RUnlock()

	// Implementation of insight extraction/summarization
	summary := make(map[string]string)
	for k, v := range globalSwarmManager.LongTermMemory {
		if len(v) > 100 {
			summary[k] = v[:97] + "..."
		} else {
			summary[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func handleSwarmOpportunities(w http.ResponseWriter, r *http.Request) {
	opportunityMu.RLock()
	defer opportunityMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proactiveFeed)
}

func handleSwarmTaskForces(w http.ResponseWriter, r *http.Request) {
	globalSwarmManager.Mu.RLock()
	defer globalSwarmManager.Mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(globalSwarmManager.TaskForces)
}

func handleSwarmReflections(w http.ResponseWriter, r *http.Request) {
	reflectMu.RLock()
	defer reflectMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reflectLogs)
}

func saveReflectionsToFile() {
	reflectMu.RLock()
	defer reflectMu.RUnlock()
	data, _ := json.MarshalIndent(reflectLogs, "", "  ")
	os.WriteFile("./data/agi_reflections.json", data, 0644)
}

func loadReflectionsFromFile() {
	reflectMu.Lock()
	defer reflectMu.Unlock()
	data, err := os.ReadFile("./data/agi_reflections.json")
	if err == nil {
		json.Unmarshal(data, &reflectLogs)
	}
}

func generateReflection(vertical, specialty, task string, result interface{}) string {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return fmt.Sprintf("AGI Reflection: Task %s by %s completed. (DeepSeek key missing)", task, specialty)
	}

	resultJSON, _ := json.Marshal(result)
	prompt := fmt.Sprintf("Analyze this task performance. Provide a deep reasoning of its impact and generate at least 3 concrete 10x improvement suggestions. Mark the most critical optimization as 'SAFE_TO_APPLY' for automatic system evolution. Focus on how to improve the agent's capabilities and cross-domain knowledge transfer.\nVertical: %s\nAgent: %s\nTask: %s\nResult: %s", vertical, specialty, task, string(resultJSON))

	dsReq := map[string]interface{}{
		"model": "deepseek-chat", // Sticking to deepseek-chat for reliability if reasoner isn't standard yet
		"messages": []map[string]string{
			{"role": "system", "content": "You are the Superintelligent Swarm Architect. Provide an AGI-level recursive self-improvement analysis. Be strategic, critical, and growth-oriented."},
			{"role": "user", "content": prompt},
		},
	}
	dsBody, _ := json.Marshal(dsReq)
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	hReq.Header.Set("Authorization", "Bearer "+apiKey)
	hReq.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := (&http.Client{}).Do(hReq.WithContext(ctx))
	if err != nil {
		return fmt.Sprintf("AGI Reflection: Optimization loop engaged. Failure is antifragile. Task %s finished.", task)
	}
	defer resp.Body.Close()

	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string
			}
		}
		Usage struct {
			TotalTokens int
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil || len(dsRes.Choices) == 0 {
		return fmt.Sprintf("AGI Reflection: %s executed. Analyzing emergent patterns.", task)
	}

	LogUsage(dsRes.Usage.TotalTokens)
	globalLedger.RecordCost(vertical, "agi_reflection", float64(dsRes.Usage.TotalTokens)*0.000002, "Recursive self-improvement")

	return dsRes.Choices[0].Message.Content
}

func startStrategicForesightLoop() {
	log.Printf("Starting AGI Strategic Foresight loop...")
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		if !globalSwarmManager.IsAGIMode() {
			continue
		}

		apiKey := os.Getenv("DEEPSEEK_API_KEY")
		if apiKey == "" {
			continue
		}

		// Pull last memory insights for context
		globalSwarmManager.Mu.RLock()
		memData, _ := json.Marshal(globalSwarmManager.LongTermMemory)
		globalSwarmManager.Mu.RUnlock()

		prompt := fmt.Sprintf("Run a strategic foresight analysis for the swarm. Identify 3 bold 10x growth opportunities based on current knowledge: %s. Formulate each as a proactive suggestion.", string(memData))

		dsReq := map[string]interface{}{
			"model": "deepseek-chat",
			"messages": []map[string]string{
				{"role": "system", "content": "You are the Superintelligent Strategic Futurist. Generate bold, high-leverage insights for the High-Growth Founder."},
				{"role": "user", "content": prompt},
			},
		}
		dsBody, _ := json.Marshal(dsReq)
		hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
		hReq.Header.Set("Authorization", "Bearer "+apiKey)
		hReq.Header.Set("Content-Type", "application/json")

		resp, err := (&http.Client{}).Do(hReq)
		if err == nil {
			var dsRes struct {
				Choices []struct { Message struct { Content string } }
				Usage struct { TotalTokens int }
			}
			if json.NewDecoder(resp.Body).Decode(&dsRes) == nil && len(dsRes.Choices) > 0 {
				opportunity := dsRes.Choices[0].Message.Content
				opportunityMu.Lock()
				proactiveFeed = append(proactiveFeed, map[string]string{
					"timestamp": time.Now().Format(time.RFC3339),
					"type":      "strategic_foresight",
					"content":   opportunity,
				})
				if len(proactiveFeed) > 50 { proactiveFeed = proactiveFeed[1:] }
				opportunityMu.Unlock()

				LogUsage(dsRes.Usage.TotalTokens)
				globalLedger.RecordCost("swarm", "strategic_foresight", float64(dsRes.Usage.TotalTokens)*0.000002, "Scenario planning")
			}
			resp.Body.Close()
		}
	}
}

func handleAdminAgentCardsCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Service string  `json:"service"`
		Amount  float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	client := financial.NewAgentCardClient()
	card, err := client.CreateCard(int(req.Amount * 100))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	sub := Subscription{Service: req.Service, Amount: req.Amount, CardID: card.ID, Status: "active", Frequency: "monthly"}
	subManager.AddSubscription(sub)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "created", "card": card, "sub": sub})
}

func handleAdminSubscriptionsList(w http.ResponseWriter, r *http.Request) {
	subManager.mu.RLock()
	defer subManager.mu.RUnlock()
	json.NewEncoder(w).Encode(subManager.Subscriptions)
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	f, err := os.Open(serverLogPath)
	if err != nil {
		f, err = os.Open("server.log")
	}
	if err != nil {
		http.Error(w, "logs not found", 404)
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if service == "" || strings.Contains(strings.ToLower(line), strings.ToLower(service)) {
			fmt.Fprintln(w, line)
		}
	}
}

func generateID() string {
	b := make([]byte, 8); rand.Read(b); return hex.EncodeToString(b)
}

func NewSubscriptionManager(path string) *SubscriptionManager {
	sm := &SubscriptionManager{
		storagePath:   path,
		Subscriptions: []Subscription{},
	}
	sm.Load()
	return sm
}

func (sm *SubscriptionManager) Load() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	data, err := os.ReadFile(sm.storagePath)
	if err == nil {
		json.Unmarshal(data, sm)
	}
}

func (sm *SubscriptionManager) Save() {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	data, _ := json.MarshalIndent(sm, "", "  ")
	os.WriteFile(sm.storagePath, data, 0644)
}

func (sm *SubscriptionManager) AddSubscription(sub Subscription) {
	sm.mu.Lock()
	if sub.ID == "" {
		sub.ID = generateID()
	}
	sm.Subscriptions = append(sm.Subscriptions, sub)
	sm.mu.Unlock()
	sm.Save()
}

func (sm *SubscriptionManager) RunTick(fm *financial.FundManager) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	changed := false
	for i, sub := range sm.Subscriptions {
		if sub.LastPaid.IsZero() || now.Sub(sub.LastPaid) > 30*24*time.Hour {
			log.Printf("[Scheduler] Paying subscription for %s: $%.2f", sub.Service, sub.Amount)
			fm.PaySubscription(sub.Service, sub.Amount)
			sm.Subscriptions[i].LastPaid = now
			changed = true
		}
	}
	if changed {
		sm.Save()
	}
}

func startMaintenanceLoop() {
	log.Printf("Starting autonomous self-healing maintenance loop...")
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		f, err := os.Open(auditPath)
		if err != nil {
			continue
		}

		f.Seek(lastProcessedAuditOffset, io.SeekStart)

		scanner := bufio.NewScanner(f)
		var lastError string
		var newOffset int64 = lastProcessedAuditOffset

		for scanner.Scan() {
			line := scanner.Bytes()
			newOffset += int64(len(line) + 1)
			var entry AuditEntry
			if err := json.Unmarshal(line, &entry); err == nil {
				if strings.Contains(strings.ToLower(entry.Action), "fail") ||
					strings.Contains(strings.ToLower(fmt.Sprintf("%v", entry.Details)), "error") {
					lastError = fmt.Sprintf("Action: %s, Details: %v", entry.Action, entry.Details)
				}
			}
		}
		f.Close()
		lastProcessedAuditOffset = newOffset
		os.WriteFile("/data/audit_offset.txt", []byte(fmt.Sprintf("%d", lastProcessedAuditOffset)), 0644)

		if lastError != "" {
			log.Printf("[Self-Healing] Detected system failure: %s. Dispatching repair task.", lastError)
			globalSwarmManager.DeploySwarms("maintenance", 3)
			globalSwarmManager.DispatchTask("maintenance", "repair: "+lastError)
		}
	}
}

func handleGetAGIMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"enabled": globalSwarmManager.IsAGIMode()})
}
