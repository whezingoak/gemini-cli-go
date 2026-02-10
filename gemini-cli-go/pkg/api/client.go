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

func (s *ChatSession) AddContext(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.History = append(s.History, Content{
		Role: "user",
		Parts: []Part{
			{Text: text},
		},
	})
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

// SendMessageStream sends a message and streams the response via callback
func (s *ChatSession) SendMessageStream(ctx context.Context, message string, onChunk func(string)) (*GenerateContentResponse, error) {
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

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent", BaseURL, s.client.Model)

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

	dec := json.NewDecoder(resp.Body)

	// Expect start of array
	t, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON token: %w", err)
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		// It might be a single object error if something went wrong but status was 200 (unlikely)
		return nil, fmt.Errorf("expected JSON array start, got %v", t)
	}

	var fullText string
	for dec.More() {
		var chunk GenerateContentResponse
		if err := dec.Decode(&chunk); err != nil {
			return nil, fmt.Errorf("failed to decode chunk: %w", err)
		}
		if len(chunk.Candidates) > 0 && len(chunk.Candidates[0].Content.Parts) > 0 {
			text := chunk.Candidates[0].Content.Parts[0].Text
			fullText += text
			if onChunk != nil {
				onChunk(text)
			}
		}
	}

	// Consume end of array
	_, _ = dec.Token()

	// Append full model response to history
	modelContent := Content{
		Role: "model",
		Parts: []Part{{Text: fullText}},
	}
	s.History = append(s.History, modelContent)

	return &GenerateContentResponse{Candidates: []Candidate{{Content: modelContent}}}, nil
}

// Keep GenerateContent for single-turn requests if needed
func (c *Client) GenerateContent(ctx context.Context, prompt string) (*GenerateContentResponse, error) {
	session := c.StartChat()
	return session.SendMessage(ctx, prompt)
}
