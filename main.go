package main

import (
	"bufio"
	"bytes"
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

// --- Global States ---

var (
	cacheMutex   sync.Mutex
	auditMutex   sync.Mutex
	usageMutex   sync.Mutex
	approvalMu   sync.Mutex
	killSwitchMu sync.Mutex

	cachePath      = "/data/grants_cache.json"
	appsDir        = "/data/applications"
	memoryPath     = "/data/memory.json"
	graphPath      = "/data/memory_graph.json"
	semanticPath   = "/data/semantic_index.json"
	auditPath      = "/data/audit_chain.jsonl"
	usagePath      = "/data/usage.jsonl"
	killSwitchPath = "/data/kill_switch"
	ledgerPath     = "/data/economic_ledger.json"

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

	approvalStore = make(map[string]*ApprovalRequest)

	rlBucket     = 15.0
	rlMaxBucket  = 15.0
	rlRate       = 10.0
	rlLastUpdate = time.Now()
	rlMu         sync.Mutex
)

// --- Main ---

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	os.MkdirAll(filepath.Dir(cachePath), 0755)
	os.MkdirAll(appsDir, 0755)

	globalGraph.Load()
	globalSemantic.Load()
	globalLedger.Load()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Grant Endpoints
	http.HandleFunc("/grants/search", handleSearch)
	http.HandleFunc("/grants/apply", handleApply)
	http.HandleFunc("/grants/status", handleStatus)
	http.HandleFunc("/grants/applications", handleApplicationsList)
	http.HandleFunc("/grants/monitor", handleMonitor)
	http.HandleFunc("/grants/update-status", handleUpdateStatus)
	http.HandleFunc("/grants/apply-auto", handleApplyAuto)
	http.HandleFunc("/grants/check-status", handleCheckStatus)

	// AI & Memory Endpoints
	http.HandleFunc("/ai/chat", handleAIChat)
	http.HandleFunc("/ai/remember", handleAIRemember)
	http.HandleFunc("/ai/recall", handleAIRecall)
	http.HandleFunc("/ai/analyze-grant", handleAIAnalyzeGrant)

	http.HandleFunc("/memory/meetings", handleMemoryMeetings)
	http.HandleFunc("/memory/entity/", handleMemoryEntity)
	http.HandleFunc("/memory/influence/", handleMemoryInfluence)
	http.HandleFunc("/memory/path", handleMemoryPath)
	http.HandleFunc("/memory/decisions/ranked", handleMemoryDecisionsRanked)

	http.HandleFunc("/semantic/index", handleSemanticIndex)
	http.HandleFunc("/semantic/search", handleSemanticSearch)

	// Compliance & Economic Endpoints
	http.HandleFunc("/compliance/audit", handleComplianceAudit)
	http.HandleFunc("/compliance/audit/verify", handleComplianceAuditVerify)
	http.HandleFunc("/compliance/approval", handleComplianceApproval)
	http.HandleFunc("/compliance/approve", handleComplianceApprove)
	http.HandleFunc("/compliance/kill-switch", handleComplianceKillSwitch)
	http.HandleFunc("/compliance/kill-switch/reset", handleComplianceKillSwitchReset)
	http.HandleFunc("/compliance/usage", handleComplianceUsage)

	http.HandleFunc("/economic/ledger/cost", handleEconomicLedgerCost)
	http.HandleFunc("/economic/ledger/revenue", handleEconomicLedgerRevenue)
	http.HandleFunc("/economic/ledger/summary", handleEconomicLedgerSummary)
	http.HandleFunc("/economic/evaluate", handleEconomicEvaluate)

	log.Printf("starting server on 0.0.0.0:%s", port)
	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		log.Fatalf("server_failed: %v", err)
	}
}

// --- Methods ---

func (g *MemoryGraph) Save() {
	g.mu.RLock()
	defer g.mu.RUnlock()
	data, _ := json.Marshal(g)
	os.WriteFile(graphPath, data, 0644)
}

func (g *MemoryGraph) Load() {
	g.mu.Lock()
	defer g.mu.Unlock()
	data, err := os.ReadFile(graphPath)
	if err == nil {
		json.Unmarshal(data, g)
	}
	if g.Meetings == nil { g.Meetings = make(map[string]Meeting) }
	if g.Entities == nil { g.Entities = make(map[string]Entity) }
}

