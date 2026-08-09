package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	s := &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("HTTP server is running on port 8081")

	log.Fatal(s.ListenAndServe())

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Payment service is running")
}
