package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTriggerPhoneAlert(t *testing.T) {
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	entered := []Witness{{Address: "TEntered1", DisplayName: "New SR", VoteCount: 1000}}
	left := []string{"TLeft1"}
	nameCache := map[string]string{"TLeft1": "Old SR"}

	if err := triggerPhoneAlert(server.URL, entered, left, nameCache); err != nil {
		t.Fatalf("triggerPhoneAlert returned error: %v", err)
	}

	message, _ := gotBody["message"].(string)
	if message == "" {
		t.Fatalf("expected non-empty message field")
	}
	if !strings.Contains(message, "New SR") {
		t.Errorf("expected message to mention entered witness 'New SR', got %q", message)
	}
	if !strings.Contains(message, "Old SR") {
		t.Errorf("expected message to mention left witness 'Old SR', got %q", message)
	}
}
