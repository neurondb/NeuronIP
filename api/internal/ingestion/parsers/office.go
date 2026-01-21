package parsers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
)

/* OfficeParser parses Microsoft Office documents (Word, Excel, PowerPoint) */
type OfficeParser struct {
	extractImages bool
}

/* NewOfficeParser creates a new Office document parser */
func NewOfficeParser(extractImages bool) *OfficeParser {
	return &OfficeParser{
		extractImages: extractImages,
	}
}

/* Parse parses an Office document based on file type */
func (p *OfficeParser) Parse(reader io.Reader, fileType string) (*OfficeContent, error) {
	switch strings.ToLower(fileType) {
	case "docx", "doc":
		return p.parseWord(reader)
	case "xlsx", "xls":
		return p.parseExcel(reader)
	case "pptx", "ppt":
		return p.parsePowerPoint(reader)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
}

/* parseWord parses a Word document */
func (p *OfficeParser) parseWord(reader io.Reader) (*OfficeContent, error) {
	// Read all bytes
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read Word document: %w", err)
	}

	content := &OfficeContent{
		Type:     "word",
		Sections: []SectionContent{},
		Metadata: make(map[string]interface{}),
	}

	// DOCX is a ZIP archive containing XML files
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open DOCX as ZIP: %w", err)
	}

	var fullText strings.Builder
	sectionLevel := 0

	// Parse document.xml for content
	for _, file := range zipReader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			if err != nil {
				continue
			}

			// Read XML content
			xmlData, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// Parse XML to extract text (simplified)
			// In production, use proper DOCX library like github.com/unidoc/unioffice
			xmlStr := string(xmlData)
			
			// Extract text between <w:t> tags (simplified approach)
			textParts := extractTextFromXML(xmlStr)
			for _, text := range textParts {
				if text != "" {
					fullText.WriteString(text)
					fullText.WriteString(" ")
				}
			}

			// Create sections from paragraphs
			paragraphs := extractParagraphsFromXML(xmlStr)
			for _, paraText := range paragraphs {
				if paraText != "" {
					sectionLevel++
					section := SectionContent{
						Title:   "",
						Content: paraText,
						Level:   sectionLevel,
					}
					content.Sections = append(content.Sections, section)
				}
			}
		}

		// Extract metadata from core.xml
		if file.Name == "docProps/core.xml" {
			rc, err := file.Open()
			if err == nil {
				xmlData, _ := io.ReadAll(rc)
				rc.Close()
				extractMetadataFromXML(string(xmlData), content.Metadata)
			}
		}
	}

	content.Text = fullText.String()
	return content, nil
}

/* extractTextFromXML extracts text from DOCX XML */
func extractTextFromXML(xmlStr string) []string {
	var texts []string
	// Simple text extraction between <w:t> tags
	start := 0
	for {
		startIdx := strings.Index(xmlStr[start:], "<w:t>")
		if startIdx == -1 {
			break
		}
		startIdx += start
		endIdx := strings.Index(xmlStr[startIdx:], "</w:t>")
		if endIdx == -1 {
			break
		}
		text := xmlStr[startIdx+5 : startIdx+endIdx]
		// Unescape XML entities
		text = strings.ReplaceAll(text, "&amp;", "&")
		text = strings.ReplaceAll(text, "&lt;", "<")
		text = strings.ReplaceAll(text, "&gt;", ">")
		text = strings.ReplaceAll(text, "&quot;", "\"")
		texts = append(texts, text)
		start = startIdx + endIdx + 7
	}
	return texts
}

/* extractParagraphsFromXML extracts paragraphs from DOCX XML */
func extractParagraphsFromXML(xmlStr string) []string {
	var paragraphs []string
	var currentPara strings.Builder
	
	texts := extractTextFromXML(xmlStr)
	for _, text := range texts {
		if strings.Contains(xmlStr, "<w:p>") {
			if currentPara.Len() > 0 {
				paragraphs = append(paragraphs, currentPara.String())
				currentPara.Reset()
			}
		}
		currentPara.WriteString(text)
		currentPara.WriteString(" ")
	}
	
	if currentPara.Len() > 0 {
		paragraphs = append(paragraphs, currentPara.String())
	}
	
	return paragraphs
}

/* extractMetadataFromXML extracts metadata from DOCX core.xml */
func extractMetadataFromXML(xmlStr string, metadata map[string]interface{}) {
	// Extract common metadata fields
	metadataFields := map[string]string{
		"dc:title":   "title",
		"dc:creator": "author",
		"dc:subject": "subject",
		"cp:keywords": "keywords",
	}
	
	for xmlKey, metaKey := range metadataFields {
		startIdx := strings.Index(xmlStr, "<"+xmlKey+">")
		if startIdx > 0 {
			startIdx += len(xmlKey) + 2
			endIdx := strings.Index(xmlStr[startIdx:], "</"+xmlKey+">")
			if endIdx > 0 {
				value := xmlStr[startIdx : startIdx+endIdx]
				metadata[metaKey] = value
			}
		}
	}
}

