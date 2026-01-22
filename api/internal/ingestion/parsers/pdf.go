package parsers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"strings"

	// PDF parsing requires go-fitz - make it optional
	// "github.com/gen2brain/go-fitz"
)

/* PDFParser parses PDF files and extracts text */
type PDFParser struct {
	extractImages bool
}

/* NewPDFParser creates a new PDF parser */
func NewPDFParser(extractImages bool) *PDFParser {
	return &PDFParser{
		extractImages: extractImages,
	}
}

/* Parse parses a PDF file and extracts text content */
func (p *PDFParser) Parse(reader io.Reader) (*PDFContent, error) {
	// Read all bytes from reader
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF data: %w", err)
	}

	// Verify PDF format
	if len(data) < 5 || string(data[0:5]) != "%PDF-" {
		return nil, fmt.Errorf("invalid PDF format")
	}

	// Use basic PDF parsing (go-fitz dependency removed for build compatibility)
	// For advanced PDF parsing with go-fitz, uncomment and add dependency
	return p.parseBasicPDF(data)
}

/* parseBasicPDF provides fallback parsing when go-fitz is unavailable */
func (p *PDFParser) parseBasicPDF(data []byte) (*PDFContent, error) {
	content := &PDFContent{
		Pages:    []PageContent{},
		Metadata: make(map[string]interface{}),
	}

	dataStr := string(data)

	// Basic metadata extraction
	if strings.Contains(dataStr, "/Title") {
		titleStart := strings.Index(dataStr, "/Title")
		if titleStart > 0 && titleStart < len(dataStr)-20 {
			titleEnd := strings.Index(dataStr[titleStart:], "\n")
			if titleEnd > 0 {
				titlePart := dataStr[titleStart+6 : titleStart+titleEnd]
				content.Metadata["title"] = strings.Trim(titlePart, "()")
			}
		}
	}

	if strings.Contains(dataStr, "/Author") {
		authorStart := strings.Index(dataStr, "/Author")
		if authorStart > 0 && authorStart < len(dataStr)-20 {
			authorEnd := strings.Index(dataStr[authorStart:], "\n")
			if authorEnd > 0 {
				authorPart := dataStr[authorStart+7 : authorStart+authorEnd]
				content.Metadata["author"] = strings.Trim(authorPart, "()")
			}
		}
	}

	// Count pages
	pageCount := strings.Count(dataStr, "/Type/Page") + strings.Count(dataStr, "/Type /Page")
	if pageCount == 0 {
		pageCount = 1
	}
	content.Metadata["page_count"] = pageCount

	// Extract text using basic method
	var fullText strings.Builder
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		pageText := p.extractTextFromPage(data, pageNum)
		pageContent := PageContent{
			PageNumber: pageNum,
			Text:       pageText,
			Images:     []string{},
		}
		content.Pages = append(content.Pages, pageContent)
		if pageText != "" {
			fullText.WriteString(pageText)
			fullText.WriteString("\n")
		}
	}

	content.Text = fullText.String()
	return content, nil
}

/* imageToBase64 converts an image to base64 encoded PNG */
func (p *PDFParser) imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

/* PDFContent represents extracted PDF content */
type PDFContent struct {
	Pages    []PageContent           `json:"pages"`
	Metadata map[string]interface{}  `json:"metadata"`
	Text     string                   `json:"text"` // Full text concatenation
}

/* PageContent represents content from a single PDF page */
type PageContent struct {
	PageNumber int    `json:"page_number"`
	Text       string `json:"text"`
	Images     []string `json:"images,omitempty"` // Base64 encoded images if extractImages=true
}

/* ExtractText extracts plain text from PDF content */
func (p *PDFContent) ExtractText() string {
	if p.Text != "" {
		return p.Text
	}
	
	var builder strings.Builder
	for _, page := range p.Pages {
		builder.WriteString(page.Text)
		builder.WriteString("\n")
	}
	
	return builder.String()
}