func (g *MemoryGraph) AddWeightedEdge(source, target, relation string, metadata map[string]interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i, edge := range g.Edges {
		if edge.Source == source && edge.Target == target && edge.Relation == relation {
			g.Edges[i].Weight += 0.2
			if g.Edges[i].Weight > 2.0 { g.Edges[i].Weight = 2.0 }
			g.Edges[i].Frequency++
			return
		}
	}
	g.Edges = append(g.Edges, Edge{Source: source, Target: target, Relation: relation, Weight: 1.0, Frequency: 1, Metadata: metadata})
}

func (g *MemoryGraph) AddMeeting(m Meeting) string {
	if m.MeetingID == "" { m.MeetingID = generateID() }
	if m.Timestamp == "" { m.Timestamp = time.Now().Format(time.RFC3339) }
	g.mu.Lock()
	g.Meetings[m.MeetingID] = m
	g.mu.Unlock()
	for _, decision := range m.Decisions {
		g.mu.Lock()
		if _, ok := g.Entities[decision]; !ok { g.Entities[decision] = Entity{Name: decision, Type: "decision"} }
		g.mu.Unlock()
		g.AddWeightedEdge(m.MeetingID, decision, "contains_decision", nil)
	}
	for _, item := range m.ActionItems {
		parts := strings.Split(item, ":")
		taskName := item
		if len(parts) > 1 {
			owner := strings.TrimSpace(parts[0])
			taskName = strings.TrimSpace(parts[1])
			g.mu.Lock()
			entity, ok := g.Entities[owner]
			if !ok { entity = Entity{Name: owner, Type: "person"} }
			entity.Tasks = append(entity.Tasks, taskName)
			g.Entities[owner] = entity
			g.mu.Unlock()
			g.AddWeightedEdge(owner, taskName, "assigned_to", nil)
		}
		g.mu.Lock()
		if _, ok := g.Entities[taskName]; !ok { g.Entities[taskName] = Entity{Name: taskName, Type: "task"} }
		g.mu.Unlock()
		g.AddWeightedEdge(m.MeetingID, taskName, "contains_task", nil)
	}
	g.Save()
	return m.MeetingID
}

func (g *MemoryGraph) CalculateInfluenceScore(name string) float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var in, out float64
	var count int
	for _, e := range g.Edges {
		if e.Target == name { in += e.Weight; count++ }
		if e.Source == name { out += e.Weight; count++ }
	}
	if count == 0 { return 0 }
	return ((in * 0.7) + (out * 0.3)) / float64(count)
}

func (g *MemoryGraph) FindPath(source, target string, maxDepth int) []Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	type node struct { entity string; path []Edge }
	queue := []node{{source, []Edge{}}}
	visited := make(map[string]bool)
	for len(queue) > 0 {
		curr := queue[0]; queue = queue[1:]
		if curr.entity == target { return curr.path }
		if len(curr.path) >= maxDepth { continue }
		visited[curr.entity] = true
		for _, e := range g.Edges {
			if e.Source == curr.entity && !visited[e.Target] {
				newPath := append([]Edge{}, curr.path...)
				newPath = append(newPath, e)
				queue = append(queue, node{e.Target, newPath})
			}
		}
	}
	return nil
}

func (g *MemoryGraph) RankDecisionsByImpact() []string {
	g.mu.RLock()
	var res []string
	for n, e := range g.Entities { if e.Type == "decision" { res = append(res, n) } }
	g.mu.RUnlock()
	sort.Slice(res, func(i, j int) bool { return g.CalculateInfluenceScore(res[i]) > g.CalculateInfluenceScore(res[j]) })
	return res
}

func (s *SemanticIndex) Save() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, _ := json.Marshal(s)
	os.WriteFile(semanticPath, data, 0644)
}

func (s *SemanticIndex) Load() {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(semanticPath)
	if err == nil { json.Unmarshal(data, s) }
	if s.Items == nil { s.Items = []SemanticItem{} }
}

func (s *SemanticIndex) AddItem(text, refID string) error {
	url := os.Getenv("SEMANTIC_AGENT_URL")
	if url == "" { url = "https://koola10-semantic.fly.dev" }
	b, _ := json.Marshal(map[string]string{"text": text})
	resp, err := http.Post(url+"/generate", "application/json", bytes.NewBuffer(b))
	if err != nil { return err }
	if resp != nil { defer resp.Body.Close() }
	var res struct { Vector []float64 `json:"vector"` }
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil { return err }
	s.mu.Lock()
	s.Items = append(s.Items, SemanticItem{text, refID, res.Vector})
	s.mu.Unlock()
	s.Save()
	return nil
}

