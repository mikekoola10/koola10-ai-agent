package cronjob

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Client singleton
var defaultClient *Client

func GetClient() *Client {
	if defaultClient == nil {
		defaultClient = NewClient()
	}
	return defaultClient
}

// HandleListJobs lists all cron jobs
func HandleListJobs(w http.ResponseWriter, r *http.Request) {
	c := GetClient()
	if !c.IsConfigured() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "CRONJOB_API_KEY not configured",
			"message": "Add CRONJOB_API_KEY to your Render environment variables. Get your key at https://cron-job.org/console/settings",
		})
		return
	}

	resp, err := c.ListJobs()
	if err != nil {
		log.Printf("[CronJob] ListJobs error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleGetJob returns details for a specific job
func HandleGetJob(w http.ResponseWriter, r *http.Request) {
	c := GetClient()
	if !c.IsConfigured() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "CRONJOB_API_KEY not configured"})
		return
	}

	jobID, err := strconv.Atoi(r.URL.Query().Get("jobId"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "jobId query parameter required"})
		return
	}

	resp, err := c.GetJob(jobID)
	if err != nil {
		log.Printf("[CronJob] GetJob(%d) error: %v", jobID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleCreateJob creates a new cron job
func HandleCreateJob(w http.ResponseWriter, r *http.Request) {
	c := GetClient()
	if !c.IsConfigured() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "CRONJOB_API_KEY not configured",
			"message": "Add CRONJOB_API_KEY to your Render environment variables.",
		})
		return
	}

	var req struct {
		Title     string `json:"title"`
		URL       string `json:"url"`
		Frequency string `json:"frequency"`
		Method    string `json:"method"` // GET or POST
		Headers   map[string]string `json:"headers"`
		Body      string `json:"body"`
		Timezone  string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	if req.URL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "url is required"})
		return
	}
	if req.Frequency == "" {
		req.Frequency = "hourly"
	}
	if req.Timezone == "" {
		req.Timezone = "America/New_York"
	}
	if req.Method == "" {
		req.Method = "GET"
	}
	if req.Title == "" {
		req.Title = "Koola10 Cron Job"
	}

	methodInt := 0 // GET
	if strings.ToUpper(req.Method) == "POST" {
		methodInt = 1
	}

	schedule := buildSchedule(req.Frequency, req.Timezone)
	job := DetailedJob{
		Enabled:         true,
		Title:           req.Title,
		URL:             req.URL,
		SaveResponses:   true,
		Schedule:        schedule,
		RequestMethod:   methodInt,
		RequestTimeout:  300,
		Auth:            JobAuth{},
		Notification: JobNotificationSettings{
			OnFailure:       true,
			OnFailureCount:  1,
			OnSslCertExpiry: true,
			OnSslCertExpirySec: 604800,
		},
		ExtendedData: JobExtendedData{
			Headers: req.Headers,
			Body:    req.Body,
		},
	}

	jobID, err := c.CreateJob(job)
	if err != nil {
		log.Printf("[CronJob] CreateJob error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	log.Printf("[CronJob] Created job %d: %s -> %s", jobID, req.Title, req.URL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobId":   jobID,
		"message": fmt.Sprintf("Job '%s' created successfully", req.Title),
	})
}

