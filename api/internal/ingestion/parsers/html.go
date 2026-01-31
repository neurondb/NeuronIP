package parsers

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

/* HTMLParser parses HTML documents and extracts text content */
type HTMLParser struct {
	extractMetadata bool
}

/* NewHTMLParser creates a new HTML parser */
func NewHTMLParser(extractMetadata bool) *HTMLParser {
	return &HTMLParser{
		extractMetadata: extractMetadata,
	}
}

/* HTMLContent represents parsed HTML content */
type HTMLContent struct {
	Title       string                 `json:"title"`
	Text        string                 `json:"text"`
	Sections    []SectionContent       `json:"sections,omitempty"`
	Links       []string               `json:"links,omitempty"`
	Metadata    map[string]interface{}  `json:"metadata,omitempty"`
}

/* Parse parses an HTML document */
func (p *HTMLParser) Parse(reader io.Reader) (*HTMLContent, error) {
	content := &HTMLContent{
		Sections: []SectionContent{},
		Links:    []string{},
		Metadata: make(map[string]interface{}),
	}
	
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	var textBuilder strings.Builder
	var titleBuilder strings.Builder
	
	// Extract text and metadata
	p.extractContent(doc, &textBuilder, &titleBuilder, content)
	
	content.Text = strings.TrimSpace(textBuilder.String())
	content.Title = strings.TrimSpace(titleBuilder.String())
	
	if content.Title == "" {
		content.Title = "Untitled HTML Document"
	}
	
	return content, nil
}

/* extractContent recursively extracts content from HTML nodes */
func (p *HTMLParser) extractContent(n *html.Node, textBuilder, titleBuilder *strings.Builder, content *HTMLContent) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			textBuilder.WriteString(text)
			textBuilder.WriteString(" ")
		}
	} else if n.Type == html.ElementNode {
		switch n.Data {
		case "title":
			// Extract title
			if n.FirstChild != nil {
				titleBuilder.WriteString(n.FirstChild.Data)
			}
		case "a":
			// Extract links
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					content.Links = append(content.Links, attr.Val)
				}
			}
		case "meta":
			// Extract metadata
			if p.extractMetadata {
				var name, contentVal string
				for _, attr := range n.Attr {
					if attr.Key == "name" || attr.Key == "property" {
						name = attr.Val
					}
					if attr.Key == "content" {
						contentVal = attr.Val
					}
				}
				if name != "" && contentVal != "" {
					content.Metadata[name] = contentVal
				}
			}
		case "h1", "h2", "h3", "h4", "h5", "h6":
			// Extract headings as sections
			if n.FirstChild != nil {
				headingText := strings.TrimSpace(n.FirstChild.Data)
				if headingText != "" {
					level := int(n.Data[1] - '0') // Extract number from h1, h2, etc.
					content.Sections = append(content.Sections, SectionContent{
						Title:   headingText,
						Content: "",
						Level:   level,
					})
				}
			}
		}
	}
	
	// Recursively process child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.extractContent(c, textBuilder, titleBuilder, content)
	}
}
