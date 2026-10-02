package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"accuscript-backend/internal/services"
)

func main() {
	fmt.Println("Generating sample studies.xlsx and protocol.xlsx...")

	sampleSvc := services.NewSampleService()

	// Ensure samples directory exists
	outDir := filepath.Join("..", "samples")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		outDir = "samples"
		_ = os.MkdirAll(outDir, 0755)
	}

	// 1. Generate studies.xlsx
	studiesBuf, err := sampleSvc.GenerateSampleStudiesExcel()
	if err != nil {
		log.Fatalf("Failed to generate studies excel: %v", err)
	}

	studiesPath := filepath.Join(outDir, "studies.xlsx")
	if err := os.WriteFile(studiesPath, studiesBuf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to save studies.xlsx: %v", err)
	}
	fmt.Printf("Generated: %s\n", studiesPath)

	// Also write to current directory if running in root
	_ = os.WriteFile(filepath.Join("samples", "studies.xlsx"), studiesBuf.Bytes(), 0644)

	// 2. Generate protocol.xlsx
	protocolBuf, err := sampleSvc.GenerateSampleProtocolExcel()
	if err != nil {
		log.Fatalf("Failed to generate protocol excel: %v", err)
	}

	protocolPath := filepath.Join(outDir, "protocol.xlsx")
	if err := os.WriteFile(protocolPath, protocolBuf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to save protocol.xlsx: %v", err)
	}
	fmt.Printf("Generated: %s\n", protocolPath)

	_ = os.WriteFile(filepath.Join("samples", "protocol.xlsx"), protocolBuf.Bytes(), 0644)

	fmt.Println("Sample datasets successfully created!")
}
