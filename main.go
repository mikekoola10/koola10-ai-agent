package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	os.MkdirAll(filepath.Dir(cachePath), 0755)
	os.MkdirAll(appsDir, 0755)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/grants/search", handleSearch)
	http.HandleFunc("/grants/apply", handleApply)
	http.HandleFunc("/grants/status", handleStatus)

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

	body, _ := json.Marshal(searchReq)
	resp, err := http.Post("https://api.grants.gov/v1/api/search2", "application/json", bytes.NewBuffer(body))
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
	cacheData, _ := json.Marshal(enrichedCache)
	os.WriteFile(cachePath, cacheData, 0644)
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
		http.Error(w, "grant not found in cache", http.StatusNotFound)
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
	appData, _ := json.Marshal(draft)
	os.WriteFile(appPath, appData, 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(draft)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	appID := r.URL.Query().Get("application_id")
	if appID == "" {
		http.Error(w, "application_id is required", http.StatusBadRequest)
		return
	}

	appPath := filepath.Join(appsDir, appID+".json")
	data, err := os.ReadFile(appPath)
	if err != nil {
		http.Error(w, "application not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
