package prompttools

import (
	"google.golang.org/api/docs/v1"
)

func (p *PromptBuilder) WithGoogleDocs(docsSrv *docs.Service) *PromptBuilder {
	if p.err != nil {
		return p
	}
	p.docsSrv = docsSrv
	return p
}

func (p *PromptBuilder) AddGoogleDocs(fileID string) *PromptBuilder {
	if p.err != nil {
		return p
	}
	googleDocsTemplateStr := `
{{ .OriginalPrompt }}
Google Doc id: {{ .Path }}
Content:
{{ .Content }}
`
	return p.addToPrompt(func() (string, error) {
		return getGoogleFileContent(p.docsSrv, fileID)
	}, googleDocsTemplateStr, fileID)
}

func getGoogleFileContent(docsSrv *docs.Service, fileID string) (string, error) {
	doc, err := docsSrv.Documents.Get(fileID).Do()
	if err != nil {
		return "", err
	}

	var content string
	for _, cs := range doc.Body.Content {
		if cs.Paragraph != nil {
			for _, pe := range cs.Paragraph.Elements {
				if pe.TextRun != nil {
					content += pe.TextRun.Content
				}
			}
		}
	}

	return content, nil
}
