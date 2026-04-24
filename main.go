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

var (
	cacheMutex sync.Mutex
	cachePath  = "/data/grants_cache.json"
	appsDir    = "/data/applications"
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

	log.Printf("starting server on 0.0.0.0:%s", port)

	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		log.Fatalf("server_failed: %v", err)
	}
}

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

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
