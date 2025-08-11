package main

import (
	"log"
	"os"

	"github.com/teilomillet/gollm"
)

func NewAI() gollm.LLM {
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
	return llm
}
