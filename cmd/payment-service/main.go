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

	s := &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Payment service is listening on port :8081")
	log.Fatal(s.ListenAndServe())
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if isDown.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"error","service":"payment-service","message":"simulated outage"}`)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"payment-service"}`)
}

func simulateOutageHandler(w http.ResponseWriter, r *http.Request) {
	downParam := r.URL.Query().Get("down")
	if downParam == "false" || downParam == "0" {
		isDown.Store(false)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","message":"Payment service recovered"}`)
	} else {
		isDown.Store(true)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","message":"Payment service simulated outage activated"}`)
	}
}
