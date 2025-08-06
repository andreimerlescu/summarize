package main

import (
	"sync"

	"github.com/teilomillet/gollm"
)

type (
	AI interface {
		Ask(context string, question ...string) (*Response, error)
	}
	// agent represents a new ai llm interface
	agent struct {
		llm gollm.LLM
	}

	Response struct {
		agent      AI
		prompt     *gollm.Prompt
		Request    string   `json:"request"`
		Directives []string `json:"directives"`
		Context    string   `json:"context"`
		Response   string   `json:"response"`
	}

	// Result contains the scanned path in the kSourceDir that matched the conditions and shall be included in the Final summary of kFilename
	Result struct {
		Path     string `yaml:"path" json:"path"`
		Contents []byte `yaml:"contents" json:"contents"`
		Size     int64  `yaml:"size" json:"size"`
	}

	// Final contains the rendered Result of the matched path that gets written to kFilename
	Final struct {
		Path     string `yaml:"path" json:"path"`
		Contents string `yaml:"contents" json:"contents"`
		Size     int64  `yaml:"size" json:"size"`
	}

	// M defines a Message that should be rendered to JSON
	M struct {
		Message string `json:"message"`
	}

	// seenStrings captures a concurrent safe map of strings and booleans that indicate whether the string has been seen
	seenStrings struct {
		mu sync.RWMutex
		m  map[string]bool
	}
)
