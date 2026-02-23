package main

import (
	"fmt"

	"github.com/seth4242/snet/cmd"
	"github.com/seth4242/snet/internal/errorhandler"
)

func main() {
	// Ensure cursor is always restored on exit (safety net)
	defer func() {
		fmt.Print("\033[?25h") // Show cursor
		fmt.Print("\033[0m")   // Reset colors
	}()

	// Recover from panics and ensure terminal cleanup
	defer errorhandler.RecoverPanic()

	cmd.Execute()
}
