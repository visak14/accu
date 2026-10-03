package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"accuscript-backend/internal/db"
	"accuscript-backend/internal/models"
	"accuscript-backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProjectHandler struct {
	db     *db.MongoDB
	crypto *services.CryptoService
	excel  *services.ExcelService
}

func NewProjectHandler(database *db.MongoDB, crypto *services.CryptoService, excel *services.ExcelService) *ProjectHandler {
	return &ProjectHandler{
		db:     database,
		crypto: crypto,
		excel:  excel,
	}
}

// Upload handles POST /api/projects/upload
func (h *ProjectHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Max upload size 32 MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		respondError(w, http.StatusBadRequest, "Project Name is required")
		return
	}

	description := strings.TrimSpace(r.FormValue("description"))
	llmProvider := strings.TrimSpace(r.FormValue("llmProvider"))
	if llmProvider == "" {
		llmProvider = "groq"
	}
	llmModel := strings.TrimSpace(r.FormValue("llmModel"))
	rawApiKey := strings.TrimSpace(r.FormValue("llmApiKey"))

	// Retrieve Studies Excel
	studiesFile, _, err := r.FormFile("studies")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Studies Excel file ('studies') is required")
		return
	}
	defer studiesFile.Close()

	// Retrieve Protocol Excel
	protocolFile, _, err := r.FormFile("protocol")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Protocol Excel file ('protocol') is required")
		return
	}
	defer protocolFile.Close()

	projectId := "proj-" + uuid.New().String()[:8]

	// 1. Parse Studies Excel
	studies, err := h.excel.ParseStudiesExcel(studiesFile, projectId)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse studies excel: "+err.Error())
		return
	}

	// 2. Parse Protocol Excel
	protocol, err := h.excel.ParseProtocolExcel(protocolFile)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse protocol excel: "+err.Error())
		return
	}

	// 3. Encrypt LLM API Key if provided
	encryptedApiKey := ""
	if rawApiKey != "" {
		enc, err := h.crypto.Encrypt(rawApiKey)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to securely encrypt API key: "+err.Error())
			return
		}
		encryptedApiKey = enc
	}

	now := time.Now().UTC()
	project := models.Project{
		ProjectId:    projectId,
		Name:         name,
		Description:  description,
		LLMProvider:  llmProvider,
		LLMModel:     llmModel,
		LLMApiKey:    encryptedApiKey,
		TotalStudies: len(studies),
		Protocol:     *protocol,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// Insert Project
	insertRes, err := h.db.Projects().InsertOne(ctx, project)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to store project in database: "+err.Error())
		return
	}
	if oid, ok := insertRes.InsertedID.(bson.ObjectID); ok {
		project.ID = oid
	}

	// Bulk Insert Studies
	var studyDocs []interface{}
	for i := range studies {
		studyDocs = append(studyDocs, studies[i])
	}

	if len(studyDocs) > 0 {
		_, err = h.db.Studies().InsertMany(ctx, studyDocs)
		if err != nil {
			log.Printf("[Upload Error] Failed to insert %d studies: %v", len(studyDocs), err)
			// Rollback project if studies insert fails
			_, _ = h.db.Projects().DeleteOne(ctx, bson.M{"projectId": projectId})
			respondError(w, http.StatusInternalServerError, "Failed to store studies in database: "+err.Error())
			return
		}
	}

	log.Printf("[Upload Success] Created project '%s' (ID: %s) with %d studies", name, projectId, len(studies))

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"projectId":    projectId,
		"totalStudies": len(studies),
		"message":      "Project and studies successfully imported with protocol!",
	})
}

// List handles GET /api/projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := h.db.Projects().Find(ctx, bson.M{}, opts)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch projects: "+err.Error())
		return
	}
	defer cursor.Close(ctx)

	var projects []models.Project
	if err := cursor.All(ctx, &projects); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decode projects: "+err.Error())
		return
	}

	if projects == nil {
		projects = []models.Project{}
	}

	// Populate real-time counts for each project
	for i := range projects {
		pId := projects[i].ProjectId
		inc, _ := h.db.Studies().CountDocuments(ctx, bson.M{"projectId": pId, "decision": "included"})
		exc, _ := h.db.Studies().CountDocuments(ctx, bson.M{"projectId": pId, "decision": "excluded"})
		und, _ := h.db.Studies().CountDocuments(ctx, bson.M{"projectId": pId, "decision": "undecided"})
		aiCount, _ := h.db.Studies().CountDocuments(ctx, bson.M{"projectId": pId, "aiSuggestion": bson.M{"$ne": nil}})

		projects[i].IncludedCount = int(inc)
		projects[i].ExcludedCount = int(exc)
		projects[i].UndecidedCount = int(und)
		projects[i].AIScreenedCount = int(aiCount)
		// Mask the API key in listing
		if projects[i].LLMApiKey != "" {
			projects[i].LLMApiKey = "••••••••"
		}
	}

	respondJSON(w, http.StatusOK, projects)
}

// Get handles GET /api/projects/{id}
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"projectId": id},
		},
	}
	if oid, err := bson.ObjectIDFromHex(id); err == nil {
		filter["$or"] = append(filter["$or"].([]bson.M), bson.M{"_id": oid})
	}

	var project models.Project
	err := h.db.Projects().FindOne(ctx, filter).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			respondError(w, http.StatusNotFound, "Project not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve project: "+err.Error())
		return
	}

	// Populate statistics
	pId := project.ProjectId
	inc, _ := h.db.Studies().CountDocuments(ctx, bson.M{"projectId": pId, "decision": "included"})
	exc, _ := h.db.Studies().CountDocuments(ctx, bson.M{"projectId": pId, "decision": "excluded"})
	und, _ := h.db.Studies().CountDocuments(ctx, bson.M{"projectId": pId, "decision": "undecided"})
	aiCount, _ := h.db.Studies().CountDocuments(ctx, bson.M{"projectId": pId, "aiSuggestion": bson.M{"$ne": nil}})

	project.IncludedCount = int(inc)
	project.ExcludedCount = int(exc)
	project.UndecidedCount = int(und)
	project.AIScreenedCount = int(aiCount)
	if project.LLMApiKey != "" {
		project.LLMApiKey = "••••••••"
	}

	respondJSON(w, http.StatusOK, project)
}

// Delete handles DELETE /api/projects/{id}
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Find project first to get projectId
	filter := bson.M{
		"$or": []bson.M{
			{"projectId": id},
		},
	}
	if oid, err := bson.ObjectIDFromHex(id); err == nil {
		filter["$or"] = append(filter["$or"].([]bson.M), bson.M{"_id": oid})
	}

	var project models.Project
	err := h.db.Projects().FindOne(ctx, filter).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			respondError(w, http.StatusNotFound, "Project not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to find project: "+err.Error())
		return
	}

	// Delete studies
	_, _ = h.db.Studies().DeleteMany(ctx, bson.M{"projectId": project.ProjectId})

	// Delete project
	_, err = h.db.Projects().DeleteOne(ctx, bson.M{"projectId": project.ProjectId})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete project: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Project and all associated studies successfully deleted"})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
