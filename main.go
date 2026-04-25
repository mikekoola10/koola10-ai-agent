package main

import (
	"bytes"
	"crypto/rand"
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

// Grant represents the unified grant structure
type Grant struct {
	ID          string `json:"grant_id"`
	Title       string `json:"title"`
	Agency      string `json:"agency"`
	Deadline    string `json:"deadline"`
	Amount      string `json:"amount"`
	Eligibility string `json:"eligibility"`
	Description string `json:"description"`
}

// GrantsGovSearchResponse for parsing search2 results
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

// GrantsGovDetailsResponse for parsing details results
type GrantsGovDetailsResponse struct {
	Synopsis struct {
		SynDesc                  string `json:"synopsisDesc"`
		EstimatedFunding         string `json:"estimatedFunding"`
		ApplicantEligibilityDesc string `json:"applicantEligibilityDesc"`
	} `json:"synopsis"`
}

// ApplyRequest represents the incoming POST body for /grants/apply
type ApplyRequest struct {
	GrantID    string `json:"grant_id"`
	OrgName    string `json:"org_name"`
	OrgMission string `json:"org_mission"`
	OrgBudget  string `json:"org_budget"`
	OrgTaxID   string `json:"org_tax_id"`
}

// ApplicationDraft represents the stored application
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

// ApplicationSummary for the list endpoint
type ApplicationSummary struct {
	ApplicationID string `json:"application_id"`
	GrantTitle    string `json:"grant_title"`
	Status        string `json:"status"`
	Deadline      string `json:"deadline"`
}

// MonitorResult represents a generated follow-up
type MonitorResult struct {
	ApplicationID string `json:"application_id"`
	GrantTitle    string `json:"grant_title"`
	FollowUpEmail string `json:"follow_up_email"`
}

// AI structs
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

// Graph Memory Structs
type Meeting struct {
	MeetingID   string   `json:"meeting_id"`
	Timestamp   string   `json:"timestamp"`
	Summary     string   `json:"summary"`
	Decisions   []string `json:"decisions"`
	ActionItems []string `json:"action_items"`
}

type Entity struct {
	Name  string   `json:"name"`
	Type  string   `json:"type"` // "person", "decision", "task"
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

var (
	cacheMutex sync.Mutex
	cachePath  = "/data/grants_cache.json"
	appsDir    = "/data/applications"
	memoryPath = "/data/memory.json"
	graphPath  = "/data/memory_graph.json"
	globalGraph = &MemoryGraph{
		Meetings: make(map[string]Meeting),
		Entities: make(map[string]Entity),
		Edges:    []Edge{},
	}
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		log.Printf("failed to create cache dir: %v", err)
	}
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		log.Printf("failed to create apps dir: %v", err)
	}

	// Load graph if exists
	globalGraph.Load()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/grants/search", handleSearch)
	http.HandleFunc("/grants/apply", handleApply)
	http.HandleFunc("/grants/status", handleStatus)
	http.HandleFunc("/grants/applications", handleApplicationsList)
	http.HandleFunc("/grants/monitor", handleMonitor)
	http.HandleFunc("/grants/update-status", handleUpdateStatus)
	http.HandleFunc("/grants/apply-auto", handleApplyAuto)
	http.HandleFunc("/grants/check-status", handleCheckStatus)

	// AI Endpoints
	http.HandleFunc("/ai/chat", handleAIChat)
	http.HandleFunc("/ai/remember", handleAIRemember)
	http.HandleFunc("/ai/recall", handleAIRecall)
	http.HandleFunc("/ai/analyze-grant", handleAIAnalyzeGrant)

	// Memory Graph Endpoints
	http.HandleFunc("/memory/meetings", handleMemoryMeetings)
	http.HandleFunc("/memory/entity/", handleMemoryEntity)
	http.HandleFunc("/memory/influence/", handleMemoryInfluence)
	http.HandleFunc("/memory/path", handleMemoryPath)
	http.HandleFunc("/memory/decisions/ranked", handleMemoryDecisionsRanked)

	log.Printf("starting server on 0.0.0.0:%s", port)

	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		log.Fatalf("server_failed: %v", err)
	}
}

// Graph methods
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
			if g.Edges[i].Weight > 2.0 {
				g.Edges[i].Weight = 2.0
			}
			g.Edges[i].Frequency++
			return
		}
	}

	g.Edges = append(g.Edges, Edge{
		Source:    source,
		Target:    target,
		Relation:  relation,
		Weight:    1.0,
		Frequency: 1,
		Metadata:  metadata,
	})
}

