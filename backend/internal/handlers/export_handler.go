package handlers

import (
	"context"
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

type ExportHandler struct {
	db    *db.MongoDB
	excel *services.ExcelService
}

func NewExportHandler(database *db.MongoDB, excel *services.ExcelService) *ExportHandler {
	return &ExportHandler{
		db:    database,
		excel: excel,
	}
}

// Export handles GET /api/projects/{id}/export
func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "id")
	filterParam := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("filter")))

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// 1. Fetch Project
	var project models.Project
	err := h.db.Projects().FindOne(ctx, bson.M{"projectId": projectId}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			respondError(w, http.StatusNotFound, "Project not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve project: "+err.Error())
		return
	}

	// 2. Fetch Studies
	studyFilter := bson.M{"projectId": projectId}
	if filterParam != "" && filterParam != "all" {
		studyFilter["decision"] = filterParam
	}

	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	cursor, err := h.db.Studies().Find(ctx, studyFilter, opts)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch studies for export: "+err.Error())
		return
	}
	defer cursor.Close(ctx)

	var studies []models.Study
	if err := cursor.All(ctx, &studies); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decode studies: "+err.Error())
		return
	}

	// 3. Generate Styled Excel
	buf, err := h.excel.ExportDecisionsToExcel(&project, studies)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate Excel export: "+err.Error())
		return
	}

	safeName := strings.ReplaceAll(project.Name, " ", "_")
	safeName = strings.ReplaceAll(safeName, "/", "-")
	filename := fmt.Sprintf("SLR_Decisions_%s_%s.xlsx", safeName, time.Now().Format("20060102"))

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
