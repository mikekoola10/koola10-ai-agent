package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"sentinel/financial"
	"sentinel/agents"
	"os"
)

func TestCashAppLimit(t *testing.T) {
	// Setup
	globalLedger = &EconomicLedger{}
	fundManager = financial.NewFundManager("test_fund.json", globalLedger)
	defer os.Remove("test_fund.json")

	globalSwarmManager = agents.NewSwarmManager()

	mux := http.NewServeMux()
	mux.HandleFunc("/cashapp/revenue", handleCashAppRevenue)

	// Test case 1: Under limit
	reqBody, _ := json.Marshal(map[string]float64{"amount": 30.0})
	req, _ := http.NewRequest("POST", "/cashapp/revenue", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	// Test case 2: Over limit
	reqBody, _ = json.Marshal(map[string]float64{"amount": 30.0})
	req, _ = http.NewRequest("POST", "/cashapp/revenue", bytes.NewBuffer(reqBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}
