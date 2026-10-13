package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var isDown atomic.Bool

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /simulate-outage", simulateOutageHandler)
	mux.HandleFunc("GET /simulate-outage", simulateOutageHandler)

	server := &http.Server{
		Addr:         ":8082",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Order service is listening on port :8082")
	log.Fatal(server.ListenAndServe())
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if isDown.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"error","service":"order-service","message":"simulated outage"}`)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"order-service"}`)
}

func simulateOutageHandler(w http.ResponseWriter, r *http.Request) {
	downParam := r.URL.Query().Get("down")
	if downParam == "false" || downParam == "0" {
		isDown.Store(false)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","message":"Order service recovered"}`)
	} else {
		isDown.Store(true)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","message":"Order service simulated outage activated"}`)
	}
}
