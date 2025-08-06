package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// simplify takes a list of strings and reduces duplicates from the slice
func simplify(t []string) []string {
	seen := make(map[string]bool)
	for _, v := range t {
		seen[v] = true
	}
	results := make([]string, len(t))
	for i, v := range t {
		if seen[v] {
			results[i] = v
		}
	}
	return results
}

// StringPrompt asks for a string value using the label and returns the trimmed input.
func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		_, _ = fmt.Fprint(os.Stderr, label+" \n")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func TermWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 88
	}
	return width
}

func TermHeight() int {
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		height = 88
	}
	return height
}

func IsSafeWord(input string) bool {
	input = strings.ToLower(strings.TrimSpace(input))
	for _, w := range safeWords {
		if strings.EqualFold(w, input) {
			return true
		}
	}
	return false
}
