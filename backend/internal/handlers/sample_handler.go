package handlers

import (
	"fmt"
	"net/http"

	"accuscript-backend/internal/services"
)

type SampleHandler struct {
	sampleService *services.SampleService
}

func NewSampleHandler(sampleService *services.SampleService) *SampleHandler {
	return &SampleHandler{sampleService: sampleService}
}

// DownloadStudies handles GET /api/sample-files/studies
func (h *SampleHandler) DownloadStudies(w http.ResponseWriter, r *http.Request) {
	buf, err := h.sampleService.GenerateSampleStudiesExcel()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate sample studies excel: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\"sample_studies.xlsx\"")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

// DownloadProtocol handles GET /api/sample-files/protocol
func (h *SampleHandler) DownloadProtocol(w http.ResponseWriter, r *http.Request) {
	buf, err := h.sampleService.GenerateSampleProtocolExcel()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate sample protocol excel: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\"sample_protocol.xlsx\"")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