/* parseExcel parses an Excel spreadsheet */
func (p *OfficeParser) parseExcel(reader io.Reader) (*OfficeContent, error) {
	// Read all bytes
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel document: %w", err)
	}

	content := &OfficeContent{
		Type:     "excel",
		Sheets:   []SheetContent{},
		Metadata: make(map[string]interface{}),
	}

	// XLSX is a ZIP archive containing XML files
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX as ZIP: %w", err)
	}

	// Find workbook and sheets
	var sheetNames []string
	sheetMap := make(map[string]*SheetContent)
	var fullText strings.Builder

	for _, file := range zipReader.File {
		// Parse workbook to get sheet names
		if file.Name == "xl/workbook.xml" {
			rc, err := file.Open()
			if err == nil {
				xmlData, _ := io.ReadAll(rc)
				rc.Close()
				sheetNames = extractSheetNamesFromXML(string(xmlData))
				for _, name := range sheetNames {
					sheetMap[name] = &SheetContent{
						Name:  name,
						Rows:  [][]interface{}{},
						Range: "",
					}
				}
			}
		}

		// Parse sheet data (xl/worksheets/sheet*.xml)
		if strings.HasPrefix(file.Name, "xl/worksheets/sheet") && strings.HasSuffix(file.Name, ".xml") {
			rc, err := file.Open()
			if err != nil {
				continue
			}

			xmlData, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			// Extract sheet name from file path
			sheetIdx := strings.TrimPrefix(file.Name, "xl/worksheets/sheet")
			sheetIdx = strings.TrimSuffix(sheetIdx, ".xml")
			
			// Get sheet name (simplified - in production use proper mapping)
			sheetName := fmt.Sprintf("Sheet%s", sheetIdx)
			if len(sheetNames) > 0 {
				// Use first available name
				if len(sheetNames) > 0 {
					sheetName = sheetNames[0]
				}
			}

			sheet, exists := sheetMap[sheetName]
			if !exists {
				sheet = &SheetContent{
					Name:  sheetName,
					Rows:  [][]interface{}{},
					Range: "",
				}
				sheetMap[sheetName] = sheet
			}

			// Extract rows from XML
			rows := extractRowsFromExcelXML(string(xmlData))
			for _, row := range rows {
				rowData := make([]interface{}, len(row))
				for i, cell := range row {
					rowData[i] = cell
					if cell != "" {
						fullText.WriteString(cell)
						fullText.WriteString("\t")
					}
				}
				fullText.WriteString("\n")
				sheet.Rows = append(sheet.Rows, rowData)
			}
		}

		// Extract metadata
		if file.Name == "docProps/core.xml" {
			rc, err := file.Open()
			if err == nil {
				xmlData, _ := io.ReadAll(rc)
				rc.Close()
				extractMetadataFromXML(string(xmlData), content.Metadata)
			}
		}
	}

	// Convert sheet map to slice
	for _, sheet := range sheetMap {
		content.Sheets = append(content.Sheets, *sheet)
	}
	content.Metadata["sheet_count"] = len(content.Sheets)
	content.Metadata["sheet_names"] = sheetNames

	content.Text = fullText.String()
	return content, nil
}

/* extractSheetNamesFromXML extracts sheet names from Excel workbook.xml */
func extractSheetNamesFromXML(xmlStr string) []string {
	var names []string
	// Simple extraction - look for sheet names in workbook
	// In production, use proper XML parser
	start := 0
	for {
		startIdx := strings.Index(xmlStr[start:], "name=")
		if startIdx == -1 {
			break
		}
		startIdx += start + 5
		if startIdx < len(xmlStr) {
			quote := xmlStr[startIdx]
			if quote == '"' || quote == '\'' {
				endIdx := strings.Index(xmlStr[startIdx+1:], string(quote))
				if endIdx > 0 {
					name := xmlStr[startIdx+1 : startIdx+1+endIdx]
					names = append(names, name)
					start = startIdx + 1 + endIdx
				} else {
					break
				}
			} else {
				break
			}
		} else {
			break
		}
	}
	return names
}

