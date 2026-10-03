package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"accuscript-backend/internal/config"
	"accuscript-backend/internal/models"
)

type LLMService struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewLLMService(cfg *config.Config) *LLMService {
	return &LLMService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

// BuildPrompt constructs the structured SLR prompt
func (s *LLMService) BuildPrompt(project *models.Project, study *models.Study) string {
	criteriaStr := strings.Join(project.Protocol.Criteria, ", ")
	if criteriaStr == "" {
		criteriaStr = "None specified"
	}

	inclusionStr := project.Protocol.InclusionCriteria
	if inclusionStr == "" {
		inclusionStr = "Primary empirical studies matching project objectives."
	}

	exclusionStr := project.Protocol.ExclusionCriteria
	if exclusionStr == "" {
		exclusionStr = "Non-primary studies, editorials, reviews, animal/in-vitro models."
	}

	return fmt.Sprintf(`You are an expert systematic literature review assistant.
PROJECT PROTOCOL:
Inclusion Criteria: %s
Exclusion Criteria: %s
Key Criteria: %s

STUDY:
Title: %s
Authors: %s (%d)
Article Type: %s
Abstract: %s

TASK: Determine if this study should be INCLUDED or EXCLUDED based on the protocol.
Respond ONLY with valid JSON matching this schema:
{
  "suggestion": "include"|"exclude",
  "confidence": 0.0-1.0,
  "reasoning": "Detailed explanation (2-3 sentences)",
  "matches": ["criterion1", "criterion2"]
}`,
		inclusionStr,
		exclusionStr,
		criteriaStr,
		study.Title,
		study.Author,
		study.Year,
		study.ArticleType,
		study.Abstract,
	)
}

// GenerateAISuggestion executes LLM call to configured provider and parses structured JSON
func (s *LLMService) GenerateAISuggestion(ctx context.Context, project *models.Project, study *models.Study, overrideProvider, overrideApiKey, overrideModel string) (*models.AISuggestionResult, error) {
	provider := strings.ToLower(strings.TrimSpace(project.LLMProvider))
	if overrideProvider != "" {
		provider = strings.ToLower(strings.TrimSpace(overrideProvider))
	}
	if provider == "" {
		provider = s.cfg.DefaultLLMProv
	}
	if provider == "" {
		provider = "groq"
	}

	apiKey := overrideApiKey
	if apiKey == "" {
		apiKey = project.LLMApiKey
	}

	// Fallbacks to server-level env variables if project key is empty
	if apiKey == "" {
		switch provider {
		case "groq":
			apiKey = s.cfg.GroqApiKey
		case "openai":
			apiKey = s.cfg.OpenAIApiKey
		case "gemini":
			apiKey = s.cfg.GeminiApiKey
		case "claude":
			apiKey = s.cfg.ClaudeApiKey
		}
	}

	model := overrideModel
	if model == "" {
		model = project.LLMModel
	}

	prompt := s.BuildPrompt(project, study)

	var rawResponse string
	var err error

	switch provider {
	case "groq":
		if model == "" {
			model = "llama-3.3-70b-versatile"
		}
		rawResponse, err = s.callGroq(ctx, apiKey, model, prompt)
	case "openai":
		if model == "" {
			model = "gpt-4o-mini"
		}
		rawResponse, err = s.callOpenAI(ctx, apiKey, model, prompt)
	case "gemini":
		if model == "" {
			model = "gemini-1.5-flash"
		}
		rawResponse, err = s.callGemini(ctx, apiKey, model, prompt)
	case "claude":
		if model == "" {
			model = "claude-3-5-haiku-20241022"
		}
		rawResponse, err = s.callClaude(ctx, apiKey, model, prompt)
	default:
		// Default to Groq if unrecognized
		if model == "" {
			model = "llama-3.3-70b-versatile"
		}
		rawResponse, err = s.callGroq(ctx, apiKey, model, prompt)
	}

	if err != nil {
		return nil, fmt.Errorf("LLM API error (%s): %w", provider, err)
	}

	return s.parseAndSanitizeJSON(rawResponse, project)
}

// callGroq calls Groq Cloud API
func (s *LLMService) callGroq(ctx context.Context, apiKey, model, userPrompt string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("Groq API key is required. Provide it in the project upload or set GROQ_API_KEY in .env")
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a specialized systematic literature review screener. Always respond with strict JSON."},
			{"role": "user", "content": userPrompt},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.1,
	}

	jsonBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(jsonBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq API status %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("failed to parse groq response structure: %w", err)
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("groq returned empty choices")
	}

	return parsed.Choices[0].Message.Content, nil
}