/* extractTextFromPage extracts text from a specific PDF page */
func (p *PDFParser) extractTextFromPage(data []byte, pageNum int) string {
	dataStr := string(data)
	var textBuilder strings.Builder
	
	// Look for text objects in PDF content streams
	// PDF text is typically in content streams between stream/endstream
	// Text operators: Tj, TJ, ', "
	
	// Find text between parentheses (common in PDF text objects)
	// Pattern: (text content) Tj or (text) TJ
	start := 0
	for {
		// Find opening parenthesis
		openIdx := strings.Index(dataStr[start:], "(")
		if openIdx == -1 {
			break
		}
		openIdx += start
		
		// Find closing parenthesis
		closeIdx := strings.Index(dataStr[openIdx:], ")")
		if closeIdx == -1 {
			break
		}
		closeIdx += openIdx
		
		// Extract text between parentheses
		text := dataStr[openIdx+1 : closeIdx]
		
		// Check if followed by text operator (Tj, TJ, ', ")
		afterClose := closeIdx + 1
		if afterClose < len(dataStr) {
			remaining := dataStr[afterClose:]
			// Look for text operators nearby (within 20 chars)
			if len(remaining) > 0 {
				checkLen := 20
				if len(remaining) < checkLen {
					checkLen = len(remaining)
				}
				checkStr := remaining[:checkLen]
				if strings.Contains(checkStr, "Tj") || strings.Contains(checkStr, "TJ") ||
					strings.Contains(checkStr, "'") || strings.Contains(checkStr, "\"") {
					// This looks like PDF text
					// Unescape PDF string escapes
					text = unescapePDFString(text)
					if text != "" {
						textBuilder.WriteString(text)
						textBuilder.WriteString(" ")
					}
				}
			}
		}
		
		start = closeIdx + 1
	}
	
	result := textBuilder.String()
	if result == "" {
		// Fallback: try to find readable text patterns
		result = p.extractReadableText(dataStr)
	}
	
	return strings.TrimSpace(result)
}

/* unescapePDFString unescapes PDF string escape sequences */
func unescapePDFString(s string) string {
	// PDF strings use backslash escapes
	result := strings.ReplaceAll(s, "\\n", "\n")
	result = strings.ReplaceAll(result, "\\r", "\r")
	result = strings.ReplaceAll(result, "\\t", "\t")
	result = strings.ReplaceAll(result, "\\(", "(")
	result = strings.ReplaceAll(result, "\\)", ")")
	result = strings.ReplaceAll(result, "\\\\", "\\")
	
	// Remove octal escapes (simplified)
	// In production, properly decode all PDF escape sequences
	
	return result
}

/* extractReadableText extracts readable text patterns from PDF */
func (p *PDFParser) extractReadableText(dataStr string) string {
	var textBuilder strings.Builder
	
	// Look for readable ASCII text sequences (words)
	// This is a fallback for when proper text extraction fails
	words := strings.Fields(dataStr)
	for _, word := range words {
		// Check if word looks like readable text (contains letters)
		if hasLetters(word) && len(word) > 2 {
			// Filter out PDF keywords and operators
			if !isPDFKeyword(word) {
				textBuilder.WriteString(word)
				textBuilder.WriteString(" ")
			}
		}
	}
	
	return strings.TrimSpace(textBuilder.String())
}

/* hasLetters checks if string contains letters */
func hasLetters(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

/* isPDFKeyword checks if string is a PDF keyword/operator */
func isPDFKeyword(s string) bool {
	keywords := []string{"obj", "endobj", "stream", "endstream", "xref", "trailer",
		"startxref", "PDF", "xref", "BT", "ET", "Tj", "TJ", "Td", "Tm", "Tf",
		"q", "Q", "cm", "rg", "RG", "g", "G", "w", "W", "m", "l", "c", "v", "y", "h", "S", "s", "f", "F"}
	for _, kw := range keywords {
		if s == kw {
			return true
		}
	}
	return false
}