/* extractRowsFromExcelXML extracts rows from Excel sheet XML */
func extractRowsFromExcelXML(xmlStr string) [][]string {
	var rows [][]string
	
	// Parse XML to extract cell values
	// Simplified approach - in production use proper XML parser
	// Look for <c> tags (cells) and <v> tags (values)
	
	// Split by row tags
	rowParts := strings.Split(xmlStr, "<row")
	for _, rowPart := range rowParts[1:] { // Skip first empty part
		var row []string
		
		// Extract values between <v> tags
		start := 0
		for {
			startIdx := strings.Index(rowPart[start:], "<v>")
			if startIdx == -1 {
				break
			}
			startIdx += start
			endIdx := strings.Index(rowPart[startIdx:], "</v>")
			if endIdx > 0 {
				value := rowPart[startIdx+3 : startIdx+endIdx]
				row = append(row, value)
				start = startIdx + endIdx + 4
			} else {
				break
			}
		}
		
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}
	
	return rows
}

/* parsePowerPoint parses a PowerPoint presentation */
func (p *OfficeParser) parsePowerPoint(reader io.Reader) (*OfficeContent, error) {
	// Read all bytes
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read PowerPoint document: %w", err)
	}

	content := &OfficeContent{
		Type:     "powerpoint",
		Slides:   []SlideContent{},
		Metadata: make(map[string]interface{}),
	}

	// PPTX is a ZIP archive containing XML files
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open PPTX as ZIP: %w", err)
	}

	var fullText strings.Builder
	slideNum := 0

	// Parse presentation slides
	for _, file := range zipReader.File {
		// Parse slide files (ppt/slides/slide*.xml)
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			slideNum++
			rc, err := file.Open()
			if err != nil {
				continue
			}

			xmlData, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			slideContent := SlideContent{
				SlideNumber: slideNum,
				Title:       "",
				Content:     "",
				Notes:       "",
			}

			// Extract text from slide XML
			slideText := extractTextFromXML(string(xmlData))
			slideContent.Content = strings.Join(slideText, " ")
			
			// Try to identify title (usually first text element)
			if len(slideText) > 0 {
				slideContent.Title = slideText[0]
			}

			fullText.WriteString(fmt.Sprintf("Slide %d: %s\n", slideNum, slideContent.Title))
			fullText.WriteString(slideContent.Content)
			fullText.WriteString("\n\n")

			content.Slides = append(content.Slides, slideContent)
		}

		// Extract metadata
		if file.Name == "docProps/core.xml" {
			rc, err := file.Open()
			if err == nil {
				xmlData, _ := io.ReadAll(rc)
				rc.Close()
				extractMetadataFromXML(string(xmlData), content.Metadata)
			}
		}
	}

	content.Text = fullText.String()
	content.Metadata["slide_count"] = slideNum

	return content, nil
}

/* OfficeContent represents extracted Office document content */
type OfficeContent struct {
	Type     string                   `json:"type"` // "word", "excel", "powerpoint"
	Text     string                   `json:"text"` // Full text content
	Metadata map[string]interface{}  `json:"metadata"`
	
	// Word-specific
	Sections []SectionContent `json:"sections,omitempty"`
	
	// Excel-specific
	Sheets []SheetContent `json:"sheets,omitempty"`
	
	// PowerPoint-specific
	Slides []SlideContent `json:"slides,omitempty"`
}

/* SectionContent represents a section in a Word document */
type SectionContent struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content"`
	Level   int    `json:"level,omitempty"`
}

/* SheetContent represents a sheet in an Excel workbook */
type SheetContent struct {
	Name  string                   `json:"name"`
	Rows  [][]interface{}          `json:"rows"`
	Range string                   `json:"range,omitempty"`
}

/* SlideContent represents a slide in a PowerPoint presentation */
type SlideContent struct {
	SlideNumber int    `json:"slide_number"`
	Title       string `json:"title,omitempty"`
	Content     string `json:"content"`
	Notes       string `json:"notes,omitempty"`
}

/* ExtractText extracts plain text from Office content */
func (o *OfficeContent) ExtractText() string {
	if o.Text != "" {
		return o.Text
	}
	
	var builder strings.Builder
	
	switch o.Type {
	case "word":
		for _, section := range o.Sections {
			if section.Title != "" {
				builder.WriteString(section.Title)
				builder.WriteString("\n")
			}
			builder.WriteString(section.Content)
			builder.WriteString("\n\n")
		}
	case "excel":
		for _, sheet := range o.Sheets {
			builder.WriteString(fmt.Sprintf("Sheet: %s\n", sheet.Name))
			for _, row := range sheet.Rows {
				for _, cell := range row {
					builder.WriteString(fmt.Sprintf("%v\t", cell))
				}
				builder.WriteString("\n")
			}
			builder.WriteString("\n")
		}
	case "powerpoint":
		for _, slide := range o.Slides {
			if slide.Title != "" {
				builder.WriteString(fmt.Sprintf("Slide %d: %s\n", slide.SlideNumber, slide.Title))
			}
			builder.WriteString(slide.Content)
			builder.WriteString("\n\n")
		}
	}
	
	return builder.String()
}
