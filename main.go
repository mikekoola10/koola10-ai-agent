package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	log.Printf("starting server on 0.0.0.0:%s", port)

	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		log.Fatalf("server_failed: %v", err)
	}
}