// callOpenAI calls OpenAI API
func (s *LLMService) callOpenAI(ctx context.Context, apiKey, model, userPrompt string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key is required. Provide it in the project upload or set OPENAI_API_KEY in .env")
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a specialized systematic literature review screener. Always respond with strict JSON."},
			{"role": "user", "content": userPrompt},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.1,
	}

	jsonBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openAI API status %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("openAI returned empty choices")
	}

	return parsed.Choices[0].Message.Content, nil
}

// callGemini calls Google Gemini API using official REST generateContent endpoint
func (s *LLMService) callGemini(ctx context.Context, apiKey, model, userPrompt string) (string, error) {
	cleanKey := strings.TrimSpace(apiKey)
	if cleanKey == "" {
		return "", fmt.Errorf("Gemini API key is required. Provide it in project upload or set GEMINI_API_KEY in .env")
	}

	cleanModel := strings.TrimPrefix(strings.TrimSpace(model), "models/")
	if cleanModel == "" || !strings.HasPrefix(cleanModel, "gemini") {
		cleanModel = "gemini-3.8-flash"
	}

	type candidate struct {
		apiVersion string
		modelName  string
	}

	candidates := []candidate{
		{"v1beta", cleanModel},
		{"v1beta", "gemini-3.8-flash"},
		{"v1beta", "gemini-1.5-flash-latest"},
		{"v1beta", "gemini-1.5-pro-latest"},
		{"v1beta", "gemini-2.0-flash-lite"},
		{"v1beta", "gemini-1.5-flash-002"},
		{"v1", "gemini-1.5-flash-latest"},
	}

	// Deduplicate candidates
	seen := make(map[string]bool)
	var uniqueCandidates []candidate
	for _, c := range candidates {
		key := c.apiVersion + ":" + c.modelName
		if !seen[key] {
			seen[key] = true
			uniqueCandidates = append(uniqueCandidates, c)
		}
	}

	var errorLogs []string

	for _, cand := range uniqueCandidates {
		endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/%s/models/%s:generateContent?key=%s", cand.apiVersion, cand.modelName, cleanKey)

		reqBody := map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"parts": []map[string]string{
						{"text": userPrompt},
					},
				},
			},
			"generationConfig": map[string]interface{}{
				"responseMimeType": "application/json",
				"temperature":      0.1,
			},
		}

		jsonBytes, err := json.Marshal(reqBody)
		if err != nil {
			return "", err
		}

		// Attempt call up to 3 times for 503 temporary high-demand spikes
		var resp *http.Response
		var body []byte
		var lastStatus int

		for attempt := 1; attempt <= 3; attempt++ {
			req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBytes))
			if err != nil {
				return "", err
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("x-goog-api-key", cleanKey)

			resp, err = s.httpClient.Do(req)
			if err != nil {
				log.Printf("[Gemini %s/%s attempt %d] HTTP error: %v", cand.apiVersion, cand.modelName, attempt, err)
				time.Sleep(1 * time.Second)
				continue
			}

			body, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			lastStatus = resp.StatusCode

			if resp.StatusCode == http.StatusOK {
				break
			}

			if resp.StatusCode == 503 || resp.StatusCode == 429 {
				log.Printf("[Gemini %s/%s attempt %d] High demand (status %d). Retrying in 1.5s...", cand.apiVersion, cand.modelName, attempt, resp.StatusCode)
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			// If status is 404 or other non-retriable, break inner loop to try next candidate
			break
		}

		if lastStatus != http.StatusOK {
			errStr := fmt.Sprintf("[%s/%s status %d: %s]", cand.apiVersion, cand.modelName, lastStatus, strings.TrimSpace(string(body)))
			log.Println(errStr)
			errorLogs = append(errorLogs, errStr)
			continue
		}

		var parsed struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}

		if err := json.Unmarshal(body, &parsed); err != nil {
			errStr := fmt.Sprintf("[%s/%s json parse error: %v]", cand.apiVersion, cand.modelName, err)
			log.Println(errStr)
			errorLogs = append(errorLogs, errStr)
			continue
		}

		if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
			errStr := fmt.Sprintf("[%s/%s empty candidates]", cand.apiVersion, cand.modelName)
			log.Println(errStr)
			errorLogs = append(errorLogs, errStr)
			continue
		}

		log.Printf("[Gemini Success] Used %s/%s", cand.apiVersion, cand.modelName)
		return parsed.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("all Gemini model candidates failed: %s", strings.Join(errorLogs, " | "))
}

