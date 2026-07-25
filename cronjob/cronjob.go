package cronjob

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const baseURL = "https://api.cron-job.org"

// Client wraps the cron-job.org REST API
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new cron-job.org API client
func NewClient() *Client {
	return &Client{
		apiKey: os.Getenv("CRONJOB_API_KEY"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsConfigured returns true if the API key is set
func (c *Client) IsConfigured() bool {
	return c.apiKey != ""
}

// --- Data Types ---

type JobSchedule struct {
	Timezone  string `json:"timezone"`
	ExpiresAt int64  `json:"expiresAt"`
	Hours     []int  `json:"hours"`
	MDays     []int  `json:"mdays"`
	Minutes   []int  `json:"minutes"`
	Months    []int  `json:"months"`
	WDays     []int  `json:"wdays"`
}

type JobAuth struct {
	Enable   bool   `json:"enable"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type JobNotificationSettings struct {
	OnFailure          bool `json:"onFailure"`
	OnFailureCount     int  `json:"onFailureCount"`
	OnSuccess          bool `json:"onSuccess"`
	OnDisable          bool `json:"onDisable"`
	OnSslCertExpiry    bool `json:"onSslCertExpiry"`
	OnSslCertExpirySec int  `json:"onSslCertExpirySeconds"`
}

type JobExtendedData struct {
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type Job struct {
	JobID           int    `json:"jobId"`
	Enabled         bool   `json:"enabled"`
	Title           string `json:"title"`
	SaveResponses   bool   `json:"saveResponses"`
	URL             string `json:"url"`
	LastStatus      int    `json:"lastStatus"`
	LastDuration    int    `json:"lastDuration"`
	LastExecution   int64  `json:"lastExecution"`
	SSLCertExpiry   int64  `json:"sslCertExpiry"`
	NextExecution   int64  `json:"nextExecution"`
	Type            int    `json:"type"`
	RequestTimeout  int    `json:"requestTimeout"`
	RedirectSuccess bool   `json:"redirectSuccess"`
	FolderID        int    `json:"folderId"`
	Schedule        JobSchedule `json:"schedule"`
	RequestMethod   int    `json:"requestMethod"`
}

type DetailedJob struct {
	Job
	Auth         JobAuth                 `json:"auth"`
	Notification JobNotificationSettings `json:"notification"`
	ExtendedData JobExtendedData         `json:"extendedData"`
}

type JobsResponse struct {
	Jobs       []Job `json:"jobs"`
	SomeFailed bool  `json:"someFailed"`
}

type JobDetailsResponse struct {
	JobDetails DetailedJob `json:"jobDetails"`
}

type CreateJobResponse struct {
	JobID int `json:"jobId"`
}

type HistoryItem struct {
	JobLogID     int    `json:"jobLogId"`
	JobID        int    `json:"jobId"`
	Identifier   string `json:"identifier"`
	Date         int64  `json:"date"`
	DatePlanned  int64  `json:"datePlanned"`
	Jitter       int    `json:"jitter"`
	URL          string `json:"url"`
	Duration     int    `json:"duration"`
	Status       int    `json:"status"`
	StatusText   string `json:"statusText"`
	HTTPStatus   int    `json:"httpStatus"`
	Headers      string `json:"headers"`
	Body         string `json:"body"`
	SSLCertExpiry int64 `json:"sslCertExpiry"`
	Stats        struct {
		NameLookup   int `json:"nameLookup"`
		Connect      int `json:"connect"`
		AppConnect   int `json:"appConnect"`
		PreTransfer  int `json:"preTransfer"`
		StartTransfer int `json:"startTransfer"`
		Total        int `json:"total"`
	} `json:"stats"`
}

type HistoryResponse struct {
	History     []HistoryItem `json:"history"`
	Predictions []int64       `json:"predictions"`
}

// --- API Methods ---

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ListJobs returns all cron jobs
func (c *Client) ListJobs() (*JobsResponse, error) {
	body, err := c.doRequest("GET", "/jobs", nil)
	if err != nil {
		return nil, err
	}
	var resp JobsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &resp, nil
}

// GetJob returns details for a specific job
func (c *Client) GetJob(jobID int) (*DetailedJob, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/jobs/%d", jobID), nil)
	if err != nil {
		return nil, err
	}
	var resp JobDetailsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &resp.JobDetails, nil
}

// CreateJob creates a new cron job
func (c *Client) CreateJob(job DetailedJob) (int, error) {
	body, err := c.doRequest("PUT", "/jobs", map[string]interface{}{"job": job})
	if err != nil {
		return 0, err
	}
	var resp CreateJobResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, fmt.Errorf("unmarshal: %w", err)
	}
	return resp.JobID, nil
}

// UpdateJob updates an existing cron job
func (c *Client) UpdateJob(jobID int, patch map[string]interface{}) error {
	_, err := c.doRequest("PATCH", fmt.Sprintf("/jobs/%d", jobID), map[string]interface{}{"job": patch})
	return err
}

// DeleteJob deletes a cron job
func (c *Client) DeleteJob(jobID int) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/jobs/%d", jobID), nil)
	return err
}

// GetJobHistory returns execution history for a job
func (c *Client) GetJobHistory(jobID int) (*HistoryResponse, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/jobs/%d/history", jobID), nil)
	if err != nil {
		return nil, err
	}
	var resp HistoryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &resp, nil
}

// --- Convenience Methods ---

// CreateSprintJob creates a cron job that triggers a full revenue sprint
func (c *Client) CreateSprintJob(frequency string, timezone string) (int, error) {
	baseURL := os.Getenv("RENDER_EXTERNAL_URL")
	if baseURL == "" {
		baseURL = "https://koola10-ai-agent.onrender.com"
	}
	adminKey := os.Getenv("ADMIN_API_KEY")

	schedule := buildSchedule(frequency, timezone)

	job := DetailedJob{
		Enabled:  true,
		Title:    "Koola10 Revenue Sprint",
		URL:      baseURL + "/admin/run-scheduled-sprint",
		SaveResponses: true,
		Schedule: schedule,
		RequestMethod: 1, // POST
		RequestTimeout: 300,
		Notification: JobNotificationSettings{
			OnFailure:       true,
			OnFailureCount:  1,
			OnSuccess:       true,
			OnSslCertExpiry: true,
			OnSslCertExpirySec: 604800,
		},
		ExtendedData: JobExtendedData{
			Headers: map[string]string{
				"Authorization": "Bearer " + adminKey,
				"Content-Type":  "application/json",
			},
		},
	}

	return c.CreateJob(job)
}

// CreateHealthCheckJob creates a cron job that pings /health
func (c *Client) CreateHealthCheckJob(timezone string) (int, error) {
	baseURL := os.Getenv("RENDER_EXTERNAL_URL")
	if baseURL == "" {
		baseURL = "https://koola10-ai-agent.onrender.com"
	}

	job := DetailedJob{
		Enabled:  true,
		Title:    "Koola10 Health Check",
		URL:      baseURL + "/health",
		SaveResponses: true,
		Schedule: buildSchedule("every-10-min", timezone),
		RequestMethod: 0, // GET
		RequestTimeout: 30,
		Notification: JobNotificationSettings{
			OnFailure:      true,
			OnFailureCount: 3,
		},
	}

	return c.CreateJob(job)
}

// buildSchedule creates a JobSchedule from a frequency string
func buildSchedule(frequency, timezone string) JobSchedule {
	now := time.Now()
	schedule := JobSchedule{
		Timezone:  timezone,
		ExpiresAt: 0,
	}

	switch frequency {
	case "every-minute":
		schedule.Minutes = []int{-1}
		schedule.Hours = []int{-1}
		schedule.MDays = []int{-1}
		schedule.Months = []int{-1}
		schedule.WDays = []int{-1}
	case "every-5-min":
		minutes := []int{}
		for m := 0; m < 60; m += 5 {
			minutes = append(minutes, m)
		}
		schedule.Minutes = minutes
		schedule.Hours = []int{-1}
		schedule.MDays = []int{-1}
		schedule.Months = []int{-1}
		schedule.WDays = []int{-1}
	case "every-10-min":
		minutes := []int{}
		for m := 0; m < 60; m += 10 {
			minutes = append(minutes, m)
		}
		schedule.Minutes = minutes
		schedule.Hours = []int{-1}
		schedule.MDays = []int{-1}
		schedule.Months = []int{-1}
		schedule.WDays = []int{-1}
	case "hourly":
		schedule.Minutes = []int{now.Minute()}
		schedule.Hours = []int{-1}
		schedule.MDays = []int{-1}
		schedule.Months = []int{-1}
		schedule.WDays = []int{-1}
	case "every-6-hours":
		schedule.Minutes = []int{0}
		schedule.Hours = []int{0, 6, 12, 18}
		schedule.MDays = []int{-1}
		schedule.Months = []int{-1}
		schedule.WDays = []int{-1}
	case "daily":
		schedule.Minutes = []int{0}
		schedule.Hours = []int{6} // 6 AM
		schedule.MDays = []int{-1}
		schedule.Months = []int{-1}
		schedule.WDays = []int{-1}
	case "daily-morning":
		schedule.Minutes = []int{0}
		schedule.Hours = []int{8}
		schedule.MDays = []int{-1}
		schedule.Months = []int{-1}
		schedule.WDays = []int{-1}
	case "daily-evening":
		schedule.Minutes = []int{0}
		schedule.Hours = []int{20}
		schedule.MDays = []int{-1}
		schedule.Months = []int{-1}
		schedule.WDays = []int{-1}
	case "weekly":
		schedule.Minutes = []int{0}
		schedule.Hours = []int{6}
		schedule.MDays = []int{-1}
		schedule.Months = []int{-1}
		schedule.WDays = []int{1} // Monday
	default:
		// Default to hourly
		schedule.Minutes = []int{0}
		schedule.Hours = []int{-1}
		schedule.MDays = []int{-1}
		schedule.Months = []int{-1}
		schedule.WDays = []int{-1}
	}

	return schedule
}
