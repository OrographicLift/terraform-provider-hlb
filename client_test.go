package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientCreateHLB(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "POST" {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}

		// Check request path
		if r.URL.Path != "/hlb" {
			t.Errorf("Expected request to '/hlb', got '%s'", r.URL.Path)
		}

		// Check API key header
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("Expected API Key header 'test-api-key', got '%s'", r.Header.Get("x-api-key"))
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"test-hlb-id","targetGroupArn":"arn:test","fqdn":"test.example.com","route53ZoneId":"Z12345","status":"active"}`))
	}))
	defer server.Close()

	// Create client pointing to test server
	client := NewClient("test-api-key", server.URL)

	// Test CreateHLB
	hlb := &HLB{
		TargetGroupARN: "arn:test",
		FQDN:           "test.example.com",
		Route53ZoneID:  "Z12345",
	}

	createdHLB, err := client.CreateHLB(hlb)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if createdHLB.ID != "test-hlb-id" {
		t.Errorf("Expected HLB ID 'test-hlb-id', got '%s'", createdHLB.ID)
	}

	if createdHLB.Status != "active" {
		t.Errorf("Expected HLB status 'active', got '%s'", createdHLB.Status)
	}
}

// Add more tests for GetHLB, UpdateHLB, and DeleteHLB...