func (g *MemoryGraph) AddMeeting(m Meeting) string {
	if m.MeetingID == "" {
		m.MeetingID = generateID()
	}
	if m.Timestamp == "" {
		m.Timestamp = time.Now().Format(time.RFC3339)
	}

	g.mu.Lock()
	g.Meetings[m.MeetingID] = m
	g.mu.Unlock()

	for _, decision := range m.Decisions {
		g.mu.Lock()
		if _, ok := g.Entities[decision]; !ok {
			g.Entities[decision] = Entity{Name: decision, Type: "decision"}
		}
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
			if !ok {
				entity = Entity{Name: owner, Type: "person"}
			}
			entity.Tasks = append(entity.Tasks, taskName)
			g.Entities[owner] = entity
			g.mu.Unlock()

			g.AddWeightedEdge(owner, taskName, "assigned_to", nil)
		}

		g.mu.Lock()
		if _, ok := g.Entities[taskName]; !ok {
			g.Entities[taskName] = Entity{Name: taskName, Type: "task"}
		}
		g.mu.Unlock()
		g.AddWeightedEdge(m.MeetingID, taskName, "contains_task", nil)
	}

	g.Save()
	return m.MeetingID
}

func (g *MemoryGraph) CalculateInfluenceScore(name string) float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var incomingWeight, outgoingWeight float64
	var totalEdges int

	for _, edge := range g.Edges {
		if edge.Target == name {
			incomingWeight += edge.Weight
			totalEdges++
		}
		if edge.Source == name {
			outgoingWeight += edge.Weight
			totalEdges++
		}
	}

	if totalEdges == 0 {
		return 0
	}

	score := (incomingWeight * 0.7) + (outgoingWeight * 0.3)
	return score / float64(totalEdges)
}

func (g *MemoryGraph) FindPath(source, target string, maxDepth int) []Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	type pathNode struct {
		entity string
		path   []Edge
	}

	queue := []pathNode{{entity: source, path: []Edge{}}}
	visited := make(map[string]bool)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.entity == target {
			return current.path
		}

		if len(current.path) >= maxDepth {
			continue
		}

		visited[current.entity] = true

		for _, edge := range g.Edges {
			if edge.Source == current.entity && !visited[edge.Target] {
				newPath := make([]Edge, len(current.path))
				copy(newPath, current.path)
				newPath = append(newPath, edge)
				queue = append(queue, pathNode{entity: edge.Target, path: newPath})
			}
		}
	}

	return nil
}

func (g *MemoryGraph) RankDecisionsByImpact() []string {
	g.mu.RLock()
	decisions := []string{}
	for name, entity := range g.Entities {
		if entity.Type == "decision" {
			decisions = append(decisions, name)
		}
	}
	g.mu.RUnlock()

	sort.Slice(decisions, func(i, j int) bool {
		return g.CalculateInfluenceScore(decisions[i]) > g.CalculateInfluenceScore(decisions[j])
	})

	return decisions
}

// Handler Implementations
func handleMemoryMeetings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var m Meeting
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		id := globalGraph.AddMeeting(m)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"meeting_id": id})
		return
	}

	globalGraph.mu.RLock()
	defer globalGraph.mu.RUnlock()
	meetings := []Meeting{}
	for _, m := range globalGraph.Meetings {
		meetings = append(meetings, m)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meetings)
}

func handleMemoryEntity(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/memory/entity/")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	globalGraph.mu.RLock()
	defer globalGraph.mu.RUnlock()
	entity, ok := globalGraph.Entities[name]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

func handleMemoryInfluence(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/memory/influence/")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	score := globalGraph.CalculateInfluenceScore(name)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"name": name, "influence_score": score})
}

func handleMemoryPath(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")
	target := r.URL.Query().Get("target")
	if source == "" || target == "" {
		http.Error(w, "source and target required", http.StatusBadRequest)
		return
	}
	path := globalGraph.FindPath(source, target, 5)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(path)
}

func handleMemoryDecisionsRanked(w http.ResponseWriter, r *http.Request) {
	ranked := globalGraph.RankDecisionsByImpact()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ranked)
}