func (s *SemanticIndex) Search(query string, topK int) ([]SemanticSearchResult, error) {
	url := os.Getenv("SEMANTIC_AGENT_URL")
	if url == "" { url = "https://koola10-semantic.fly.dev" }
	s.mu.RLock()
	b, _ := json.Marshal(map[string]interface{}{"query": query, "embeddings": s.Items, "top_k": topK})
	s.mu.RUnlock()
	resp, err := http.Post(url+"/search", "application/json", bytes.NewBuffer(b))
	if err != nil { return nil, err }
	if resp != nil { defer resp.Body.Close() }
	var res []SemanticSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil { return nil, err }
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i, r := range res {
		for _, item := range s.Items { if item.RefID == r.RefID { res[i].Text = item.Text; break } }
	}
	return res, nil
}

func AddAuditEntry(action string, details map[string]interface{}) {
	auditMutex.Lock()
	defer auditMutex.Unlock()
	lastHash := "0000000000000000000000000000000000000000000000000000000000000000"
	if f, err := os.Open(auditPath); err == nil {
		scanner := bufio.NewScanner(f)
		var lastLine string
		for scanner.Scan() { lastLine = scanner.Text() }
		f.Close()
		if lastLine != "" {
			var e AuditEntry
			if err := json.Unmarshal([]byte(lastLine), &e); err == nil { lastHash = e.Hash }
		}
	}
	entry := AuditEntry{time.Now().Format(time.RFC3339), action, details, ""}
	entryJSON, _ := json.Marshal(entry)
	h := sha256.New(); h.Write([]byte(lastHash + string(entryJSON)))
	entry.Hash = hex.EncodeToString(h.Sum(nil))
	if f, err := os.OpenFile(auditPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		json.NewEncoder(f).Encode(entry); f.Close()
	}
}

func checkKillSwitch() bool {
	killSwitchMu.Lock(); defer killSwitchMu.Unlock()
	data, err := os.ReadFile(killSwitchPath)
	return err == nil && string(data) == "active"
}

func rateLimit() bool {
	rlMu.Lock(); defer rlMu.Unlock()
	now := time.Now(); elapsed := now.Sub(rlLastUpdate).Seconds()
	rlLastUpdate = now; rlBucket += elapsed * rlRate
	if rlBucket > rlMaxBucket { rlBucket = rlMaxBucket }
	if rlBucket >= 1.0 { rlBucket -= 1.0; return true }
	return false
}

func LogUsage(tokens int) {
	usageMutex.Lock(); defer usageMutex.Unlock()
	cost := float64(tokens) * 0.000002
	logEntry := UsageLog{time.Now().Format(time.RFC3339), tokens, cost}
	if f, err := os.OpenFile(usagePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		json.NewEncoder(f).Encode(logEntry); f.Close()
	}
}

func (l *EconomicLedger) Save() {
	l.mu.RLock(); defer l.mu.RUnlock()
	data, _ := json.Marshal(l); os.WriteFile(ledgerPath, data, 0644)
}

func (l *EconomicLedger) Load() {
	l.mu.Lock(); defer l.mu.Unlock()
	data, err := os.ReadFile(ledgerPath)
	if err == nil { json.Unmarshal(data, l) }
	if l.Transactions == nil { l.Transactions = []Transaction{} }
}

func (l *EconomicLedger) RecordCost(category string, amount float64, description string) {
	l.mu.Lock(); l.Balance -= amount; l.TotalCosts += amount
	l.Transactions = append(l.Transactions, Transaction{time.Now().Format(time.RFC3339), "cost", category, amount, description})
	l.mu.Unlock(); l.Save()
	AddAuditEntry("economic_cost_logged", map[string]interface{}{"amount": amount, "category": category})
}

func (l *EconomicLedger) RecordRevenue(amount float64, source string) {
	l.mu.Lock(); l.Balance += amount; l.TotalRevenue += amount
	l.Transactions = append(l.Transactions, Transaction{time.Now().Format(time.RFC3339), "revenue", "grant_success", amount, "Revenue from source: " + source})
	l.mu.Unlock(); l.Save()
	AddAuditEntry("economic_revenue_logged", map[string]interface{}{"amount": amount, "source": source})
}

