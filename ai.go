package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/teilomillet/gollm"
)

func CanAI(figsK string) bool {
	if len(figsK) == 0 {
		return false
	}
	return *figs.Bool(kAiEnabled) && *figs.Bool(figsK)
}

func Agent() AI {
	capture("env FORCE_COLOR set", os.Setenv("FORCE_COLOR", "1"))
	capture("env TERM set", os.Setenv("TERM", "xterm-256color"))
	if aiPtr == nil {
		provider, model, seed := *figs.String(kAiProvider), *figs.String(kAiModel), *figs.Int(kAiSeed)
		maxTokens := *figs.Int(kAiMaxTokens)
		var opts []gollm.ConfigOption
		opts = append(opts, gollm.SetProvider(provider))
		opts = append(opts, gollm.SetModel(model))
		if seed > 1 {
			opts = append(opts, gollm.SetSeed(seed))
		}
		if maxTokens > 0 {
			opts = append(opts, gollm.SetMaxTokens(maxTokens))
		}
		switch provider {
		case "ollama":
			capture("unset OLLAMA_API_KEY env", os.Unsetenv("OLLAMA_API_KEY"))
			opts = append(opts, gollm.SetTemperature(0.99))
			opts = append(opts, gollm.SetLogLevel(gollm.LogLevelError))
		default:
			apiKey := *figs.String(kAiApiKey)
			opts = append(opts, gollm.SetAPIKey(apiKey))
		}
		llm, err := gollm.NewLLM(opts...)
		if err != nil {
			log.Fatal(err)
		}
		aiPtr = &agent{
			llm: llm,
		}
	}
	return aiPtr
}

func (a *agent) Ask(summary string, question ...string) (*Response, error) {
	var (
		r          = &Response{}
		response   = ""
		err        = errors.New("failed to ask agent for a response")
		directives = []string{
			"Be concise and offer complete solutions",
			"Act as Commander Data from the USS Starship Enterprise acting as an AI Agent assisting the user",
			"Refer to the user as Commander",
			"Speak as if you were on a Military Base as a member of the USS Starship Enterprise",
			"Speak as if you are on duty with fellow crew mates",
			"When replying to followup requests, build on your previous answer",
			"When a mistake is identified by the user, use the full previous response to modify and return",
			"Do not be afraid to offend and always give an honest answer in as few words as possible",
			"Do not format the output in markdown, use just plain text to STDOUT through a redirect or pipe, thus no formatting at all.",
			"Only reply in raw ASCII.",
		}

		inputContext = strings.Clone(summary)
		input        = strings.Join(question, "\n")

		prompt = gollm.NewPrompt(input,
			gollm.WithContext(inputContext),
			gollm.WithDirectives(directives...),
		)
	)
	response, err = a.llm.Generate(context.Background(), prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate the response: %w", err)
	}

	r.Response = strings.Clone(strings.TrimSpace(response))
	r.Request = strings.Clone(strings.TrimSpace(prompt.String()))
	r.Context = inputContext
	r.Directives = directives
	r.agent = Agent()
	r.prompt = prompt

	return r, nil
}
