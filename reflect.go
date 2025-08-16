package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func latestChatLog() string {
	entries, err := os.ReadDir(kOutputDir)
	if err != nil {
		if *figs.Bool(kDebug) {
			_, _ = fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", kOutputDir, err)
		}
		return ""
	}

	var latestFile string
	var latestModTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if strings.HasPrefix(filename, "chatlog.") && strings.HasSuffix(filename, ".md") {
			fullPath := filepath.Join(kOutputDir, filename)

			fileInfo, err := entry.Info()
			if err != nil {
				if *figs.Bool(kDebug) {
					_, _ = fmt.Fprintf(os.Stderr, "Error getting file info for %s: %v\n", fullPath, err)
				}
				continue
			}

			modTime := fileInfo.ModTime()
			if latestFile == "" || modTime.After(latestModTime) {
				latestFile = fullPath
				latestModTime = modTime
			}
		}
	}

	if latestFile == "" && *figs.Bool(kDebug) {
		_, _ = fmt.Fprintf(os.Stderr, "No summary files found in directory %s\n", kOutputDir)
	}

	return latestFile
}
