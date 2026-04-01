package utils

// ANSI escape codes for terminal color output.
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	ColorCyan   = "\033[36m"
	ColorGreen  = "\033[32m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"   // Claude brand blue
	ColorOrange = "\033[38;5;208m"
)

// Dim wraps text with dim ANSI styling.
func Dim(s string) string { return ColorDim + s + ColorReset }

// Bold wraps text with bold ANSI styling.
func Bold(s string) string { return ColorBold + s + ColorReset }

// Green wraps text with green ANSI color.
func Green(s string) string { return ColorGreen + s + ColorReset }

// Red wraps text with red ANSI color.
func Red(s string) string { return ColorRed + s + ColorReset }