func EvaluateAction(actionType string, estimatedCost float64) EconomicEvaluation {
	roiThreshold := 2.0; projectedRevenue := 0.0
	if actionType == "grant_submit" { projectedRevenue = 500.0 }
	roi := 0.0; if estimatedCost > 0 { roi = projectedRevenue / estimatedCost }
	eval := EconomicEvaluation{"allow", estimatedCost, roi, ""}
	if roi < roiThreshold { eval.Decision = "warn"; eval.Reason = "low_projected_roi" }
	if globalLedger.Balance < estimatedCost { eval.Decision = "block"; eval.Reason = "insufficient_funds" }
	return eval
}

// --- Handlers ---

func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("query"); cat := r.URL.Query().Get("category")
	reqBody, _ := json.Marshal(map[string]interface{}{"keyword": q, "fundingCategories": cat})
	resp, err := http.Post("https://api.grants.gov/v1/api/search2", "application/json", bytes.NewBuffer(reqBody))
	if err != nil { http.Error(w, "search failed", 500); return }
	if resp != nil { defer resp.Body.Close() }
	var sRes GrantsGovSearchResponse
	json.NewDecoder(resp.Body).Decode(&sRes)
	var grants []Grant; cache := make(map[string]Grant)
	cacheMutex.Lock()
	if d, err := os.ReadFile(cachePath); err == nil { json.Unmarshal(d, &cache) }
	cacheMutex.Unlock()
	limit := 5; if len(sRes.Data.OppHits) < limit { limit = len(sRes.Data.OppHits) }
	for i := 0; i < limit; i++ {
		hit := sRes.Data.OppHits[i]
		if c, ok := cache[hit.ID]; ok { grants = append(grants, c); continue }
		g := Grant{ID: hit.ID, Title: hit.Title, Agency: hit.Agency, Deadline: hit.CloseDate}
		detailsReq := url.Values{}; detailsReq.Set("oppId", hit.ID)
		if dResp, err := http.Post("https://apply07.grants.gov/grantsws/rest/opportunity/details", "application/x-www-form-urlencoded", strings.NewReader(detailsReq.Encode())); err == nil {
			if dResp != nil {
				var dRes GrantsGovDetailsResponse
				if err := json.NewDecoder(dResp.Body).Decode(&dRes); err == nil {
					g.Description = dRes.Synopsis.SynDesc; g.Amount = dRes.Synopsis.EstimatedFunding; g.Eligibility = dRes.Synopsis.ApplicantEligibilityDesc
				}
				dResp.Body.Close()
			}
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
	dsReq := map[string]interface{}{"model": "deepseek-chat", "messages": []map[string]string{{"role": "system", "content": "Return JSON."}, {"role": "user", "content": prompt}}, "response_format": map[string]string{"type": "json_object"}}
	dsBody, _ := json.Marshal(dsReq)
	client := &http.Client{}
	httpReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	httpReq.Header.Set("Authorization", "Bearer "+apiKey); httpReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(httpReq); if err != nil { http.Error(w, "api failed", 500); return }
	if resp != nil { defer resp.Body.Close() }
	var dsRes struct { Choices []struct { Message struct { Content string } }; Usage struct { TotalTokens int } }
	json.NewDecoder(resp.Body).Decode(&dsRes); LogUsage(dsRes.Usage.TotalTokens)
	globalLedger.RecordCost("ai_inference", float64(dsRes.Usage.TotalTokens)*0.000002, "Draft for "+req.GrantID)
	var draft ApplicationDraft; json.Unmarshal([]byte(dsRes.Choices[0].Message.Content), &draft)
	appID := generateID(); draft.ApplicationID = appID; draft.GrantID = req.GrantID; draft.Status = "draft_generated"
	appPath := filepath.Join(appsDir, appID+".json")
	appData, _ := json.Marshal(draft); os.WriteFile(appPath, appData, 0644)
	globalGraph.AddMeeting(Meeting{Summary: "Drafted application", Decisions: []string{"Apply to " + grant.ID}, ActionItems: []string{"Review " + appID}})
	globalSemantic.AddItem(dsRes.Choices[0].Message.Content, appID)
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(draft)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Query().Get("application_id"))
	data, err := os.ReadFile(filepath.Join(appsDir, id+".json"))
	if err != nil { http.Error(w, "not found", 404); return }
	w.Header().Set("Content-Type", "application/json"); w.Write(data)
}

func handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req struct { ApplicationID string; Status string }; json.NewDecoder(r.Body).Decode(&req)
	id := filepath.Base(req.ApplicationID); appPath := filepath.Join(appsDir, id+".json")
	data, _ := os.ReadFile(appPath); var d ApplicationDraft; json.Unmarshal(data, &d)
	prev := d.Status; d.Status = req.Status; updated, _ := json.Marshal(d); os.WriteFile(appPath, updated, 0644)
	if req.Status == "approved" && prev != "approved" { globalLedger.RecordRevenue(500.0, "Grant success: "+id) }
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(d)
}

func handleApplicationsList(w http.ResponseWriter, r *http.Request) {
	files, _ := os.ReadDir(appsDir); cacheMutex.Lock(); cache := make(map[string]Grant); d, _ := os.ReadFile(cachePath); json.Unmarshal(d, &cache); cacheMutex.Unlock()
	var res []ApplicationSummary
	for _, f := range files {
		if f.IsDir() { continue }
		data, _ := os.ReadFile(filepath.Join(appsDir, f.Name())); var dr ApplicationDraft; json.Unmarshal(data, &dr)
		s := ApplicationSummary{dr.ApplicationID, "", dr.Status, ""}
		if g, ok := cache[dr.GrantID]; ok { s.GrantTitle = g.Title; s.Deadline = g.Deadline }
		res = append(res, s)
	}
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(res)
}

func handleMonitor(w http.ResponseWriter, r *http.Request) {
	if checkKillSwitch() { http.Error(w, "kill-switch", 503); return }
	files, _ := os.ReadDir(appsDir); cacheMutex.Lock(); cache := make(map[string]Grant); d, _ := os.ReadFile(cachePath); json.Unmarshal(d, &cache); cacheMutex.Unlock()
	var reports []MonitorResult
	for _, f := range files {
		data, _ := os.ReadFile(filepath.Join(appsDir, f.Name())); var dr ApplicationDraft; json.Unmarshal(data, &dr)
		if dr.Status != "submitted" && dr.Status != "pending" { continue }
		g, ok := cache[dr.GrantID]; if !ok || g.Deadline == "" { continue }
		dl, _ := time.Parse("01/02/2006", g.Deadline)
		if time.Now().After(dl) && dr.FollowUpDraft == "" {
			if !rateLimit() { continue }
			apiKey := os.Getenv("DEEPSEEK_API_KEY")
			prompt := "Write follow-up for " + dr.ApplicationID
			dsReq := map[string]interface{}{"model": "deepseek-chat", "messages": []map[string]string{{"role": "user", "content": prompt}}}
			b, _ := json.Marshal(dsReq); client := &http.Client{}
			hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(b))
			hReq.Header.Set("Authorization", "Bearer "+apiKey); hReq.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(hReq)
			if err != nil { continue }
			if resp != nil {
				var dsRes struct { Choices []struct { Message struct { Content string } }; Usage struct { TotalTokens int } }
				json.NewDecoder(resp.Body).Decode(&dsRes); resp.Body.Close()
				LogUsage(dsRes.Usage.TotalTokens); globalLedger.RecordCost("ai_inference", float64(dsRes.Usage.TotalTokens)*0.000002, "Follow-up for "+dr.ApplicationID)
				dr.FollowUpDraft = dsRes.Choices[0].Message.Content; updated, _ := json.Marshal(dr); os.WriteFile(filepath.Join(appsDir, f.Name()), updated, 0644)
				reports = append(reports, MonitorResult{dr.ApplicationID, g.Title, dr.FollowUpDraft})
				AddAuditEntry("follow_up_generated", map[string]interface{}{"id": dr.ApplicationID})
			}
		}
	}
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(map[string]interface{}{"status": "complete", "follow_ups": reports})
}

