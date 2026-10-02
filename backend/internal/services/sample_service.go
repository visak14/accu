package services

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

type SampleService struct{}

func NewSampleService() *SampleService {
	return &SampleService{}
}

// GenerateSampleStudiesExcel creates a realistic 12-study Excel file for testing SLR AI screening
func (s *SampleService) GenerateSampleStudiesExcel() (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Studies"
	f.SetSheetName(f.GetSheetName(0), sheetName)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"0284C7"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	headers := []string{"ID", "title", "year", "author", "abstract", "article_type"}
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}
	f.SetRowStyle(sheetName, 1, 1, headerStyle)
	f.SetRowHeight(sheetName, 1, 25)

	sampleData := [][]string{
		{
			"1",
			"Continuous Glucose Monitoring and Telehealth Coaching in Adults with Type 2 Diabetes: A Randomized Clinical Trial",
			"2023",
			"Martinez, E. et al.",
			"Objective: To determine whether real-time continuous glucose monitoring (CGM) combined with remote telehealth coaching improves glycemic control in adults with poorly controlled type 2 diabetes. Methods: In this 12-month, multi-center, parallel-group, double-blind randomized controlled trial (RCT), 340 adult participants aged 18-70 with type 2 diabetes and HbA1c > 8.0% were randomized 1:1 to either CGM with weekly telehealth coaching (n=170) or standard self-monitoring of blood glucose (n=170). Results: The intervention group achieved a significant HbA1c reduction of -1.1% compared to -0.3% in standard care (mean difference -0.8%; 95% CI, -1.0% to -0.6%; P < .001). No severe hypoglycemia events occurred. Conclusion: Telehealth-assisted CGM significantly improves glycemic outcomes in adult type 2 diabetes.",
			"Journal Article",
		},
		{
			"2",
			"Cellular Mechanisms of SGLT2 Inhibitors in Murine Renal Podocytes: An In-Vitro and Animal Model",
			"2022",
			"Chen, L. & Zhao, W.",
			"Background: SGLT2 inhibitors confer renal protection in diabetes. We investigated the direct protective effects of empagliflozin on cultured mouse podocytes exposed to high-glucose media and in diabetic C57BL/6J mice. Methods: Mouse podocyte cell lines were treated with empagliflozin (10-100 nM). Apoptosis and cytoskeletal rearrangements were quantified by Western blot and immunofluorescence. Results: SGLT2 inhibition prevented actin cytoskeleton remodeling and reduced podocyte apoptosis by 45% in vitro. Conclusion: SGLT2 inhibitors provide direct cytoprotection in murine diabetic kidney cells.",
			"Journal Article",
		},
		{
			"3",
			"Digital Health Interventions for Glycemic Control in Adults with Type 2 Diabetes: A Systematic Review and Meta-Analysis",
			"2023",
			"Kowalski, J. et al.",
			"Background: Systematic synthesis of literature on digital mobile health applications for diabetes care. Methods: We searched PubMed, Embase, and Cochrane Library for RCTs published up to 2023. Results: A total of 42 RCTs involving 6,800 patients were pooled using random-effects meta-analysis. Digital interventions led to an overall weighted mean reduction in HbA1c of 0.44% (95% CI 0.32-0.56%). Conclusion: Mobile health applications are effective adjuncts for diabetes management.",
			"Systematic Review",
		},
		{
			"4",
			"Nurse-Led Mobile App Teleconsultation for Type 2 Diabetes Management in Rural Clinics: A Pragmatic RCT",
			"2024",
			"O'Connor, S. et al.",
			"Background: Access to endocrinology specialists is limited in rural populations. We evaluated a nurse-led smartphone teleconsultation protocol in adult type 2 diabetes patients. Methods: A pragmatic two-arm randomized controlled trial was conducted across 14 rural health centers (n=412 adults, mean age 58.4 years). Primary outcome was 6-month change in HbA1c. Secondary outcomes included diabetes-related distress and patient satisfaction. Results: Intervention participants demonstrated a 0.75% greater reduction in HbA1c compared to usual care (P = 0.002) and reported significantly higher satisfaction scores. Conclusion: Nurse-delivered telehealth is an effective and scalable primary care strategy for rural type 2 diabetes.",
			"Journal Article",
		},
		{
			"5",
			"Pediatric Type 1 Diabetes Mellitus and Continuous Subcutaneous Insulin Infusion: An Observational Cohort Study",
			"2021",
			"Schmidt, K. et al.",
			"Objective: To evaluate long-term glycemic stability and quality of life in pediatric patients aged 4-15 with Type 1 Diabetes initiating insulin pump therapy. Methods: Prospective observational cohort of 185 children followed for 3 years at tertiary pediatric endocrinology centers. Results: HbA1c decreased from baseline 8.6% to 7.8% at year 1 with sustained benefits. Conclusion: CSII provides durable glycemic improvement in pediatric T1D.",
			"Journal Article",
		},
		{
			"6",
			"Artificial Intelligence-Driven Smartphone Messaging for Glycemic Control in Type 2 Diabetes: A Double-Blind RCT",
			"2023",
			"Gupta, A. et al.",
			"Objective: To evaluate the efficacy of an automated conversational AI chatbot delivering personalized lifestyle and medication adherence support in adult type 2 diabetes. Methods: A 6-month prospective, double-blind randomized controlled trial (n=260 adult patients with T2D, baseline HbA1c 8.4%). Participants were allocated to AI chatbot intervention vs automated generic health SMS. Primary endpoint: change in HbA1c at 24 weeks. Results: Mean HbA1c decreased by 0.92% in the AI group vs 0.28% in control (difference -0.64%, p < 0.001). Medication adherence improved by 28%. Conclusion: AI-driven automated conversational support significantly enhances glycemic management.",
			"Journal Article",
		},
		{
			"7",
			"Editorial: The Future of Digital Therapeutics in Modern Chronic Disease Management",
			"2024",
			"Editorial Board",
			"Recent advances in remote patient monitoring and telemedicine platforms have transformed chronic disease treatment paradigms. In this editorial, we highlight key regulatory milestones, reimbursement hurdles, and future research priorities for digital therapeutics in metabolic disorders.",
			"Editorial",
		},
		{
			"8",
			"Remote Video Consultations vs Standard In-Person Follow-up for Type 2 Diabetes: A Multi-Center Non-Inferiority RCT",
			"2022",
			"Al-Mansoor, H. et al.",
			"Background: Determining the non-inferiority of telemedicine video clinics compared to traditional in-person physician visits in adult type 2 diabetes. Methods: Multi-center randomized non-inferiority trial involving 520 adult patients with established T2D. Patients were randomized 1:1 to quarterly video consultations or quarterly in-person visits over 12 months. Results: Mean HbA1c change was -0.41% in video group vs -0.45% in in-person group (margin difference within 0.15% predefined non-inferiority boundary). No difference in adverse glycemic events was observed. Conclusion: Video telemedicine visits are non-inferior to clinic visits for routine type 2 diabetes maintenance.",
			"Journal Article",
		},
	}

	for rowIdx, study := range sampleData {
		rowNum := rowIdx + 2
		for colIdx, val := range study {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowNum)
			f.SetCellValue(sheetName, cell, val)
		}
		f.SetRowHeight(sheetName, rowNum, 35)
	}

	f.SetColWidth(sheetName, "A", "A", 6)
	f.SetColWidth(sheetName, "B", "B", 35)
	f.SetColWidth(sheetName, "C", "C", 8)
	f.SetColWidth(sheetName, "D", "D", 18)
	f.SetColWidth(sheetName, "E", "E", 65)
	f.SetColWidth(sheetName, "F", "F", 16)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate sample studies excel: %w", err)
	}

	return &buf, nil
}

