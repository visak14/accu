package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"accuscript-backend/internal/db"
	"accuscript-backend/internal/models"
	"accuscript-backend/internal/services"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AIHandler struct {
	db     *db.MongoDB
	llm    *services.LLMService
	crypto *services.CryptoService
}

func NewAIHandler(database *db.MongoDB, llm *services.LLMService, crypto *services.CryptoService) *AIHandler {
	return &AIHandler{
		db:     database,
		llm:    llm,
		crypto: crypto,
	}
}

// Suggest handles POST /api/studies/{studyId}/ai-suggest
func (h *AIHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	studyId := chi.URLParam(r, "studyId")

	var req models.AISuggestRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	ctx, cancel := context.WithTimeout(r.Context(), 50*time.Second)
	defer cancel()

	// 1. Fetch Study
	var study models.Study
	err := h.db.Studies().FindOne(ctx, bson.M{"studyId": studyId}).Decode(&study)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			respondError(w, http.StatusNotFound, "Study not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve study: "+err.Error())
		return
	}

	// 2. Fetch Project
	var project models.Project
	err = h.db.Projects().FindOne(ctx, bson.M{"projectId": study.ProjectId}).Decode(&project)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve project protocol: "+err.Error())
		return
	}

	// 3. Decrypt Project LLM API Key if available
	decryptedApiKey := ""
	if project.LLMApiKey != "" {
		dec, err := h.crypto.Decrypt(project.LLMApiKey)
		if err == nil {
			decryptedApiKey = dec
		}
	}

	// Allow request-time override if provided
	apiKey := req.ApiKey
	if apiKey == "" {
		apiKey = decryptedApiKey
	}

	provider := req.Provider
	if provider == "" {
		provider = project.LLMProvider
	}

	model := req.Model
	if model == "" {
		model = project.LLMModel
	}

	// 4. Generate AI Suggestion via LLM Service
	aiResult, err := h.llm.GenerateAISuggestion(ctx, &project, &study, provider, apiKey, model)
	if err != nil {
		respondError(w, http.StatusBadGateway, "AI screening error: "+err.Error())
		return
	}

	// 5. Update Study in MongoDB
	now := time.Now().UTC()
	aiReasonFormatted := fmt.Sprintf("LLM (%s): '%s'", strings.ToUpper(provider), aiResult.Reasoning)

	updateDoc := bson.M{
		"$set": bson.M{
			"aiSuggestion":   aiResult.Suggestion,
			"aiConfidence":   aiResult.Confidence,
			"aiReason":       aiReasonFormatted,
			"aiMatches":      aiResult.Matches,
			"aiJsonResponse": aiResult,
			"aiScreenedAt":   now,
		},
	}

	_, err = h.db.Studies().UpdateOne(ctx, bson.M{"studyId": studyId}, updateDoc)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save AI screening result: "+err.Error())
		return
	}

	// 6. Return response matching assignment contract
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"suggestion":     aiResult.Suggestion,
		"confidence":     aiResult.Confidence,
		"reasoning":      aiResult.Reasoning,
		"matches":        aiResult.Matches,
		"aiReason":       aiReasonFormatted,
		"aiJsonResponse": aiResult,
		"studyId":        studyId,
	})
}

// BatchSuggest handles POST /api/projects/{id}/batch-ai-suggest
func (h *AIHandler) BatchSuggest(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "id")

	var req struct {
		Limit          int    `json:"limit"`
		UndecidedOnly  bool   `json:"undecidedOnly"`
		UnscreenedOnly bool   `json:"unscreenedOnly"`
		Provider       string `json:"provider,omitempty"`
		ApiKey         string `json:"apiKey,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.Limit < 1 || req.Limit > 50 {
		req.Limit = 10
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	// 1. Fetch Project
	var project models.Project
	err := h.db.Projects().FindOne(ctx, bson.M{"projectId": projectId}).Decode(&project)
	if err != nil {
		respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	decryptedApiKey := ""
	if project.LLMApiKey != "" {
		if dec, err := h.crypto.Decrypt(project.LLMApiKey); err == nil {
			decryptedApiKey = dec
		}
	}
	if req.ApiKey != "" {
		decryptedApiKey = req.ApiKey
	}

	provider := req.Provider
	if provider == "" {
		provider = project.LLMProvider
	}

	// Build filter
	filter := bson.M{"projectId": projectId}
	if req.UndecidedOnly {
		filter["decision"] = "undecided"
	}
	if req.UnscreenedOnly {
		filter["aiSuggestion"] = nil
	}

	findOpts := options.Find().SetLimit(int64(req.Limit))
	cursor, err := h.db.Studies().Find(ctx, filter, findOpts)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to query studies for batch screening: "+err.Error())
		return
	}
	defer cursor.Close(ctx)

	var studies []models.Study
	if err := cursor.All(ctx, &studies); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decode studies: "+err.Error())
		return
	}

	type BatchResultItem struct {
		StudyId    string  `json:"studyId"`
		Title      string  `json:"title"`
		Suggestion string  `json:"suggestion"`
		Confidence float64 `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
		Error      string  `json:"error,omitempty"`
	}

	var results []BatchResultItem

	for i := range studies {
		st := &studies[i]
		aiRes, err := h.llm.GenerateAISuggestion(ctx, &project, st, provider, decryptedApiKey, project.LLMModel)
		if err != nil {
			results = append(results, BatchResultItem{
				StudyId: st.StudyId,
				Title:   st.Title,
				Error:   err.Error(),
			})
			continue
		}

		now := time.Now().UTC()
		aiReasonFormatted := fmt.Sprintf("LLM (%s): '%s'", strings.ToUpper(provider), aiRes.Reasoning)

		_, _ = h.db.Studies().UpdateOne(ctx, bson.M{"studyId": st.StudyId}, bson.M{
			"$set": bson.M{
				"aiSuggestion":   aiRes.Suggestion,
				"aiConfidence":   aiRes.Confidence,
				"aiReason":       aiReasonFormatted,
				"aiMatches":      aiRes.Matches,
				"aiJsonResponse": aiRes,
				"aiScreenedAt":   now,
			},
		})

		results = append(results, BatchResultItem{
			StudyId:    st.StudyId,
			Title:      st.Title,
			Suggestion: aiRes.Suggestion,
			Confidence: aiRes.Confidence,
			Reasoning:  aiRes.Reasoning,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"processedCount": len(results),
		"results":        results,
	})
}
