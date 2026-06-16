package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/health", nil)
    rr := httptest.NewRecorder()

    healthHandler := NewHealthHandler("v1.0.0")

    handler := http.HandlerFunc(healthHandler.HandleHealth)
    handler.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", rr.Code)
    }

    var resp HealthResponse
    if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
        t.Fatalf("failed to decode body: %v", err)
    }

    if resp.Status != "ok" {
        t.Errorf("expected status 'ok', got '%s'", resp.Status)
    }

    if resp.Version != "v1.0.0" {
        t.Errorf("expected version 'v1.0.0', got '%s'", resp.Version)
    }
}
