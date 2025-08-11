package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func latestSummaryFile() string {
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
		// Skip directories
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		// Check if file matches pattern: starts with "summary." and ends with ".md"
		if strings.HasPrefix(filename, "summary.") && strings.HasSuffix(filename, ".md") {
			fullPath := filepath.Join(kOutputDir, filename)

			fileInfo, err := entry.Info()
			if err != nil {
				if *figs.Bool(kDebug) {
					_, _ = fmt.Fprintf(os.Stderr, "Error getting file info for %s: %v\n", fullPath, err)
				}
				continue
			}

			modTime := fileInfo.ModTime()
			// If this is the first matching file or it's newer than the current latest
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
