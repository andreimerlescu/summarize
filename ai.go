package main

import (
	"fmt"
	"os"
	"time"

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
	opts = append(opts, gollm.SetMemory(*figs.Int(kMemory)))
	opts = append(opts, gollm.SetEnableCaching(*figs.Bool(kAiCachingEnabled)))
	timeout := *figs.UnitDuration(kAiTimeout)
	if timeout < time.Second {
		timeout = dTimeout * dTimeoutUnit
	}
	opts = append(opts, gollm.SetTimeout(*figs.UnitDuration(kAiTimeout)))
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
		fmt.Printf("❌ Failed to initialize AI: %v\n", err)
		return nil
	}
	return llm
}
