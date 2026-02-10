package auth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2/google"
	"github.com/spf13/viper"
)

// AuthType defines the authentication method
type AuthType string

const (
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeADC    AuthType = "adc"
)

// GetClient returns an authenticated HTTP client
func GetClient(ctx context.Context) (*http.Client, error) {
	// First check for API Key in config/env
	apiKey := viper.GetString("api_key")
	if apiKey != "" {
		return &http.Client{
			Transport: &apiKeyTransport{
				apiKey: apiKey,
				base:   http.DefaultTransport,
			},
		}, nil
	}

	// Fallback to ADC
	_, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/generative-language")
	if err != nil {
		return nil, fmt.Errorf("failed to find default credentials: %w", err)
	}

	return google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/generative-language")
}

type apiKeyTransport struct {
	apiKey string
	base   http.RoundTripper
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	q := req.URL.Query()
	q.Add("key", t.apiKey)
	req.URL.RawQuery = q.Encode()
	return t.base.RoundTrip(req)
}

// Login initiates the login flow (placeholder for now, assumes ADC or API Key setup)
func Login() error {
	// In a real CLI, this might open a browser for OAuth flow if client ID is provided,
	// or guide the user to set up ADC/API Key.
	fmt.Println("Please set GEMINI_API_KEY environment variable or run 'gcloud auth application-default login'")
	return nil
}
