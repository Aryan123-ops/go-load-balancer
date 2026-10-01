package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Server  string `json:"server"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

func main() {
	port := os.Getenv("PORT")
	host := os.Getenv("HOST")
	serverName := os.Getenv("SERVER_NAME")

	if port == "" {
		port = "9001"
	}
	if host == "" {
		host = "127.0.0.1"
	}
	if serverName == "" {
		serverName = "server-1"
	}

	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := map[string]string{
			"status": "UP",
			"server": serverName,
		}

		json.NewEncoder(w).Encode(response)
	})

	// Actual API
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := Response{
			Server:  serverName,
			Message: "Request handled successfully",
			Time:    time.Now().Format(time.RFC3339),
		}

		json.NewEncoder(w).Encode(response)
	})

	server := &http.Server{
		Addr:    host + ":" + port,
		Handler: mux,

		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("%s started on port %s", serverName, port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