// Rest of the grant handlers... (keep them as they were, but I'll integrate them in next steps)
func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	category := r.URL.Query().Get("category")

	searchReq := map[string]interface{}{
		"keyword": query,
	}
	if category != "" {
		searchReq["fundingCategories"] = category
	}

	reqBody, err := json.Marshal(searchReq)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	resp, err := http.Post("https://api.grants.gov/v1/api/search2", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w, "failed to search grants", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var searchRes GrantsGovSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchRes); err != nil {
		http.Error(w, "failed to parse search results", http.StatusInternalServerError)
		return
	}

	var grants []Grant
	enrichedCache := make(map[string]Grant)

	// Load existing cache
	cacheMutex.Lock()
	if data, err := os.ReadFile(cachePath); err == nil {
		json.Unmarshal(data, &enrichedCache)
	}
	cacheMutex.Unlock()

	limit := 5
	if len(searchRes.Data.OppHits) < limit {
		limit = len(searchRes.Data.OppHits)
	}

	for i := 0; i < limit; i++ {
		hit := searchRes.Data.OppHits[i]
		if cached, ok := enrichedCache[hit.ID]; ok {
			grants = append(grants, cached)
			continue
		}

		// Fetch details
		grant := Grant{
			ID:       hit.ID,
			Title:    hit.Title,
			Agency:   hit.Agency,
			Deadline: hit.CloseDate,
		}

		detailsReq := url.Values{}
		detailsReq.Set("oppId", hit.ID)
		dResp, err := http.Post("https://apply07.grants.gov/grantsws/rest/opportunity/details", "application/x-www-form-urlencoded", strings.NewReader(detailsReq.Encode()))
		if err == nil {
			var dRes GrantsGovDetailsResponse
			if err := json.NewDecoder(dResp.Body).Decode(&dRes); err == nil {
				grant.Description = dRes.Synopsis.SynDesc
				grant.Amount = dRes.Synopsis.EstimatedFunding
				grant.Eligibility = dRes.Synopsis.ApplicantEligibilityDesc
			}
			dResp.Body.Close()
		}

		grants = append(grants, grant)
		enrichedCache[hit.ID] = grant
	}

	// Save cache
	cacheMutex.Lock()
	if cacheData, err := json.Marshal(enrichedCache); err == nil {
		os.WriteFile(cachePath, cacheData, 0644)
	}
	cacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grants)
}

func handleApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Load cache
	cacheMutex.Lock()
	enrichedCache := make(map[string]Grant)
	if data, err := os.ReadFile(cachePath); err == nil {
		json.Unmarshal(data, &enrichedCache)
	}
	cacheMutex.Unlock()

	grant, ok := enrichedCache[req.GrantID]
	if !ok {
		http.Error(w, "grant not found in cache. please search for it first.", http.StatusNotFound)
		return
	}

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		http.Error(w, "DEEPSEEK_API_KEY not set", http.StatusInternalServerError)
		return
	}

	prompt := fmt.Sprintf(`Generate a complete narrative grant application draft for the following grant and organization.
Grant Title: %s
Agency: %s
Description: %s
Amount: %s
Eligibility: %s

Organization Name: %s
Mission: %s
Annual Budget: %s
Tax ID: %s

Provide the response in JSON format with the following keys: executive_summary, statement_of_need, project_description, budget_justification, organizational_capacity.`,
		grant.Title, grant.Agency, grant.Description, grant.Amount, grant.Eligibility,
		req.OrgName, req.OrgMission, req.OrgBudget, req.OrgTaxID)

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a professional grant writer. Return ONLY JSON."},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	dsBody, err := json.Marshal(dsReq)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	client := &http.Client{}
	httpReq, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, "failed to call DeepSeek API", http.StatusInternalServerError)
		return
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
		http.Error(w, "failed to parse DeepSeek response", http.StatusInternalServerError)
		return
	}

	if len(dsRes.Choices) == 0 {
		http.Error(w, "no response from DeepSeek", http.StatusInternalServerError)
		return
	}

	var draft ApplicationDraft
	if err := json.Unmarshal([]byte(dsRes.Choices[0].Message.Content), &draft); err != nil {
		// Fallback if JSON is weird
		draft.ExecutiveSummary = dsRes.Choices[0].Message.Content
	}

	appID := generateID()
	draft.ApplicationID = appID
	draft.GrantID = req.GrantID
	draft.Status = "draft_generated"

	// Save draft
	appPath := filepath.Join(appsDir, appID+".json")
	if appData, err := json.Marshal(draft); err == nil {
		os.WriteFile(appPath, appData, 0644)
	}

	// Record in graph
	globalGraph.AddMeeting(Meeting{
		Summary:   fmt.Sprintf("Drafted application for grant: %s", grant.Title),
		Decisions: []string{fmt.Sprintf("Apply to %s", grant.ID)},
		ActionItems: []string{fmt.Sprintf("System: Review %s draft", appID)},
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(draft)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	appID := r.URL.Query().Get("application_id")
	if appID == "" {
		http.Error(w, "application_id is required", http.StatusBadRequest)
		return
	}

	// Sanitize to prevent path traversal
	safeID := filepath.Base(appID)
	appPath := filepath.Join(appsDir, safeID+".json")
	data, err := os.ReadFile(appPath)
	if err != nil {
		http.Error(w, "application not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ApplicationID string `json:"application_id"`
		Status        string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	safeID := filepath.Base(req.ApplicationID)
	appPath := filepath.Join(appsDir, safeID+".json")
	data, err := os.ReadFile(appPath)
	if err != nil {
		http.Error(w, "application not found", http.StatusNotFound)
		return
	}

	var draft ApplicationDraft
	if err := json.Unmarshal(data, &draft); err != nil {
		http.Error(w, "failed to parse application data", http.StatusInternalServerError)
		return
	}

	draft.Status = req.Status
	updatedData, _ := json.Marshal(draft)
	os.WriteFile(appPath, updatedData, 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(draft)
}

func handleApplicationsList(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(appsDir)
	if err != nil {
		http.Error(w, "failed to read applications", http.StatusInternalServerError)
		return
	}

	// Load cache for titles and deadlines
	cacheMutex.Lock()
	enrichedCache := make(map[string]Grant)
	if data, err := os.ReadFile(cachePath); err == nil {
		json.Unmarshal(data, &enrichedCache)
	}
	cacheMutex.Unlock()

	var summaries []ApplicationSummary
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			data, err := os.ReadFile(filepath.Join(appsDir, file.Name()))
			if err != nil {
				continue
			}
			var draft ApplicationDraft
			if err := json.Unmarshal(data, &draft); err == nil {
				summary := ApplicationSummary{
					ApplicationID: draft.ApplicationID,
					Status:        draft.Status,
				}
				if grant, ok := enrichedCache[draft.GrantID]; ok {
					summary.GrantTitle = grant.Title
					summary.Deadline = grant.Deadline
				}
				summaries = append(summaries, summary)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

func handleMonitor(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(appsDir)
	if err != nil {
		http.Error(w, "failed to read applications", http.StatusInternalServerError)
		return
	}

	cacheMutex.Lock()
	enrichedCache := make(map[string]Grant)
	if data, err := os.ReadFile(cachePath); err == nil {
		json.Unmarshal(data, &enrichedCache)
	}
	cacheMutex.Unlock()

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	var reports []MonitorResult

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		appPath := filepath.Join(appsDir, file.Name())
		data, err := os.ReadFile(appPath)
		if err != nil {
			continue
		}

		var draft ApplicationDraft
		if err := json.Unmarshal(data, &draft); err != nil {
			continue
		}

		// Also check draft_generated if they haven't manually updated status but we still want to monitor?
		// User said: "For each application with status 'submitted' or 'pending'"
		if draft.Status != "submitted" && draft.Status != "pending" {
			continue
		}

		grant, ok := enrichedCache[draft.GrantID]
		if !ok || grant.Deadline == "" {
			continue
		}

		// Parse deadline: "MM/DD/YYYY"
		deadline, err := time.Parse("01/02/2006", grant.Deadline)
		if err != nil {
			continue
		}

		if time.Now().After(deadline) && draft.FollowUpDraft == "" && apiKey != "" {
			// Generate follow-up
			prompt := fmt.Sprintf("Write a polite, professional follow-up email to the agency '%s' regarding our application for the grant '%s' (ID: %s). The deadline has passed and we are checking on the status.", grant.Agency, grant.Title, grant.ID)

			dsReq := map[string]interface{}{
				"model": "deepseek-chat",
				"messages": []map[string]string{
					{"role": "system", "content": "You are a professional grant consultant. Return ONLY the email draft text."},
					{"role": "user", "content": prompt},
				},
			}
			dsBody, _ := json.Marshal(dsReq)
			client := &http.Client{}
			httpReq, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
			if err != nil {
				continue
			}
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("Authorization", "Bearer "+apiKey)

			resp, err := client.Do(httpReq)
			if err == nil {
				var dsRes struct {
					Choices []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					} `json:"choices"`
				}
				if json.NewDecoder(resp.Body).Decode(&dsRes) == nil && len(dsRes.Choices) > 0 {
					draft.FollowUpDraft = dsRes.Choices[0].Message.Content
					updatedData, _ := json.Marshal(draft)
					os.WriteFile(appPath, updatedData, 0644)
					reports = append(reports, MonitorResult{
						ApplicationID: draft.ApplicationID,
						GrantTitle:    grant.Title,
						FollowUpEmail: draft.FollowUpDraft,
					})
				}
				resp.Body.Close()
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "scan_complete",
		"follow_ups": reports,
	})
}

func handleApplyAuto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL      string            `json:"url"`
		FormData map[string]string `json:"form_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	browserAgentURL := os.Getenv("BROWSER_AGENT_URL")
	if browserAgentURL == "" {
		browserAgentURL = "https://koola10-browser.fly.dev"
	}

	jsonBody, _ := json.Marshal(req)
	resp, err := http.Post(browserAgentURL+"/browser/submit-form", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		http.Error(w, "failed to call browser agent: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func handleCheckStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL         string `json:"url"`
		Instruction string `json:"instruction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	browserAgentURL := os.Getenv("BROWSER_AGENT_URL")
	if browserAgentURL == "" {
		browserAgentURL = "https://koola10-browser.fly.dev"
	}

	jsonBody, _ := json.Marshal(req)
	resp, err := http.Post(browserAgentURL+"/browser/extract", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		http.Error(w, "failed to call browser agent: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func handleAIChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		http.Error(w, "DEEPSEEK_API_KEY not set", http.StatusInternalServerError)
		return
	}

	systemPrompt := "You are Koola10, an autonomous AI agent for Spiral Grant Services. You are an expert in federal grants and help users find, analyze, and apply for funding."

	// Add graph context
	if strings.Contains(req.Prompt, "influence") || strings.Contains(req.Prompt, "path") || strings.Contains(req.Prompt, "decision") {
		globalGraph.mu.RLock()
		graphData, _ := json.Marshal(globalGraph)
		systemPrompt += "\n\nCurrent Memory Graph Data: " + string(graphData)
		globalGraph.mu.RUnlock()
	}

	if req.Context != "" {
		systemPrompt += "\n\nContext Memory: " + req.Context
	}

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": req.Prompt},
		},
	}

	dsBody, _ := json.Marshal(dsReq)
	client := &http.Client{}
	httpReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, "failed to call DeepSeek API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil {
		http.Error(w, "failed to parse DeepSeek response", http.StatusInternalServerError)
		return
	}

	if len(dsRes.Choices) == 0 {
		http.Error(w, "no response from DeepSeek", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{
		Response:   dsRes.Choices[0].Message.Content,
		TokensUsed: dsRes.Usage.TotalTokens,
	})
}

func handleAIRemember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MemoryEntry
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	memory := make(map[string]string)
	if data, err := os.ReadFile(memoryPath); err == nil {
		json.Unmarshal(data, &memory)
	}

	memory[req.Key] = req.Value
	memoryData, _ := json.Marshal(memory)
	os.WriteFile(memoryPath, memoryData, 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stored"})
}

func handleAIRecall(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	memory := make(map[string]string)
	if data, err := os.ReadFile(memoryPath); err == nil {
		json.Unmarshal(data, &memory)
	}

	val, ok := memory[key]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})
}

func handleAIAnalyzeGrant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeGrantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		http.Error(w, "DEEPSEEK_API_KEY not set", http.StatusInternalServerError)
		return
	}

	orgProfile, _ := json.Marshal(req.OrgProfile)
	prompt := fmt.Sprintf(`Analyze the following grant text relative to the provided organizational profile.
Grant Text: %s
Org Profile: %s

Extract structured data as JSON with the following keys:
- eligibility_score: a number from 1 to 100 indicating fit.
- key_deadlines: a list of important dates found in the text.
- required_documents: a list of documents needed for application.
- summary: a one-paragraph professional summary of the opportunity.`, req.GrantText, string(orgProfile))

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "You are an expert grant analyst. Return ONLY JSON."},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	dsBody, _ := json.Marshal(dsReq)
	client := &http.Client{}
	httpReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, "failed to call DeepSeek API", http.StatusInternalServerError)
		return
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
		http.Error(w, "failed to parse DeepSeek response", http.StatusInternalServerError)
		return
	}

	if len(dsRes.Choices) == 0 {
		http.Error(w, "no response from DeepSeek", http.StatusInternalServerError)
		return
	}

	analysisJSON := dsRes.Choices[0].Message.Content
	var analysis AnalyzeGrantResponse
	json.Unmarshal([]byte(analysisJSON), &analysis)

	// Record in graph
	globalGraph.AddMeeting(Meeting{
		Summary:   fmt.Sprintf("Analyzed grant: %s", analysis.Summary),
		Decisions: []string{fmt.Sprintf("Grant fit score: %d", analysis.EligibilityScore)},
		ActionItems: analysis.RequiredDocuments,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(analysisJSON))
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
