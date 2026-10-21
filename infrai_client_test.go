package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifyCaptchaIncludesWidgetRecordID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if got := body["widget_record_id"]; got != "widget_123" {
			t.Errorf("widget_record_id = %v, want widget_123", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data":{}}`))
	}))
	defer server.Close()

	client := NewInfraiClient("api-key", "widget_123", server.Client())
	client.baseURL = server.URL
	if err := client.VerifyCaptcha(context.Background(), "captcha-token"); err != nil {
		t.Fatalf("VerifyCaptcha() error = %v", err)
	}
}
