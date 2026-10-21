package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
)

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	widgetRecordID := os.Getenv("INFRAI_WIDGET_RECORD_ID")
	if widgetRecordID == "" {
		log.Fatal("INFRAI_WIDGET_RECORD_ID is required")
	}
	publicURL := envOr("PUBLIC_URL", "http://localhost:8080")
	service := NewCreatorLogin(
		NewInfraiClient(apiKey, widgetRecordID, nil),
		publicURL,
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GITHUB_CLIENT_ID"),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /login/start", startLoginHandler(service))
	address := envOr("LISTEN_ADDR", ":8080")
	log.Printf("creator login gateway listening on %s", address)
	log.Fatal(http.ListenAndServe(address, mux))
}

func startLoginHandler(service *CreatorLogin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var input LoginInput
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request"})
			return
		}
		output, err := service.Start(r.Context(), input)
		if err != nil {
			writeJSON(w, clientStatus(err), map[string]string{"error": publicError(err)})
			return
		}
		writeJSON(w, http.StatusOK, output)
	}
}

func publicError(err error) string {
	var inputErr *InputError
	if errors.As(err, &inputErr) {
		return inputErr.Message
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	return "upstream request failed"
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
