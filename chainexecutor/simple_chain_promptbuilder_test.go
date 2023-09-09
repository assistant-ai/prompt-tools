package chainexecutor

import (
	"fmt"
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
		TaskPromt("Generate podcast show notes from the provided transcript. The notes should be concise yet factual, summarizing key points discussed in the episode. These show notes will be published on iTunes. Avoid excessive use of buzzwords; the aim is to inform listeners rather than attract attention. For formatting and style, refer to the show notes of the podcast Radio-T as a guideline. Do not explicitly mention that the style is similar to Radio-T. Additionally, here's the Discord link for further discussion: https://discord.gg/T38WpgkHGQ. Please format the output in Markdown. Also in the beggining of your output list 3-4 suggested podcast title names.")

	// Execute
	result, err := executor.Execute()
	if err != nil {
		t.Fatalf("Error while executing: %v", err)
	}

	// Check the result
	if result == "" {
		t.Fatalf("Output is empty")
	}
	fmt.Println(result)
}
