package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"accuscript-backend/internal/config"
	"accuscript-backend/internal/db"
	"accuscript-backend/internal/handlers"
	"accuscript-backend/internal/routes"
	"accuscript-backend/internal/services"
)

func main() {
	log.Println("=== Starting AccuScript SLR AI Literature Screener Backend ===")

	// 1. Load Config
	cfg := config.LoadConfig()
	log.Printf("Configuration loaded. Port: %s, MongoURI: %s, DB: %s, Default Provider: %s",
		cfg.Port, cfg.MongoURI, cfg.DBName, cfg.DefaultLLMProv)

	// 2. Initialize Database
	database, err := db.ConnectMongo(cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("Fatal: Could not connect to MongoDB: %v\nCheck your MONGO_URI in .env or ensure MongoDB is running.", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = database.Close(ctx)
	}()
	log.Println("Connected to MongoDB successfully.")

	// 3. Initialize Services
	cryptoSvc := services.NewCryptoService(cfg.SecretKey)
	excelSvc := services.NewExcelService()
	sampleSvc := services.NewSampleService()
	llmSvc := services.NewLLMService(cfg)

	// 4. Initialize Handlers
	projectH := handlers.NewProjectHandler(database, cryptoSvc, excelSvc)
	studyH := handlers.NewStudyHandler(database)
	aiH := handlers.NewAIHandler(database, llmSvc, cryptoSvc)
	exportH := handlers.NewExportHandler(database, excelSvc)
	sampleH := handlers.NewSampleHandler(sampleSvc)

	// 5. Setup Chi Router
	routerDeps := &routes.RouterDeps{
		ProjectHandler: projectH,
		StudyHandler:   studyH,
		AIHandler:      aiH,
		ExportHandler:  exportH,
		SampleHandler:  sampleH,
	}
	r := routes.SetupRouter(routerDeps)

	// 6. Start HTTP Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("HTTP Server listening on http://localhost:%s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 7. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
