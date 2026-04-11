// testserver is a tiny HTTP server used to simulate uptime/downtime for Pingr.
// Start it: go run ./cmd/testserver   (listens on :9999)
// Stop it:  Ctrl+C  — worker detects it as DOWN within 10 seconds
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":"ok"}`)
	})
	log.Println("test server running on http://localhost:9998  — Ctrl+C to simulate downtime")
	if err := http.ListenAndServe(":9998", nil); err != nil {
		log.Fatal(err)
	}
}