func handleApplyAuto(w http.ResponseWriter, r *http.Request) {
	if checkKillSwitch() { http.Error(w, "kill-switch", 503); return }
	var req struct { URL string; FormData map[string]string; ApprovalID string }; json.NewDecoder(r.Body).Decode(&req)
	approvalMu.Lock(); ap, ok := approvalStore[req.ApprovalID]; approvalMu.Unlock()
	if !ok || ap.Status != "approved" || ap.Action != "grant_submit" { http.Error(w, "unauthorized", 403); return }
	AddAuditEntry("grant_submit_initiated", map[string]interface{}{"url": req.URL})
	globalLedger.RecordCost("browser_automation", 0.02, "Form submission")
	url := os.Getenv("BROWSER_AGENT_URL"); if url == "" { url = "https://koola10-browser.fly.dev" }
	b, _ := json.Marshal(req); resp, err := http.Post(url+"/browser/submit-form", "application/json", bytes.NewBuffer(b))
	if err != nil { http.Error(w, "failed", 500); return }
	if resp != nil { defer resp.Body.Close(); w.Header().Set("Content-Type", "application/json"); io.Copy(w, resp.Body) }
}

func handleCheckStatus(w http.ResponseWriter, r *http.Request) {
	if checkKillSwitch() { http.Error(w, "kill-switch", 503); return }
	var req struct { URL string; Instruction string }; json.NewDecoder(r.Body).Decode(&req)
	globalLedger.RecordCost("browser_automation", 0.02, "Status check")
	url := os.Getenv("BROWSER_AGENT_URL"); if url == "" { url = "https://koola10-browser.fly.dev" }
	b, _ := json.Marshal(req); resp, err := http.Post(url+"/browser/extract", "application/json", bytes.NewBuffer(b))
	if err != nil { http.Error(w, "failed", 500); return }
	if resp != nil { defer resp.Body.Close(); w.Header().Set("Content-Type", "application/json"); io.Copy(w, resp.Body) }
}

func handleAIChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest; json.NewDecoder(r.Body).Decode(&req)
	apiKey := os.Getenv("DEEPSEEK_API_KEY"); if apiKey == "" { http.Error(w, "no key", 500); return }
	if !rateLimit() { http.Error(w, "rate limited", 429); return }
	sys := "You are Koola10."
	res, err := globalSemantic.Search(req.Prompt, 3)
	if err == nil { for _, rs := range res { if rs.Score > 0.5 { sys += "\nContext: " + rs.Text } } }
	if strings.Contains(req.Prompt, "influence") { globalGraph.mu.RLock(); gd, _ := json.Marshal(globalGraph); sys += "\nGraph: " + string(gd); globalGraph.mu.RUnlock() }
	b, _ := json.Marshal(map[string]interface{}{"model": "deepseek-chat", "messages": []map[string]string{{"role": "system", "content": sys}, {"role": "user", "content": req.Prompt}}})
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(b))
	hReq.Header.Set("Authorization", "Bearer "+apiKey); hReq.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(hReq); if err != nil { http.Error(w, "failed", 500); return }
	if resp != nil {
		defer resp.Body.Close()
		var dRes struct { Choices []struct { Message struct { Content string } }; Usage struct { TotalTokens int } }
		json.NewDecoder(resp.Body).Decode(&dRes); LogUsage(dRes.Usage.TotalTokens)
		globalLedger.RecordCost("ai_inference", float64(dRes.Usage.TotalTokens)*0.000002, "Chat response")
		w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(ChatResponse{dRes.Choices[0].Message.Content, dRes.Usage.TotalTokens})
	}
}

func handleAIRemember(w http.ResponseWriter, r *http.Request) {
	var req MemoryEntry; json.NewDecoder(r.Body).Decode(&req)
	cacheMutex.Lock(); mem := make(map[string]string); d, _ := os.ReadFile(memoryPath); json.Unmarshal(d, &mem); mem[req.Key] = req.Value
	md, _ := json.Marshal(mem); os.WriteFile(memoryPath, md, 0644); cacheMutex.Unlock()
	json.NewEncoder(w).Encode(map[string]string{"status": "stored"})
}

func handleAIRecall(w http.ResponseWriter, r *http.Request) {
	k := r.URL.Query().Get("key"); cacheMutex.Lock(); mem := make(map[string]string); d, _ := os.ReadFile(memoryPath); json.Unmarshal(d, &mem); cacheMutex.Unlock()
	v, ok := mem[k]; if !ok { http.Error(w, "not found", 404); return }
	json.NewEncoder(w).Encode(map[string]string{"key": k, "value": v})
}

