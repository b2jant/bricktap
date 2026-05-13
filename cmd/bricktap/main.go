package main

import (
	"fmt"
	"os"

	"github.com/b2jant/bricktap/internal/tui"
)

func main() {
	// Initialize and run the TUI
	if err := tui.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
