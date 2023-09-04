package chainexecutor

import (
	"errors"
	"fmt"

	"github.com/assistant-ai/llmchat-client/client"
	"github.com/assistant-ai/prompt-tools/prompttools"
)

var DEBUG = true

var MAP_PROMPT = `User asked you to do the task with the big text so you have do it in steps, since text is too big.
Current chanks of text is part of the bigger text, you have to extract TLDR of this text with the most importan information that you might need to acihve the final goal that users wants you to do.

Your output should strictly be:
* no longer than 3000 words
* include all the important details that you learned from the user's input that might be needed to achive the goal

Do not add any explanation text, or anything, you output should be strictly text with learning, later it will be give back to you as as to achive final task (but not now).`

var REDUCE_PROMPT = `You are working in scope of the Map/Reduce algorithm. Currently there is a reduce step in progerss that you have tod.

Below you given two texts snippets and user task. Each snippet includes important information that might be needed to achive user goal. You crrent task is to generate new text snippet that:
* includes all the important information from both snippets
* no longer than 3000 words

Keep in mind that you do not have to actually do user's final task yet, your goal is just to merge two snippets so lataer on final task can be achived. Please provide output as text snippet, no coments or explanation should be added.`

var FINAL_PROMPT = `User gave you a big text that you already have analyzed and reduced to a resonably small snippet of text. It should include all the important information you need to achive the task. Below you will find text and user task. Your output SHOULD be result of user task.`

type Provider interface {
	Provide() (string, error)
}

type SimpleChainExecutor struct {
	taskPrompt string
	err        error
	text       string
	llmClient  *client.Client
}

func NewSimpleChainExecutor() *SimpleChainExecutor {
	return &SimpleChainExecutor{}
}

func (p *SimpleChainExecutor) Text(text string) *SimpleChainExecutor {
	if p.err != nil {
		return p
	}
	p.text = text
	return p
}

func (p *SimpleChainExecutor) LlmClient(llmClient *client.Client) *SimpleChainExecutor {
	if p.err != nil {
		return p
	}
	p.llmClient = llmClient
	return p
}

func (p *SimpleChainExecutor) TaskPromt(taskPrompt string) *SimpleChainExecutor {
	if p.err != nil {
		return p
	}
	p.taskPrompt = taskPrompt
	return p
}

func (p *SimpleChainExecutor) Execute() (string, error) {
	if p.err != nil {
		return "", p.err
	}
	if p.text == "" {
		return "", errors.New("text is empty")
	}
	chunks := splitStringIntoChunksOfSize(p.text, 6000)
	memory, err := p.executeMapReduce(chunks, 0, len(chunks)-1)
	if err != nil {
		return "", err
	}
	finalReducedText, err := p.finalReduce(memory)
	if err != nil {
		return "", err
	}
	debugShowTextAndWaitKeybordEnter("final reduced text: " + finalReducedText + "\n" + "final memory: " + memory)
	return finalReducedText, nil
}

func (p *SimpleChainExecutor) executeMapReduce(chunks []string, start int, end int) (string, error) {
	if start == end {
		mappedChunk, err := p.mapChunk(chunks[start])
		if err != nil {
			return "", err
		}
		debugShowTextAndWaitKeybordEnter("original text: " + chunks[start] + "\nMapped: " + mappedChunk)
		return mappedChunk, nil
	}
	middle := (start + end) / 2
	left, err := p.executeMapReduce(chunks, start, middle)
	if err != nil {
		return "", err
	}
	right, err := p.executeMapReduce(chunks, middle+1, end)
	if err != nil {
		return "", err
	}
	reducedText, err := p.reduce(left, right)
	if err != nil {
		return "", err
	}
	debugShowTextAndWaitKeybordEnter("left: " + left + "\nright: " + right + "\nreduced: " + reducedText)
	return reducedText, nil
}

func (p *SimpleChainExecutor) mapChunk(chunk string) (string, error) {
	prompt, err := prompttools.CreateInitialPrompt(MAP_PROMPT).
		AddTextToPrompt("\nUser task: " + p.taskPrompt + "\n").
		StartOfAdditionalInformationSection().
		AddTextToPrompt(chunk).
		EndOfAdditionalInformationSection().
		GenerateFinalPrompt()
	if err != nil {
		return "", err
	}
	return p.llmClient.SendNoContextMessage(prompt)
}

func (p *SimpleChainExecutor) reduce(left string, right string) (string, error) {
	prompt, err := prompttools.CreateInitialPrompt(REDUCE_PROMPT).
		AddTextToPrompt("\nUser task: " + p.taskPrompt + "\n").
		StartOfAdditionalInformationSection().
		AddTextToPrompt("left text to reduce: " + left).
		AddTextToPrompt("========").
		AddTextToPrompt("right text to reduce: " + right).
		EndOfAdditionalInformationSection().
		GenerateFinalPrompt()
	if err != nil {
		return "", err
	}
	return p.llmClient.SendNoContextMessage(prompt)
}

func (p *SimpleChainExecutor) finalReduce(memory string) (string, error) {
	prompt, err := prompttools.CreateInitialPrompt(FINAL_PROMPT).
		AddTextToPrompt("\nUser task: " + p.taskPrompt + "\n").
		StartOfAdditionalInformationSection().
		AddTextToPrompt("last memory to use to achive the task: " + memory).
		EndOfAdditionalInformationSection().
		GenerateFinalPrompt()
	if err != nil {
		return "", err
	}
	return p.llmClient.SendNoContextMessage(prompt)
}

func splitStringIntoChunksOfSize(text string, size int) []string {
	var chunks []string
	for i := 0; i < len(text); i += size {
		end := i + size
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}
	return chunks
}

func debugShowTextAndWaitKeybordEnter(text string) {
	if DEBUG {
		println("=====================================")
		println(text)
		println("Press Enter to continue...")
		var input string
		fmt.Scanln(&input)
	}
}
