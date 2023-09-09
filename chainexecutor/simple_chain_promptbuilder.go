package chainexecutor

import (
	"errors"
	"fmt"

	"github.com/assistant-ai/llmchat-client/client"
	"github.com/assistant-ai/prompt-tools/prompttools"
)

var MAP_PROMPT = `User has provided a large text and requested you to process it in steps. The current chunk of text is part of a larger body of text. Extract a TL;DR summary containing the most important information needed to achieve the user's final goal.

Your output must:
* Be no longer than 3000 words
* Include all crucial details learned from the user's input that may be necessary for achieving the goal

Do not include any explanatory text. Your output should strictly contain the learned information, which will later be used to accomplish the final task.`

var REDUCE_PROMPT = `You are operating within the framework of the Map/Reduce algorithm. You are currently in the reduce step.

You are given two text snippets and a user task. Each snippet contains important information that may be needed to achieve the user's goal. Your current task is to generate a new text snippet that:
* Includes all the important information from both snippets
* Is no longer than 3000 words

Remember, you do not have to complete the user's final task at this point. Your goal is to merge the two snippets so that the final task can be achieved later. Provide the output as a text snippet without any comments or explanations.`

var FINAL_PROMPT = `The user has provided a large text that you have already analyzed and reduced to a reasonably small snippet. This snippet should contain all the important information needed to achieve the task. Below, you will find the text and the user's task. Your output SHOULD be the result of completing the user's task.`

type Provider interface {
	Provide() (string, error)
}

type SimpleChainExecutor struct {
	taskPrompt string
	err        error
	text       string
	llmClient  *client.Client
	debug      bool
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

func (p *SimpleChainExecutor) EnableDebug() *SimpleChainExecutor {
	if p.err != nil {
		return p
	}
	p.debug = true
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
	chunks := splitStringIntoChunksOfSizeWithOverlap(p.text, 6000*3)

	resultChan := make(chan string)
	errChan := make(chan error, 1)
	go p.executeMapReduce(chunks, 0, len(chunks)-1, resultChan, errChan)

	var result string
	var err error

	select {
	case result = <-resultChan:
	case err = <-errChan:
	}

	if err != nil {
		return "", err
	}

	finalReducedText, err := p.finalReduce(result)
	if err != nil {
		return "", err
	}
	p.debugShowTextAndWaitKeybordEnter("final reduced text: " + finalReducedText + "\n" + "final memory: " + result)
	return finalReducedText, nil
}

func (p *SimpleChainExecutor) executeMapReduce(chunks []string, start int, end int, resultChan chan string, errChan chan error) {
	if start == end {
		mappedChunk, err := p.mapChunk(chunks[start], start, len(chunks))
		if err != nil {
			errChan <- err
			return
		}
		p.debugShowTextAndWaitKeybordEnter("original text: " + chunks[start] + "\nMapped: " + mappedChunk)
		resultChan <- mappedChunk
		return
	}

	middle := (start + end) / 2
	leftChan := make(chan string)
	rightChan := make(chan string)
	leftErrChan := make(chan error, 1)
	rightErrChan := make(chan error, 1)

	go p.executeMapReduce(chunks, start, middle, leftChan, leftErrChan)
	go p.executeMapReduce(chunks, middle+1, end, rightChan, rightErrChan)

	var leftResult string
	var leftErr error

	select {
	case leftResult = <-leftChan:
	case leftErr = <-leftErrChan:
	}

	if leftErr != nil {
		errChan <- leftErr
		return
	}

	var rightResult string
	var rightErr error

	select {
	case rightResult = <-rightChan:
	case rightErr = <-rightErrChan:
	}

	if rightErr != nil {
		errChan <- rightErr
		return
	}

	reducedText, err := p.reduce(leftResult, rightResult)
	if err != nil {
		errChan <- err
		return
	}
	p.debugShowTextAndWaitKeybordEnter("left: " + leftResult + "\nright: " + rightResult + "\nreduced: " + reducedText)
	resultChan <- reducedText
}

func (p *SimpleChainExecutor) mapChunk(chunk string, chunkNumber int, totalNumbersOfChunks int) (string, error) {
	prompt, err := prompttools.CreateInitialPrompt(MAP_PROMPT).
		AddTextToPrompt("\nUser task: "+p.taskPrompt+"\n").
		StartOfAdditionalInformationSection().
		AddTextToPromptf("Chunk of user text number %d, ouf of %d:", chunkNumber, totalNumbersOfChunks).
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
		AddTextToPrompt("left text chunk to reduce: " + left).
		AddTextToPrompt("========").
		AddTextToPrompt("right text chunk to reduce: " + right).
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

func splitStringIntoChunksOfSizeWithOverlap(text string, size int) []string {
	var chunks []string
	overlap := 300
	for i := 0; i < len(text); i += size {
		end := i + size
		if end > len(text) {
			end = len(text)
		}
		startIndex := i
		if startIndex-overlap > 0 {
			startIndex -= overlap
		}
		endIndex := end
		if endIndex+overlap < len(text) {
			endIndex += overlap
		}
		chunks = append(chunks, text[startIndex:endIndex])
	}
	return chunks
}

func (p *SimpleChainExecutor) debugShowTextAndWaitKeybordEnter(text string) {
	if p.debug {
		println("=====================================")
		println(text)
		println("Press Enter to continue...")
		var input string
		fmt.Scanln(&input)
	}
}
