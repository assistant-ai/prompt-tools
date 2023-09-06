package chainexecutor

import (
	"io/ioutil"
	"testing"

	"github.com/assistant-ai/llmchat-client/gpt"
)

func TestSimpleChainExecutor(t *testing.T) {
	client, err := gpt.NewGptClientFromFile("/Users/vkovalevskyi/.jess/open-ai.key", 0, gpt.ModelGPT4, "", 8000, nil)
	if err != nil {
		t.Fatalf("Unable to create client: %v", err)
	}
	// Read the content from the file
	data, err := ioutil.ReadFile("./transcript.txt")
	if err != nil {
		t.Fatalf("Unable to read file: %v", err)
	}
	podcastTranscript := string(data)

	executor := NewSimpleChainExecutor().
		LlmClient(client).
		Text(podcastTranscript).
		EnableDebug().
		TaskPromt("create podcast show notes from this that will be published to iTunes")

	// Execute
	result, err := executor.Execute()
	if err != nil {
		t.Fatalf("Error while executing: %v", err)
	}

	// Check the result
	if result == "" {
		t.Fatalf("Output is empty")
	}
}
