package services

import (
	"bytes"
	"os"
	"testing"
	"time"

	"accuscript-backend/internal/models"
)

func TestExcelService_ParseAndExport(t *testing.T) {
	excelSvc := NewExcelService()
	sampleSvc := NewSampleService()

	// 1. Generate sample studies buffer
	studiesBuf, err := sampleSvc.GenerateSampleStudiesExcel()
	if err != nil {
		t.Fatalf("Failed to generate sample studies: %v", err)
	}

	// 2. Parse studies buffer
	studies, err := excelSvc.ParseStudiesExcel(bytes.NewReader(studiesBuf.Bytes()), "proj-test1")
	if err != nil {
		t.Fatalf("Failed to parse studies excel: %v", err)
	}

	if len(studies) != 8 {
		t.Fatalf("Expected 8 parsed studies, got %d", len(studies))
	}

	if studies[0].Title == "" || studies[0].Abstract == "" {
		t.Fatalf("Study title or abstract was empty")
	}

	// 3. Generate sample protocol buffer
	protocolBuf, err := sampleSvc.GenerateSampleProtocolExcel()
	if err != nil {
		t.Fatalf("Failed to generate sample protocol: %v", err)
	}

	// 4. Parse protocol buffer
	protocol, err := excelSvc.ParseProtocolExcel(bytes.NewReader(protocolBuf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to parse protocol excel: %v", err)
	}

	if len(protocol.Criteria) == 0 {
		t.Fatalf("Expected protocol criteria to be populated")
	}

	if protocol.InclusionCriteria == "" || protocol.ExclusionCriteria == "" {
		t.Fatalf("Expected inclusion and exclusion criteria to be populated")
	}

	// 5. Test Export
	now := time.Now()
	testProject := &models.Project{
		ProjectId:    "proj-test1",
		Name:         "Diabetes Telemedicine Review",
		Description:  "Sample SLR review",
		TotalStudies: len(studies),
		Protocol:     *protocol,
		CreatedAt:    now,
	}

	// Set sample decisions and AI recommendations
	sugg := "include"
	conf := 0.95
	reason := "Matches RCT design and telemedicine criteria"
	studies[0].Decision = "included"
	studies[0].AISuggestion = &sugg
	studies[0].AIConfidence = &conf
	studies[0].AIReason = &reason
	studies[0].AIMatches = []string{"Randomized Controlled Trial", "Type 2 Diabetes"}
	studies[0].DecidedAt = &now

	exportBuf, err := excelSvc.ExportDecisionsToExcel(testProject, studies)
	if err != nil {
		t.Fatalf("Failed to generate excel export: %v", err)
	}

	if exportBuf.Len() == 0 {
		t.Fatalf("Export buffer is empty")
	}
}

func TestExcelService_RealSampleFiles(t *testing.T) {
	excelSvc := NewExcelService()

	studiesFile, err := os.Open("../../samples/studies.xlsx")
	if err != nil {
		t.Skip("Sample studies file not found, skipping file test")
	}
	defer studiesFile.Close()

	studies, err := excelSvc.ParseStudiesExcel(studiesFile, "proj-file-test")
	if err != nil {
		t.Fatalf("Failed to parse samples/studies.xlsx: %v", err)
	}
	if len(studies) == 0 {
		t.Fatalf("Expected non-empty studies from sample file")
	}
}
