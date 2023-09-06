package prompttools

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-shiori/go-readability"
)

type StringExtractor func() (string, error)

func GenerateFinalPrompt() (string, error) {
	return CreateInitialPrompt("can you read docs from the wiki and help to fix my test").
		StartOfAdditionalInformationSection().
		AddFile("project/main.go").
		AddFile("main_test.go").
		AddUrl("https://detailedwiki.com").
		EndOfAdditionalInformationSection().
		GenerateFinalPrompt()
}

func CreateEmptyPrompt(initialPrompt string) *PromptBuilder {
	return CreateInitialPrompt("")
}

func CreateInitialPrompt(initialPrompt string) *PromptBuilder {
	return &PromptBuilder{
		prompt: initialPrompt,
	}
}

func (p *PromptBuilder) GenerateFinalPrompt() (string, error) {
	if p.err != nil {
		return "", p.err
	}
	return p.prompt, nil
}

func (p *PromptBuilder) AddTextToPrompt(text string) *PromptBuilder {
	if p.err != nil {
		return p
	}
	p.prompt = fmt.Sprintf("\n%s\n%s\n", p.prompt, text)
	return p
}

func (p *PromptBuilder) AddFile(filePath string) *PromptBuilder {
	if p.err != nil {
		return p
	}

	templateStr := `
{{ .OriginalPrompt }}
File path: {{ .Path }}
Content:
{{ .Content }}
`
	return p.addToPrompt(func() (string, error) {
		return readFileContent(filePath)
	}, templateStr, filePath)
}

func (p *PromptBuilder) AddFiles(filePath []string) *PromptBuilder {
	if len(filePath) == 0 {
		return p
	}
	return p.AddFile(filePath[0]).AddFiles(filePath[1:])
}

func (p *PromptBuilder) AddUrl(url string) *PromptBuilder {
	templateStr := `
{{ .OriginalPrompt }}
Url: {{ .Path }}
Content:
{{ .Content }}
`
	return p.addToPrompt(func() (string, error) {
		return extractReadableTextFromURL(url)
	}, templateStr, url)
}

func (p *PromptBuilder) AddSeparator() *PromptBuilder {
	if p.err != nil {
		return p
	}
	return p.AddTextToPrompt("========")
}

func (p *PromptBuilder) AddUrls(urls []string) *PromptBuilder {
	if len(urls) == 0 {
		return p
	}
	return p.AddUrl(urls[0]).AddUrls(urls[1:])
}

func (p *PromptBuilder) StartOfAdditionalInformationSection() *PromptBuilder {
	if p.err != nil {
		return p
	}

	p.prompt = fmt.Sprintf("%s\nAdditional Information Provided by user:\n", p.prompt)
	return p
}

func (p *PromptBuilder) EndOfAdditionalInformationSection() *PromptBuilder {
	if p.err != nil {
		return p
	}

	p.prompt = fmt.Sprintf("%s\nEnd of Additional Information Section\n", p.prompt)
	return p
}

func extractReadableTextFromURL(urlString string) (string, error) {
	resp, err := http.Get(urlString)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to download the page")
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}
	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return "", err
	}

	return article.TextContent, nil
}

func readFileContent(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
