package main

import (
	"fmt"
	"strings"
)

func deepCopyContent(content [][]rune) [][]rune {
	newContent := make([][]rune, len(content))
	for i, line := range content {
		newContent[i] = make([]rune, len(line))
		copy(newContent[i], line)
	}
	return newContent
}

func highlightSearch(text, searchTerm string) string {
	if searchTerm == "" {
		return text
	}

	highlightStyle := "\033[43m%s\033[0m" // Yellow background
	parts := strings.Split(text, searchTerm)
	for i := 0; i < len(parts)-1; i++ {
		parts[i] += fmt.Sprintf(highlightStyle, searchTerm)
	}
	return strings.Join(parts, "")
}

func expandTabs(s string, tabSize int) string {
	var result strings.Builder
	column := 0
	for _, r := range s {
		if r == '\t' {
			spaces := tabSize - (column % tabSize)
			result.WriteString(strings.Repeat(" ", spaces))
			column += spaces
		} else {
			result.WriteRune(r)
			column++
		}
	}
	return result.String()
}
