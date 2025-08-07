package ui

import (
	"os"

	"golang.org/x/term"
)

func getTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to reasonable defaults if we can't get terminal size
		return 80, 24
	}
	return width, height
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}