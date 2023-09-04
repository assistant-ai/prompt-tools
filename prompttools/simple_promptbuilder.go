package prompttools

import (
	"html/template"
	"strings"

	"google.golang.org/api/docs/v1"
)

type PromptBuilder struct {
	prompt  string
	err     error
	docsSrv *docs.Service
}

func (p *PromptBuilder) addToPrompt(extractor StringExtractor, templateStr string, path string) *PromptBuilder {
	if p.err != nil {
		return p
	}

	content := ""
	err := error(nil)
	if content, err = extractor(); err != nil {
		p.prompt = ""
		p.err = err
		return p
	}
	var prompt strings.Builder
	tpl, err := template.New("prompt").Parse(templateStr)
	if err != nil {
		p.prompt = ""
		p.err = err
		return p
	}
	data := struct {
		OriginalPrompt string
		Path           string
		Content        string
	}{
		p.prompt,
		path,
		content,
	}
	if err = tpl.Execute(&prompt, data); err != nil {
		p.prompt = ""
		p.err = err
		return p
	}

	p.prompt = prompt.String()
	return p
}
