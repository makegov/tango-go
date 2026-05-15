// webhook-server is a minimal HTTP server that verifies Tango webhook
// signatures and prints the delivered payload.
//
// Run with:
//
//	TANGO_WEBHOOK_SECRET=whsec_… go run ./examples/webhook-server
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/makegov/tango-go/webhooks"
)

func main() {
	secret := os.Getenv("TANGO_WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("set TANGO_WEBHOOK_SECRET to run this example")
	}

	mux := http.NewServeMux()
	mux.Handle("/tango-webhook", webhooks.Middleware(secret,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var delivery map[string]any
			if err := json.NewDecoder(r.Body).Decode(&delivery); err != nil {
				http.Error(w, "bad json", http.StatusBadRequest)
				return
			}
			fmt.Printf("event=%v id=%v\n", delivery["event_type"], delivery["id"])
			w.WriteHeader(http.StatusNoContent)
		}),
	))

	addr := ":8080"
	log.Printf("listening on %s (POST /tango-webhook)", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
