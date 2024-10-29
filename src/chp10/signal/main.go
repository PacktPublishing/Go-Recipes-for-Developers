package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Create a simple HTTP echo service
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, r.Body)
	})
	server := &http.Server{Addr: ":8080"}

	// Listen for SIGINT and SIGTERM signals
	// Terminate the server with the signal
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigTerm
		server.Shutdown(context.Background())
	}()

	server.ListenAndServe()
}