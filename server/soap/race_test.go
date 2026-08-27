package soap

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// TestConcurrentRegistrationAndServe guards the embedding use case where
// handlers are registered while the server is already serving traffic.
// Under -race the unsynchronized map access is a hard failure.
func TestConcurrentRegistrationAndServe(t *testing.T) {
	h := NewHandler("", "")

	h.RegisterContextHandler("Stable", func(_ *RequestContext, _ []byte) (interface{}, error) {
		return RawXML("<ok/>"), nil
	})

	request := func(action string) *http.Request {
		return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
			`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><`+
				action+`/></s:Body></s:Envelope>`))
	}

	// Registrations racing with traffic.
	var regWG sync.WaitGroup
	regWG.Add(1)

	go func() {
		defer regWG.Done()

		for i := range 64 {
			action := "Dynamic" + strings.Repeat("X", i+1)
			h.RegisterContextHandler(action, func(_ *RequestContext, _ []byte) (interface{}, error) {
				return RawXML("<ok/>"), nil
			})
		}
	}()

	// Concurrent traffic on the stable action.
	var serveWG sync.WaitGroup
	for range 8 {
		serveWG.Add(1)

		go func() {
			defer serveWG.Done()

			for range 16 {
				w := httptest.NewRecorder()
				h.ServeHTTP(w, request("Stable"))

				if w.Code != http.StatusOK {
					t.Errorf("stable action failed under concurrency: %d", w.Code)

					return
				}
			}
		}()
	}

	serveWG.Wait()
	regWG.Wait()

	// Every dynamic registration must be reachable afterwards.
	w := httptest.NewRecorder()
	h.ServeHTTP(w, request("DynamicXXXX"))

	if w.Code != http.StatusOK {
		t.Errorf("dynamic action after concurrent registration: %d", w.Code)
	}
}