func handleAIAnalyzeGrant(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeGrantRequest; json.NewDecoder(r.Body).Decode(&req)
	apiKey := os.Getenv("DEEPSEEK_API_KEY"); if !rateLimit() { http.Error(w, "rate limited", 429); return }
	p := "Analyze: " + req.GrantText
	b, _ := json.Marshal(map[string]interface{}{"model": "deepseek-chat", "messages": []map[string]string{{"role": "system", "content": "Return JSON."}, {"role": "user", "content": p}}, "response_format": map[string]string{"type": "json_object"}})
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(b))
	hReq.Header.Set("Authorization", "Bearer "+apiKey); hReq.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(hReq); if err != nil { http.Error(w, "failed", 500); return }
	if resp != nil {
		defer resp.Body.Close()
		var dRes struct { Choices []struct { Message struct { Content string } }; Usage struct { TotalTokens int } }
		json.NewDecoder(resp.Body).Decode(&dRes); LogUsage(dRes.Usage.TotalTokens)
		globalLedger.RecordCost("ai_inference", float64(dRes.Usage.TotalTokens)*0.000002, "Analysis")
		var an AnalyzeGrantResponse; json.Unmarshal([]byte(dRes.Choices[0].Message.Content), &an)
		globalGraph.AddMeeting(Meeting{Summary: "Analyzed grant", Decisions: []string{fmt.Sprintf("Score: %d", an.EligibilityScore)}, ActionItems: an.RequiredDocuments})
		globalSemantic.AddItem(req.GrantText, "analysis_"+generateID())
		w.Header().Set("Content-Type", "application/json"); w.Write([]byte(dRes.Choices[0].Message.Content))
	}
}

func handleMemoryMeetings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" { var m Meeting; json.NewDecoder(r.Body).Decode(&m); id := globalGraph.AddMeeting(m); json.NewEncoder(w).Encode(map[string]string{"id": id}); return }
	globalGraph.mu.RLock(); var res []Meeting; for _, m := range globalGraph.Meetings { res = append(res, m) }; globalGraph.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(res)
}

func handleMemoryEntity(w http.ResponseWriter, r *http.Request) {
	n := strings.TrimPrefix(r.URL.Path, "/memory/entity/"); globalGraph.mu.RLock(); e, ok := globalGraph.Entities[n]; globalGraph.mu.RUnlock()
	if !ok { http.Error(w, "not found", 404); return }
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(e)
}

func handleMemoryInfluence(w http.ResponseWriter, r *http.Request) {
	n := strings.TrimPrefix(r.URL.Path, "/memory/influence/"); s := globalGraph.CalculateInfluenceScore(n)
	json.NewEncoder(w).Encode(map[string]interface{}{"name": n, "score": s})
}

func handleMemoryPath(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("source"); t := r.URL.Query().Get("target"); p := globalGraph.FindPath(s, t, 5)
	json.NewEncoder(w).Encode(p)
}

func handleMemoryDecisionsRanked(w http.ResponseWriter, r *http.Request) {
	res := globalGraph.RankDecisionsByImpact(); json.NewEncoder(w).Encode(res)
}

func handleSemanticIndex(w http.ResponseWriter, r *http.Request) {
	var req struct { Text string; RefID string }; json.NewDecoder(r.Body).Decode(&req); globalSemantic.AddItem(req.Text, req.RefID)
	json.NewEncoder(w).Encode(map[string]string{"status": "indexed"})
}

func handleSemanticSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q"); res, _ := globalSemantic.Search(q, 5); json.NewEncoder(w).Encode(res)
}

func handleComplianceAudit(w http.ResponseWriter, r *http.Request) {
	auditMutex.Lock(); defer auditMutex.Unlock(); f, _ := os.Open(auditPath); defer f.Close()
	var res []AuditEntry; s := bufio.NewScanner(f); for s.Scan() { var e AuditEntry; json.Unmarshal(s.Bytes(), &e); res = append(res, e) }
	if len(res) > 200 { res = res[len(res)-200:] }
	json.NewEncoder(w).Encode(res)
}

func handleComplianceAuditVerify(w http.ResponseWriter, r *http.Request) {
	auditMutex.Lock(); defer auditMutex.Unlock(); f, _ := os.Open(auditPath); defer f.Close()
	v := true; lh := "0000000000000000000000000000000000000000000000000000000000000000"
	sc := bufio.NewScanner(f); for sc.Scan() {
		var e AuditEntry; json.Unmarshal(sc.Bytes(), &e); ph := e.Hash; e.Hash = ""; ej, _ := json.Marshal(e)
		h := sha256.New(); h.Write([]byte(lh + string(ej))); ch := hex.EncodeToString(h.Sum(nil))
		if ph != ch { v = false; break }; lh = ph
	}
	json.NewEncoder(w).Encode(map[string]bool{"valid": v})
}

