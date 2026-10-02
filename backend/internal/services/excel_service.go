package services

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"accuscript-backend/internal/models"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type ExcelService struct{}

func NewExcelService() *ExcelService {
	return &ExcelService{}
}

// ParseStudiesExcel parses studies.xlsx from reader and returns slice of Study models
func (s *ExcelService) ParseStudiesExcel(r io.Reader, projectId string) ([]models.Study, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read studies excel: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("studies excel has no sheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get rows from studies sheet: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("studies excel must have a header row and at least one data row")
	}

	headerRow := rows[0]
	headerMap := make(map[string]int)
	for idx, col := range headerRow {
		clean := strings.ToLower(strings.TrimSpace(col))
		clean = strings.ReplaceAll(clean, " ", "_")
		headerMap[clean] = idx
	}

	// Helper to find column index with variations
	findCol := func(names ...string) int {
		for _, name := range names {
			if idx, ok := headerMap[name]; ok {
				return idx
			}
		}
		return -1
	}

	idIdx := findCol("id", "study_id", "article_id", "pmid", "number")
	titleIdx := findCol("title", "study_title", "article_title", "paper_title")
	yearIdx := findCol("year", "pub_year", "publication_year", "date")
	authorIdx := findCol("author", "authors", "first_author", "creator")
	abstractIdx := findCol("abstract", "full_abstract", "summary", "description")
	typeIdx := findCol("article_type", "type", "study_type", "publication_type", "design")

	if titleIdx == -1 {
		return nil, fmt.Errorf("studies excel missing required 'title' column")
	}

	var studies []models.Study
	for rowNum := 1; rowNum < len(rows); rowNum++ {
		row := rows[rowNum]
		if len(row) == 0 {
			continue
		}

		getVal := func(idx int) string {
			if idx >= 0 && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		title := getVal(titleIdx)
		if title == "" {
			continue // skip empty rows
		}

		rawID := getVal(idIdx)
		if rawID == "" {
			rawID = strconv.Itoa(len(studies) + 1)
		}

		yearStr := getVal(yearIdx)
		year := 0
		if yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil {
				year = y
			} else {
				// Try parsing if string contains a 4-digit year
				for i := 0; i+4 <= len(yearStr); i++ {
					if y, err := strconv.Atoi(yearStr[i : i+4]); err == nil && y >= 1900 && y <= 2100 {
						year = y
						break
					}
				}
			}
		}

		author := getVal(authorIdx)
		if author == "" {
			author = "Unknown Author"
		}

		abstract := getVal(abstractIdx)
		artType := getVal(typeIdx)
		if artType == "" {
			artType = "Journal Article"
		}

		study := models.Study{
			StudyId:     "study-" + uuid.New().String()[:8],
			ProjectId:   projectId,
			RawID:       rawID,
			Title:       title,
			Year:        year,
			Author:      author,
			Abstract:    abstract,
			ArticleType: artType,
			Decision:    "undecided",
		}
		studies = append(studies, study)
	}

	if len(studies) == 0 {
		return nil, fmt.Errorf("no valid study rows found in studies excel")
	}

	return studies, nil
}

