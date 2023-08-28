package prompttools

import (
	"testing"
)

func TestCreateEmptyPrompt(t *testing.T) {
	emptyPrompt, _ := CreateEmptyPrompt("").GenerateFinalPrompt()
	if emptyPrompt != "" {
		t.Errorf("CreateEmptyPrompt failed, expected %v, got %v", "", emptyPrompt)
	}
}

func TestCreateInitialPrompt(t *testing.T) {
	initialPrompt := "Initial Prompt"
	emptyPrompt, _ := CreateInitialPrompt(initialPrompt).GenerateFinalPrompt()
	if emptyPrompt != initialPrompt {
		t.Errorf("CreateInitialPrompt failed, expected %v, got %v", initialPrompt, emptyPrompt)
	}
}

func TestAddFile(t *testing.T) {
	filePath := "textfile.txt"
	prompt, _ := CreateEmptyPrompt("").AddFile(filePath).GenerateFinalPrompt()

	// Here you should insert your expected result for the file prompt
	expectedResult := `

File path: textfile.txt
Content:
test123
`

	if prompt != expectedResult {
		t.Errorf("AddFile failed, expected \"%v\", got \"%v\"", expectedResult, prompt)
	}
}
