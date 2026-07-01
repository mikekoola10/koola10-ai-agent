package api

import (
	"encoding/json"
	"net/http"
	"time"
	"koola10/solara"
)

type BriefResponse struct {
	Date    string `json:"date"`
	Content string `json:"content"`
}

func GetLatestBriefHandler(db *solara.DailyBrief) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := db.GetLatestBrief()
		if err != nil {
			http.Error(w, "No brief available", http.StatusNotFound)
			return
		}
		resp := BriefResponse{
			Date:    time.Now().Format("2006-01-02"),
			Content: content,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