// ParseProtocolExcel parses protocol.xlsx and extracts criteria, inclusion and exclusion criteria
func (s *ExcelService) ParseProtocolExcel(r io.Reader) (*models.Protocol, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read protocol excel: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("protocol excel has no sheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get rows from protocol sheet: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("protocol excel is empty")
	}

	protocol := &models.Protocol{
		Criteria: []string{},
	}

	headerRow := rows[0]
	criteriaCols := []int{}
	inclusionIdx := -1
	exclusionIdx := -1

	for idx, col := range headerRow {
		clean := strings.ToLower(strings.TrimSpace(col))
		clean = strings.ReplaceAll(clean, " ", "_")

		if strings.Contains(clean, "inclusion") {
			inclusionIdx = idx
		} else if strings.Contains(clean, "exclusion") {
			exclusionIdx = idx
		} else if strings.HasPrefix(clean, "criterion") || strings.HasPrefix(clean, "criteria") || strings.HasPrefix(clean, "key_criterion") {
			criteriaCols = append(criteriaCols, idx)
		}
	}

	criteriaMap := make(map[string]bool)

	// Read across all rows in protocol excel
	for rowNum := 1; rowNum < len(rows); rowNum++ {
		row := rows[rowNum]
		getVal := func(idx int) string {
			if idx >= 0 && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		if inclusionIdx != -1 && protocol.InclusionCriteria == "" {
			val := getVal(inclusionIdx)
			if val != "" {
				protocol.InclusionCriteria = val
			}
		}

		if exclusionIdx != -1 && protocol.ExclusionCriteria == "" {
			val := getVal(exclusionIdx)
			if val != "" {
				protocol.ExclusionCriteria = val
			}
		}

		for _, colIdx := range criteriaCols {
			val := getVal(colIdx)
			if val != "" && !criteriaMap[val] {
				criteriaMap[val] = true
				protocol.Criteria = append(protocol.Criteria, val)
			}
		}
	}

	// Fallback if structured rows were organized in 2-column key/value layout
	if protocol.InclusionCriteria == "" || len(protocol.Criteria) == 0 {
		for _, row := range rows {
			if len(row) >= 2 {
				key := strings.ToLower(strings.TrimSpace(row[0]))
				val := strings.TrimSpace(row[1])
				if strings.Contains(key, "inclusion") && protocol.InclusionCriteria == "" {
					protocol.InclusionCriteria = val
				} else if strings.Contains(key, "exclusion") && protocol.ExclusionCriteria == "" {
					protocol.ExclusionCriteria = val
				} else if (strings.HasPrefix(key, "criterion") || strings.HasPrefix(key, "criteria")) && val != "" {
					if !criteriaMap[val] {
						criteriaMap[val] = true
						protocol.Criteria = append(protocol.Criteria, val)
					}
				}
			}
		}
	}

	// Default fallbacks if empty
	if protocol.InclusionCriteria == "" {
		protocol.InclusionCriteria = "Primary empirical studies meeting research scope and quality standards."
	}
	if protocol.ExclusionCriteria == "" {
		protocol.ExclusionCriteria = "Reviews, editorials, letters, abstracts without data, animal/in-vitro models."
	}
	if len(protocol.Criteria) == 0 {
		protocol.Criteria = []string{"Relevant Population", "Target Intervention", "Comparative Outcomes"}
	}

	return protocol, nil
}

// ExportDecisionsToExcel generates an Excel workbook for the studies in a project
func (s *ExcelService) ExportDecisionsToExcel(project *models.Project, studies []models.Study) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Screening Decisions"
	defaultSheet := f.GetSheetName(0)
	f.SetSheetName(defaultSheet, sheetName)

	// Styling definitions
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"1E293B"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "475569", Style: 2},
		},
	})

	includeStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "15803D"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"DCFCE7"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	excludeStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "B91C1C"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FEE2E2"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	undecidedStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "64748B"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F1F5F9"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	bodyStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
	})

	centerBodyStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "top"},
	})

	headers := []string{
		"ID",
		"Title",
		"Author",
		"Year",
		"Article Type",
		"Abstract",
		"Decision",
		"AI Suggestion",
		"AI Confidence",
		"AI Reasoning",
		"Matched Criteria",
		"Decided At",
	}

	for colIdx, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cellName, header)
	}
	f.SetRowStyle(sheetName, 1, 1, headerStyle)
	f.SetRowHeight(sheetName, 1, 28)

	for rowIdx, study := range studies {
		rowNum := rowIdx + 2

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), study.RawID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), study.Title)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), study.Author)
		if study.Year > 0 {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), study.Year)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), study.ArticleType)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), study.Abstract)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), strings.ToUpper(study.Decision))

		aiSugg := "-"
		if study.AISuggestion != nil {
			aiSugg = strings.ToUpper(*study.AISuggestion)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), aiSugg)

		aiConf := "-"
		if study.AIConfidence != nil {
			aiConf = fmt.Sprintf("%.0f%%", *study.AIConfidence*100)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), aiConf)

		aiReason := "-"
		if study.AIReason != nil {
			aiReason = *study.AIReason
		}
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowNum), aiReason)

		matchesStr := "-"
		if len(study.AIMatches) > 0 {
			matchesStr = strings.Join(study.AIMatches, "; ")
		}
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowNum), matchesStr)

		decidedAtStr := "-"
		if study.DecidedAt != nil {
			decidedAtStr = study.DecidedAt.Format("2006-01-02 15:04:05")
		}
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowNum), decidedAtStr)

		// Apply styles
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("A%d", rowNum), centerBodyStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", rowNum), fmt.Sprintf("C%d", rowNum), bodyStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", rowNum), fmt.Sprintf("E%d", rowNum), centerBodyStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", rowNum), fmt.Sprintf("F%d", rowNum), bodyStyle)

		decisionCell := fmt.Sprintf("G%d", rowNum)
		switch study.Decision {
		case "included":
			f.SetCellStyle(sheetName, decisionCell, decisionCell, includeStyle)
		case "excluded":
			f.SetCellStyle(sheetName, decisionCell, decisionCell, excludeStyle)
		default:
			f.SetCellStyle(sheetName, decisionCell, decisionCell, undecidedStyle)
		}

		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", rowNum), fmt.Sprintf("I%d", rowNum), centerBodyStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("J%d", rowNum), fmt.Sprintf("K%d", rowNum), bodyStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("L%d", rowNum), fmt.Sprintf("L%d", rowNum), centerBodyStyle)

		f.SetRowHeight(sheetName, rowNum, 40)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 8)
	f.SetColWidth(sheetName, "B", "B", 35)
	f.SetColWidth(sheetName, "C", "C", 18)
	f.SetColWidth(sheetName, "D", "D", 10)
	f.SetColWidth(sheetName, "E", "E", 16)
	f.SetColWidth(sheetName, "F", "F", 50)
	f.SetColWidth(sheetName, "G", "G", 14)
	f.SetColWidth(sheetName, "H", "H", 16)
	f.SetColWidth(sheetName, "I", "I", 14)
	f.SetColWidth(sheetName, "J", "J", 45)
	f.SetColWidth(sheetName, "K", "K", 30)
	f.SetColWidth(sheetName, "L", "L", 20)

	// Add second sheet with Protocol Summary
	protocolSheet := "Review Protocol"
	f.NewSheet(protocolSheet)
	f.SetCellValue(protocolSheet, "A1", "Project Name")
	f.SetCellValue(protocolSheet, "B1", project.Name)
	f.SetCellValue(protocolSheet, "A2", "Description")
	f.SetCellValue(protocolSheet, "B2", project.Description)
	f.SetCellValue(protocolSheet, "A3", "Total Studies")
	f.SetCellValue(protocolSheet, "B3", project.TotalStudies)
	f.SetCellValue(protocolSheet, "A4", "Inclusion Criteria")
	f.SetCellValue(protocolSheet, "B4", project.Protocol.InclusionCriteria)
	f.SetCellValue(protocolSheet, "A5", "Exclusion Criteria")
	f.SetCellValue(protocolSheet, "B5", project.Protocol.ExclusionCriteria)
	f.SetCellValue(protocolSheet, "A6", "Key Criteria")
	f.SetCellValue(protocolSheet, "B6", strings.Join(project.Protocol.Criteria, ", "))
	f.SetCellValue(protocolSheet, "A7", "Export Generated")
	f.SetCellValue(protocolSheet, "B7", time.Now().Format("2006-01-02 15:04:05 MST"))

	f.SetColWidth(protocolSheet, "A", "A", 20)
	f.SetColWidth(protocolSheet, "B", "B", 60)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to encode excel workbook: %w", err)
	}

	return &buf, nil
}
