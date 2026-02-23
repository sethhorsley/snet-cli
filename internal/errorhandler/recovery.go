package errorhandler

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

// RecoverPanic handles panics gracefully and ensures terminal cleanup
func RecoverPanic() {
	if r := recover(); r != nil {
		// Ensure terminal is restored to normal state
		// Send these multiple times to ensure they're received
		for i := 0; i < 3; i++ {
			fmt.Print("\033[?25h")   // Show cursor
			fmt.Print("\033[0m")     // Reset colors
			fmt.Print("\033[?1049l") // Exit alt screen
		}
		fmt.Print("\n")

		// Print error message (simple, no fancy formatting)
		fmt.Fprintln(os.Stderr, "\033[31m✗ Fatal Error\033[0m")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintf(os.Stderr, "The CLI encountered an unexpected error: %v\n", r)
		fmt.Fprintln(os.Stderr, "")

		// Print stack trace in verbose mode or if SNET_DEBUG is set
		if os.Getenv("SNET_DEBUG") == "1" {
			fmt.Fprintln(os.Stderr, "\033[90mStack trace:\033[0m")
			fmt.Fprintln(os.Stderr, string(debug.Stack()))
		} else {
			fmt.Fprintln(os.Stderr, "\033[90mRun with SNET_DEBUG=1 for detailed stack trace\033[0m")
		}

		os.Exit(1)
	}
}

// FormatConnectionError formats FRP connection errors nicely
func FormatConnectionError(err error) string {
	errMsg := err.Error()

	// Common error patterns
	if containsAny(errMsg, "connection reset", "connection refused") {
		return `Connection Error

The FRP server is unreachable. This could mean:
  • Your tunnel credentials are invalid
  • The tunnel has been deleted
  • The server configuration has changed

Try these steps:
  1. Check your internet connection
  2. Verify the FRP server is running
  3. Contact support if the issue persists
`
	}

	if containsAny(errMsg, "login to the server failed") {
		return `Authentication Error

Authentication with the FRP server failed. This could mean:
  • Your tunnel credentials are invalid
  • The tunnel has been deleted
  • The server configuration has changed

Try these steps:
  1. Create a new tunnel: snet http 3000 --name new-tunnel
  2. Check your account status online
  3. Contact support if the issue persists
`
	}

	// Generic error
	return fmt.Sprintf("Error\n\n%s", errMsg)
}

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
