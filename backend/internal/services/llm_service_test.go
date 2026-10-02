package services

import (
	"testing"

	"accuscript-backend/internal/config"
	"accuscript-backend/internal/models"
)

func TestLLMService_PromptBuilding(t *testing.T) {
	cfg := &config.Config{DefaultLLMProv: "groq"}
	svc := NewLLMService(cfg)

	project := &models.Project{
		Protocol: models.Protocol{
			Criteria:          []string{"Double-Blind RCT", "Adult T2D"},
			InclusionCriteria: "Adults with T2D",
			ExclusionCriteria: "Animal models",
		},
	}

	study := &models.Study{
		Title:       "Test RCT Study",
		Author:      "Smith et al.",
		Year:        2023,
		ArticleType: "Journal",
		Abstract:    "A randomized trial evaluating remote care.",
	}

	prompt := svc.BuildPrompt(project, study)
	if prompt == "" {
		t.Fatalf("Generated prompt is empty")
	}

	if !contains(prompt, "Test RCT Study") || !contains(prompt, "Double-Blind RCT") {
		t.Fatalf("Prompt missing key study or protocol fields")
	}
}

func TestLLMService_ParseAndSanitizeJSON(t *testing.T) {
	cfg := &config.Config{DefaultLLMProv: "groq"}
	svc := NewLLMService(cfg)

	project := &models.Project{
		Protocol: models.Protocol{
			Criteria: []string{"Double-Blind RCT", "Adult T2D"},
		},
	}

	// 1. Raw JSON
	rawJSON := `{
		"suggestion": "include",
		"confidence": 0.92,
		"reasoning": "RCT design matches primary study requirement.",
		"matches": ["Double-Blind RCT"]
	}`

	res, err := svc.parseAndSanitizeJSON(rawJSON, project)
	if err != nil {
		t.Fatalf("Failed to parse clean JSON: %v", err)
	}
	if res.Suggestion != "include" || res.Confidence != 0.92 {
		t.Fatalf("Unexpected parsed result: %+v", res)
	}

	// 2. Markdown fenced JSON
	fencedJSON := "Here is the evaluation:\n```json\n{\n  \"suggestion\": \"exclude\",\n  \"confidence\": 0.88,\n  \"reasoning\": \"Fails RCT criterion.\",\n  \"matches\": []\n}\n```\nHope this helps!"

	res2, err := svc.parseAndSanitizeJSON(fencedJSON, project)
	if err != nil {
		t.Fatalf("Failed to parse markdown-fenced JSON: %v", err)
	}
	if res2.Suggestion != "exclude" || res2.Confidence != 0.88 {
		t.Fatalf("Unexpected parsed result from fenced JSON: %+v", res2)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