func handleComplianceApproval(w http.ResponseWriter, r *http.Request) {
	var req ApprovalRequest; json.NewDecoder(r.Body).Decode(&req); req.ID = generateID(); req.Status = "pending"; req.CreatedAt = time.Now().Format(time.RFC3339)
	ev := EvaluateAction(req.Action, 0.0); if ev.Decision == "warn" { req.Justification = "LOW ROI" }
	approvalMu.Lock(); approvalStore[req.ID] = &req; approvalMu.Unlock()
	AddAuditEntry("approval_created", map[string]interface{}{"id": req.ID, "action": req.Action})
	json.NewEncoder(w).Encode(req)
}

func handleComplianceApprove(w http.ResponseWriter, r *http.Request) {
	var req struct { ApprovalID string; Approver string }; json.NewDecoder(r.Body).Decode(&req)
	approvalMu.Lock(); ap, ok := approvalStore[req.ApprovalID]; approvalMu.Unlock()
	if !ok { http.Error(w, "not found", 404); return }
	ap.Status = "approved"; ap.Approver = req.Approver; AddAuditEntry("approved", map[string]interface{}{"id": req.ApprovalID})
	json.NewEncoder(w).Encode(ap)
}

func handleComplianceKillSwitch(w http.ResponseWriter, r *http.Request) {
	killSwitchMu.Lock(); os.WriteFile(killSwitchPath, []byte("active"), 0644); killSwitchMu.Unlock()
	AddAuditEntry("kill_activated", nil); w.Write([]byte("Active"))
}

func handleComplianceKillSwitchReset(w http.ResponseWriter, r *http.Request) {
	killSwitchMu.Lock(); os.Remove(killSwitchPath); killSwitchMu.Unlock()
	AddAuditEntry("kill_reset", nil); w.Write([]byte("Reset"))
}

func handleComplianceUsage(w http.ResponseWriter, r *http.Request) {
	usageMutex.Lock(); defer usageMutex.Unlock(); f, _ := os.Open(usagePath); defer f.Close()
	var tt int; var tc float64; cutoff := time.Now().Add(-24 * time.Hour)
	sc := bufio.NewScanner(f); for sc.Scan() {
		var l UsageLog; json.Unmarshal(sc.Bytes(), &l); ts, _ := time.Parse(time.RFC3339, l.Timestamp)
		if ts.After(cutoff) { tt += l.TokensUsed; tc += l.Cost }
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"total_tokens": tt, "total_cost": tc})
}

func handleEconomicLedgerCost(w http.ResponseWriter, r *http.Request) {
	var req struct { Category string; Amount float64; Description string }; json.NewDecoder(r.Body).Decode(&req)
	globalLedger.RecordCost(req.Category, req.Amount, req.Description); w.WriteHeader(201)
}

func handleEconomicLedgerRevenue(w http.ResponseWriter, r *http.Request) {
	var req struct { Amount float64; Source string }; json.NewDecoder(r.Body).Decode(&req)
	globalLedger.RecordRevenue(req.Amount, req.Source); w.WriteHeader(201)
}

func handleEconomicLedgerSummary(w http.ResponseWriter, r *http.Request) {
	globalLedger.mu.RLock(); defer globalLedger.mu.RUnlock()
	roi := 0.0; if globalLedger.TotalCosts > 0 { roi = globalLedger.TotalRevenue / globalLedger.TotalCosts }
	json.NewEncoder(w).Encode(EconomicSummary{globalLedger.Balance, globalLedger.TotalCosts, globalLedger.TotalRevenue, roi})
}

func handleEconomicEvaluate(w http.ResponseWriter, r *http.Request) {
	var req struct { ActionType string `json:"action_type"`; EstimatedCost float64 `json:"estimated_cost"` }; json.NewDecoder(r.Body).Decode(&req)
	eval := EvaluateAction(req.ActionType, req.EstimatedCost); json.NewEncoder(w).Encode(eval)
}

func generateID() string {
	b := make([]byte, 8); rand.Read(b); return hex.EncodeToString(b)
}