// HandleUpdateJob updates an existing cron job
func HandleUpdateJob(w http.ResponseWriter, r *http.Request) {
	c := GetClient()
	if !c.IsConfigured() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "CRONJOB_API_KEY not configured"})
		return
	}

	jobID, err := strconv.Atoi(r.URL.Query().Get("jobId"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "jobId query parameter required"})
		return
	}

	var patch map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	if err := c.UpdateJob(jobID, patch); err != nil {
		log.Printf("[CronJob] UpdateJob(%d) error: %v", jobID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	log.Printf("[CronJob] Updated job %d", jobID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Job updated"})
}

// HandleDeleteJob deletes a cron job
func HandleDeleteJob(w http.ResponseWriter, r *http.Request) {
	c := GetClient()
	if !c.IsConfigured() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "CRONJOB_API_KEY not configured"})
		return
	}

	jobID, err := strconv.Atoi(r.URL.Query().Get("jobId"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "jobId query parameter required"})
		return
	}

	if err := c.DeleteJob(jobID); err != nil {
		log.Printf("[CronJob] DeleteJob(%d) error: %v", jobID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	log.Printf("[CronJob] Deleted job %d", jobID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Job deleted"})
}

// HandleJobHistory returns execution history for a job
func HandleJobHistory(w http.ResponseWriter, r *http.Request) {
	c := GetClient()
	if !c.IsConfigured() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "CRONJOB_API_KEY not configured"})
		return
	}

	jobID, err := strconv.Atoi(r.URL.Query().Get("jobId"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "jobId query parameter required"})
		return
	}

	resp, err := c.GetJobHistory(jobID)
	if err != nil {
		log.Printf("[CronJob] JobHistory(%d) error: %v", jobID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleSetupDefaults creates the default Koola10 cron jobs
func HandleSetupDefaults(w http.ResponseWriter, r *http.Request) {
	c := GetClient()
	if !c.IsConfigured() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "CRONJOB_API_KEY not configured",
			"message": "Add CRONJOB_API_KEY to your Render environment variables. Get your key at https://cron-job.org/console/settings",
			"steps": []string{
				"1. Go to https://cron-job.org and create an account",
				"2. Go to Settings → API Keys and generate a key",
				"3. Add CRONJOB_API_KEY to your Render environment variables",
				"4. Re-run this endpoint to create default jobs",
			},
		})
		return
	}

	type jobResult struct {
		Name   string `json:"name"`
		JobID  int    `json:"jobId,omitempty"`
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	}

	var results []jobResult

	// 1. Revenue Sprint - every 6 hours
	sprintID, err := c.CreateSprintJob("every-6-hours", "America/New_York")
	if err != nil {
		results = append(results, jobResult{Name: "Revenue Sprint (every 6h)", Status: "failed", Error: err.Error()})
	} else {
		results = append(results, jobResult{Name: "Revenue Sprint (every 6h)", JobID: sprintID, Status: "created"})
		log.Printf("[CronJob] Created default sprint job: %d", sprintID)
	}

	// 2. Health Check - every 10 minutes
	healthID, err := c.CreateHealthCheckJob("America/New_York")
	if err != nil {
		results = append(results, jobResult{Name: "Health Check (every 10m)", Status: "failed", Error: err.Error()})
	} else {
		results = append(results, jobResult{Name: "Health Check (every 10m)", JobID: healthID, Status: "created"})
		log.Printf("[CronJob] Created default health check job: %d", healthID)
	}

	// 3. Vault Summary - daily at 6 AM
	baseURL := os.Getenv("RENDER_EXTERNAL_URL")
	if baseURL == "" {
		baseURL = "https://koola10-ai-agent.onrender.com"
	}
	vaultJob := DetailedJob{
		Enabled:         true,
		Title:           "Koola10 Daily Vault Report",
		URL:             baseURL + "/vault/summary",
		SaveResponses:   true,
		Schedule:        buildSchedule("daily", "America/New_York"),
		RequestMethod:   0, // GET
		RequestTimeout:  30,
		Notification: JobNotificationSettings{
			OnFailure:       true,
			OnFailureCount:  1,
			OnSslCertExpiry: true,
			OnSslCertExpirySec: 604800,
		},
	}
	vaultID, err := c.CreateJob(vaultJob)
	if err != nil {
		results = append(results, jobResult{Name: "Daily Vault Report (6 AM)", Status: "failed", Error: err.Error()})
	} else {
		results = append(results, jobResult{Name: "Daily Vault Report (6 AM)", JobID: vaultID, Status: "created"})
		log.Printf("[CronJob] Created default vault report job: %d", vaultID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Default cron jobs created",
		"jobs":    results,
	})
}

// HandleStatus returns integration status
func HandleStatus(w http.ResponseWriter, r *http.Request) {
	c := GetClient()
	configured := c.IsConfigured()

	status := map[string]interface{}{
		"configured": configured,
		"service":    "cron-job.org",
		"category":   "Background Jobs",
		"api_url":    baseURL,
		"docs":       "https://docs.cron-job.org/rest-api.html",
	}

	if configured {
		// Try listing jobs to verify the key works
		resp, err := c.ListJobs()
		if err != nil {
			status["connected"] = false
			status["error"] = err.Error()
		} else {
			status["connected"] = true
			status["job_count"] = len(resp.Jobs)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
