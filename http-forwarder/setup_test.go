// Copyright (c) J.Dreyer
// SPDX-License-Identifier: Apache-2.0

package http_forwarder_test

import (
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	router := mux.NewRouter()
	router.HandleFunc("/channels/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	server := httptest.NewUnstartedServer(router)
	l, err := net.Listen("tcp", ":9000")
	if err != nil {
		testLog.Error(fmt.Sprintf("Could not start listening on port : %s", err))
	}

	server.Listener = l
	server.Start()
	defer server.Close()

	code := m.Run()

	os.Exit(code)
}
