package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"accuscript-backend/internal/db"
	"accuscript-backend/internal/models"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type StudyHandler struct {
	db *db.MongoDB
}

func NewStudyHandler(database *db.MongoDB) *StudyHandler {
	return &StudyHandler{db: database}
}

// ListByProject handles GET /api/projects/{id}/studies
func (h *StudyHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "id")
	query := r.URL.Query()

	decisionFilter := strings.ToLower(strings.TrimSpace(query.Get("decision")))
	search := strings.TrimSpace(query.Get("search"))

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 500 {
		limit = 100
	}

	filter := bson.M{"projectId": projectId}

	if decisionFilter != "" && decisionFilter != "all" {
		filter["decision"] = decisionFilter
	}

	if search != "" {
		regexPattern := bson.M{"$regex": search, "$options": "i"}
		filter["$or"] = []bson.M{
			{"title": regexPattern},
			{"author": regexPattern},
			{"abstract": regexPattern},
			{"ID": regexPattern},
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	totalCount, err := h.db.Studies().CountDocuments(ctx, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to count studies: "+err.Error())
		return
	}

	findOpts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "_id", Value: 1}})

	cursor, err := h.db.Studies().Find(ctx, filter, findOpts)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch studies: "+err.Error())
		return
	}
	defer cursor.Close(ctx)

	var studies []models.Study
	if err := cursor.All(ctx, &studies); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decode studies: "+err.Error())
		return
	}

	if studies == nil {
		studies = []models.Study{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":       studies,
		"total":      totalCount,
		"page":       page,
		"limit":      limit,
		"totalPages": (int(totalCount) + limit - 1) / limit,
	})
}

// Get handles GET /api/studies/{studyId}
func (h *StudyHandler) Get(w http.ResponseWriter, r *http.Request) {
	studyId := chi.URLParam(r, "studyId")
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var study models.Study
	err := h.db.Studies().FindOne(ctx, bson.M{"studyId": studyId}).Decode(&study)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			respondError(w, http.StatusNotFound, "Study not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to fetch study: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, study)
}

// UpdateDecision handles PATCH /api/studies/{studyId}/decision
func (h *StudyHandler) UpdateDecision(w http.ResponseWriter, r *http.Request) {
	studyId := chi.URLParam(r, "studyId")

	var req models.DecisionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON request body: "+err.Error())
		return
	}

	decision := strings.ToLower(strings.TrimSpace(req.Decision))
	if decision != "included" && decision != "excluded" && decision != "undecided" {
		respondError(w, http.StatusBadRequest, "Decision must be 'included', 'excluded', or 'undecided'")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var decidedAt *time.Time
	if decision != "undecided" {
		now := time.Now().UTC()
		decidedAt = &now
	}

	update := bson.M{
		"$set": bson.M{
			"decision":  decision,
			"decidedAt": decidedAt,
		},
	}

	res, err := h.db.Studies().UpdateOne(ctx, bson.M{"studyId": studyId}, update)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update decision: "+err.Error())
		return
	}

	if res.MatchedCount == 0 {
		respondError(w, http.StatusNotFound, "Study not found")
		return
	}

	var updatedStudy models.Study
	_ = h.db.Studies().FindOne(ctx, bson.M{"studyId": studyId}).Decode(&updatedStudy)

	respondJSON(w, http.StatusOK, updatedStudy)
}