// callClaude calls Anthropic Claude API
func (s *LLMService) callClaude(ctx context.Context, apiKey, model, userPrompt string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("Claude API key is required. Provide it in project upload or set CLAUDE_API_KEY in .env")
	}

	reqBody := map[string]interface{}{
		"model":      model,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt + "\n\nRemember: Respond ONLY with valid raw JSON object, no Markdown code blocks or explanation outside the JSON."},
		},
		"temperature": 0.1,
	}

	jsonBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("claude API status %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("failed to parse Claude response: %w", err)
	}

	if len(parsed.Content) == 0 {
		return "", fmt.Errorf("claude returned empty content")
	}

	return parsed.Content[0].Text, nil
}

// parseAndSanitizeJSON extracts JSON object, handles markdown wrappers, and normalizes schema
func (s *LLMService) parseAndSanitizeJSON(rawText string, project *models.Project) (*models.AISuggestionResult, error) {
	cleaned := strings.TrimSpace(rawText)

	// Remove markdown code fences like ```json ... ``` if present
	reCodeBlock := regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)\\s*```")
	if matches := reCodeBlock.FindStringSubmatch(cleaned); len(matches) > 1 {
		cleaned = strings.TrimSpace(matches[1])
	}

	// Extract everything from first '{' to last '}'
	firstBrace := strings.Index(cleaned, "{")
	lastBrace := strings.LastIndex(cleaned, "}")
	if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
		cleaned = cleaned[firstBrace : lastBrace+1]
	}

	var result models.AISuggestionResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse structured LLM JSON (%w). Raw response: %s", err, rawText)
	}

	// Normalize Suggestion
	sugg := strings.ToLower(strings.TrimSpace(result.Suggestion))
	if strings.Contains(sugg, "include") {
		result.Suggestion = "include"
	} else if strings.Contains(sugg, "exclude") {
		result.Suggestion = "exclude"
	} else {
		result.Suggestion = "exclude"
	}

	// Validate / Clamp Confidence
	if result.Confidence < 0.0 || result.Confidence > 1.0 {
		if result.Confidence > 1.0 && result.Confidence <= 100.0 {
			result.Confidence = result.Confidence / 100.0
		} else {
			result.Confidence = 0.85
		}
	}
	if result.Confidence == 0 {
		result.Confidence = 0.85
	}

	// Ensure Reasoning is populated
	if strings.TrimSpace(result.Reasoning) == "" {
		if result.Suggestion == "include" {
			result.Reasoning = "Study matches the primary inclusion protocol criteria without triggering exclusion rules."
		} else {
			result.Reasoning = "Study fails one or more inclusion criteria or matches project exclusion criteria."
		}
	}

	// Ensure Matches slice is populated
	if len(result.Matches) == 0 && len(project.Protocol.Criteria) > 0 {
		if result.Suggestion == "include" {
			result.Matches = project.Protocol.Criteria
		}
	}

	return &result, nil
}
