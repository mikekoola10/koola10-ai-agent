package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Aether Phase 1: Autonomous AI Influencer Core"))
	})

	log.Printf("Aether starting on port %s", port)
	http.ListenAndServe(":"+port, nil)
}
