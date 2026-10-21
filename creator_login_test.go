package main

import (
	"context"
	"errors"
	"testing"
)

type fakeLoginAPI struct {
	verifyErr   error
	verifyCalls int
}

func (f *fakeLoginAPI) VerifyCaptcha(context.Context, string) error {
	f.verifyCalls++
	return f.verifyErr
}

func TestCreatorLoginHandoff(t *testing.T) {
	tests := []struct {
		name              string
		input             LoginInput
		verifyErr         error
		wantResume        string
		wantAuthorization bool
	}{
		{
			name:       "verified asset delivery continues to Google",
			input:      LoginInput{Provider: "google", CaptchaToken: "passed", CreatorID: "cr_17", AssetID: "asset_9", Next: "asset_delivery"},
			wantResume: "/creators/cr_17/assets/asset_9/delivery", wantAuthorization: true,
		},
		{
			name:       "verified subscriber update continues to GitHub",
			input:      LoginInput{Provider: "github", CaptchaToken: "passed", CreatorID: "cr_17", Next: "subscriber_updates"},
			wantResume: "/creators/cr_17/subscribers/updates", wantAuthorization: true,
		},
		{
			name:      "captcha rejection stops before OAuth",
			input:     LoginInput{Provider: "google", CaptchaToken: "rejected", CreatorID: "cr_17", Next: "content_processing"},
			verifyErr: &APIError{Code: "captcha_rejected", HTTPStatus: 422},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &fakeLoginAPI{verifyErr: tt.verifyErr}
			service := NewCreatorLogin(api, "http://localhost:8080", "google-client", "github-client")
			output, err := service.Start(context.Background(), tt.input)
			if !errors.Is(err, tt.verifyErr) {
				t.Fatalf("Start() error = %v, want %v", err, tt.verifyErr)
			}
			if tt.verifyErr != nil {
				if output.AuthorizationURL != "" {
					t.Fatalf("authorization URL = %q after captcha rejection", output.AuthorizationURL)
				}
				return
			}
			if output.ResumePath != tt.wantResume {
				t.Fatalf("handoff = %#v", output)
			}
			if tt.wantAuthorization && output.AuthorizationURL == "" {
				t.Fatal("authorization URL is empty")
			}
		})
	}
}
