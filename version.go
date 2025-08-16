package main

import (
	"embed"
	"strings"
)

//go:embed VERSION
var versionBytes embed.FS

var currentVersion string

func Version() string {
	if len(currentVersion) == 0 {
		versionBytes, err := versionBytes.ReadFile("VERSION")
		if err != nil {
			return ""
		}
		currentVersion = strings.TrimSpace(string(versionBytes))
	}
	return currentVersion
}