// GenerateSampleProtocolExcel creates a standard SLR protocol Excel file
func (s *SampleService) GenerateSampleProtocolExcel() (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Protocol"
	f.SetSheetName(f.GetSheetName(0), sheetName)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"059669"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	headers := []string{"criterion1", "criterion2", "criterion3", "criterion4", "inclusion_criteria", "exclusion_criteria"}
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}
	f.SetRowStyle(sheetName, 1, 1, headerStyle)
	f.SetRowHeight(sheetName, 1, 25)

	f.SetCellValue(sheetName, "A2", "Randomized Controlled Trial (RCT) or Parallel-Group Clinical Trial design")
	f.SetCellValue(sheetName, "B2", "Adults (aged 18+) diagnosed with Type 2 Diabetes Mellitus")
	f.SetCellValue(sheetName, "C2", "Telemedicine, mobile health app, remote monitoring, or digital health intervention vs standard care")
	f.SetCellValue(sheetName, "D2", "Reports primary quantitative glycemic outcomes (HbA1c reduction, time-in-range)")
	f.SetCellValue(sheetName, "E2", "Primary original research RCTs published from 2018-2024 evaluating remote telemedicine or digital health interventions for adult Type 2 Diabetes patients with quantitative HbA1c outcomes.")
	f.SetCellValue(sheetName, "F2", "Animal or in-vitro laboratory studies; pediatric populations (<18 years); Type 1 Diabetes only; systematic reviews, meta-analyses, editorials, letters, commentaries, or conference abstracts lacking primary data.")

	f.SetColWidth(sheetName, "A", "A", 25)
	f.SetColWidth(sheetName, "B", "B", 25)
	f.SetColWidth(sheetName, "C", "C", 30)
	f.SetColWidth(sheetName, "D", "D", 28)
	f.SetColWidth(sheetName, "E", "E", 45)
	f.SetColWidth(sheetName, "F", "F", 45)
	f.SetRowHeight(sheetName, 2, 80)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate sample protocol excel: %w", err)
	}

	return &buf, nil
}
