package main

import "sync"

type (

	// result contains the scanned path in the kSourceDir that matched the conditions and shall be included in the final summary of kFilename
	result struct {
		Path     string `yaml:"path" json:"path"`
		Contents []byte `yaml:"contents" json:"contents"`
		Size     int64  `yaml:"size" json:"size"`
	}

	// final contains the rendered result of the matched path that gets written to kFilename
	final struct {
		Path     string `yaml:"path" json:"path"`
		Contents string `yaml:"contents" json:"contents"`
		Size     int64  `yaml:"size" json:"size"`
	}

	// m defines a Message that should be rendered to JSON
	m struct {
		Message string `json:"message"`
	}

	// seenStrings captures a concurrent safe map of strings and booleans that indicate whether the string has been seen
	seenStrings struct {
		mu sync.RWMutex
		m  map[string]bool
	}
)
