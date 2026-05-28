// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	service_template_http "github.com/trustap/service_template/pkg/http"
)

type user struct {
	ID    string
	Email string
}

type userID struct {
	ID string `json:"id"`
}

var (
	users   = map[string]*user{}
	usersMu sync.Mutex
)

func main() {
	argv := os.Args
	if len(argv) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <config-yaml> <listen-addr>", argv[0])
		os.Exit(1)
	}
	configYamlPath := argv[1]
	listenAddr := argv[2]

	err := run(configYamlPath, listenAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't run server: %v", err)
		os.Exit(1)
	}
}

func run(configYamlPath, listenAddr string) error {
	fmt.Printf("configuration is at '%s'", configYamlPath)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/users", createUser)
	mux.HandleFunc("GET /heartbeat", heartbeat)

	server := &http.Server{Addr: listenAddr, Handler: mux}

	fmt.Printf("listening on %s\n", listenAddr)
	err := service_template_http.ListenAndServe(server, 3*time.Second)
	if err != nil {
		return fmt.Errorf("couldn't listen and serve: %w", err)
	}

	return nil
}

func createUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "couldn't parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")

	id, err := uuid.NewV4()
	if err != nil {
		http.Error(w, "couldn't generate ID", http.StatusInternalServerError)
		return
	}

	usersMu.Lock()
	users[id.String()] = &user{
		ID:    id.String(),
		Email: email,
	}
	usersMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&userID{ID: id.String()}); err != nil {
		fmt.Fprintf(os.Stderr, "couldn't encode response: %v\n", err)
	}
}

func heartbeat(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
