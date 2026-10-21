package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

type LoginInput struct {
	Provider     string `json:"provider"`
	CaptchaToken string `json:"captcha_token"`
	CreatorID    string `json:"creator_id"`
	AssetID      string `json:"asset_id"`
	Next         string `json:"next"`
}

type LoginOutput struct {
	AuthorizationURL string `json:"authorization_url"`
	ResumePath       string `json:"resume_path"`
}

type loginAPI interface {
	VerifyCaptcha(context.Context, string) error
}

type CreatorLogin struct {
	api         loginAPI
	publicURL   string
	redirectURI string
	clientIDs   map[string]string
}

func NewCreatorLogin(api loginAPI, publicURL, googleClientID, githubClientID string) *CreatorLogin {
	return &CreatorLogin{
		api:         api,
		publicURL:   publicURL,
		redirectURI: publicURL + "/oauth/callback",
		clientIDs: map[string]string{
			"google": googleClientID,
			"github": githubClientID,
		},
	}
}

func (s *CreatorLogin) Start(ctx context.Context, input LoginInput) (LoginOutput, error) {
	if input.Provider != "google" && input.Provider != "github" {
		return LoginOutput{}, &InputError{Message: "provider must be google or github"}
	}
	if input.CaptchaToken == "" || input.CreatorID == "" {
		return LoginOutput{}, &InputError{Message: "captcha_token and creator_id are required"}
	}
	resumePath, err := resumePath(input)
	if err != nil {
		return LoginOutput{}, err
	}
	if err := s.api.VerifyCaptcha(ctx, input.CaptchaToken); err != nil {
		return LoginOutput{}, err
	}
	returnTo := s.publicURL + resumePath
	authorizationURL, err := s.authorizationURL(input.Provider, returnTo)
	if err != nil {
		return LoginOutput{}, err
	}
	return LoginOutput{AuthorizationURL: authorizationURL, ResumePath: resumePath}, nil
}

func (s *CreatorLogin) authorizationURL(provider, returnTo string) (string, error) {
	clientID := s.clientIDs[provider]
	if clientID == "" {
		return "", &InputError{Message: provider + " client ID is required"}
	}
	values := url.Values{
		"client_id":    {clientID},
		"redirect_uri": {s.redirectURI},
		"state":        {returnTo},
	}
	if provider == "google" {
		values.Set("response_type", "code")
		values.Set("scope", "openid email profile")
		return "https://accounts.google.com/o/oauth2/v2/auth?" + values.Encode(), nil
	}
	values.Set("scope", "read:user user:email")
	return "https://github.com/login/oauth/authorize?" + values.Encode(), nil
}

func resumePath(input LoginInput) (string, error) {
	creator := url.PathEscape(input.CreatorID)
	switch input.Next {
	case "asset_delivery":
		if input.AssetID == "" {
			return "", &InputError{Message: "asset_id is required for asset_delivery"}
		}
		return fmt.Sprintf("/creators/%s/assets/%s/delivery", creator, url.PathEscape(input.AssetID)), nil
	case "subscriber_updates":
		return fmt.Sprintf("/creators/%s/subscribers/updates", creator), nil
	case "content_processing":
		return fmt.Sprintf("/creators/%s/content/process", creator), nil
	default:
		return "", &InputError{Message: "next must name a creator workflow"}
	}
}

type InputError struct {
	Message string
}

func (e *InputError) Error() string { return e.Message }

func clientStatus(err error) int {
	var inputErr *InputError
	if errors.As(err, &inputErr) {
		return 400
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
		return apiErr.HTTPStatus
	}
	return 502
}
