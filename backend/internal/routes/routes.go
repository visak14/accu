package routes

import (
	"encoding/json"
	"net/http"

	"accuscript-backend/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type RouterDeps struct {
	ProjectHandler *handlers.ProjectHandler
	StudyHandler   *handlers.StudyHandler
	AIHandler      *handlers.AIHandler
	ExportHandler  *handlers.ExportHandler
	SampleHandler  *handlers.SampleHandler
}

func SetupRouter(deps *RouterDeps) *chi.Mux {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Enable CORS for frontend development and production
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "Content-Disposition", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API Routes
	r.Route("/api", func(api chi.Router) {
		// Health check
		api.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status":  "ok",
				"service": "AccuScript SLR Screening API",
				"version": "1.0.0",
			})
		})

		// Sample files
		api.Get("/sample-files/studies", deps.SampleHandler.DownloadStudies)
		api.Get("/sample-files/protocol", deps.SampleHandler.DownloadProtocol)

		// Projects
		api.Route("/projects", func(p chi.Router) {
			p.Post("/upload", deps.ProjectHandler.Upload)
			p.Get("/", deps.ProjectHandler.List)
			p.Get("/{id}", deps.ProjectHandler.Get)
			p.Delete("/{id}", deps.ProjectHandler.Delete)
			p.Get("/{id}/studies", deps.StudyHandler.ListByProject)
			p.Get("/{id}/export", deps.ExportHandler.Export)
			p.Post("/{id}/batch-ai-suggest", deps.AIHandler.BatchSuggest)
		})

		// Studies
		api.Route("/studies", func(s chi.Router) {
			s.Get("/{studyId}", deps.StudyHandler.Get)
			s.Patch("/{studyId}/decision", deps.StudyHandler.UpdateDecision)
			s.Post("/{studyId}/ai-suggest", deps.AIHandler.Suggest)
		})
	})

	return r
}
