package logger

import (
	"os"

	"github.com/fatih/color"
)

func PrintInfo(message string) {
	color.New(color.FgHiWhite).Fprint(os.Stderr, message)
}

func PrintSuccess(message string) {
	color.New(color.FgGreen).Fprint(os.Stderr, message)
}

func PrintSuccessf(format string, a ...any) {
	color.New(color.FgGreen).Fprintf(os.Stderr, format, a...)
}

func PrintError(format string, a ...any) {
	color.New(color.FgRed).Fprintf(os.Stderr, format, a...)
}
