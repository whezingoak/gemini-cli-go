package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/whezingoak/gemini-cli-go/pkg/auth"
	"github.com/whezingoak/gemini-cli-go/pkg/config"
)

const (
	BaseURL = "https://generativelanguage.googleapis.com/v1beta"
)

type Client struct {
	httpClient *http.Client
	Project    string
	Location   string
	Model      string
}

func NewClient(ctx context.Context) (*Client, error) {
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	cfg := config.Get()
	return &Client{
		httpClient: httpClient,
		Project:    cfg.Project,
		Location:   cfg.Location,
		Model:      cfg.Model,
	}, nil
}

// Simple text part
type Part struct {
	Text string `json:"text,omitempty"`
}

type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

type GenerateContentRequest struct {
	Contents         []Content        `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig,omitempty"`
}

type GenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type GenerateContentResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type ChatSession struct {
	client  *Client
	History []Content
	mu      sync.Mutex
}

func (c *Client) StartChat() *ChatSession {
	return &ChatSession{
		client:  c,
		History: []Content{},
	}
}

func (s *ChatSession) SendMessage(ctx context.Context, message string) (*GenerateContentResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Append user message to history
	userContent := Content{
		Role: "user",
		Parts: []Part{
			{Text: message},
		},
	}
	s.History = append(s.History, userContent)

	url := fmt.Sprintf("%s/models/%s:generateContent", BaseURL, s.client.Model)

	// We need to copy history to avoid race conditions if the slice is modified concurrently?
	// But we are holding the lock.
	// However, json.Marshal will read it.
	// If `SendMessage` is called concurrently, lock prevents concurrent modification of `s.History`.

	reqBody := GenerateContentRequest{
		Contents: s.History,
		GenerationConfig: GenerationConfig{
			Temperature:     config.Get().Temperature,
			MaxOutputTokens: config.Get().MaxOutputTokens,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var apiResp GenerateContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Append model response to history
	if len(apiResp.Candidates) > 0 {
		s.History = append(s.History, apiResp.Candidates[0].Content)
	}

	return &apiResp, nil
}

// Keep GenerateContent for single-turn requests if needed
func (c *Client) GenerateContent(ctx context.Context, prompt string) (*GenerateContentResponse, error) {
	session := c.StartChat()
	return session.SendMessage(ctx, prompt)
}
