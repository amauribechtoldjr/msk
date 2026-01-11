package logger

import (
	"errors"
	"os"

	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/fatih/color"
)

const defaultErrorMessage = "ERROR"

func RenderError(err error) {
	switch {
		case errors.Is(err, app.ErrSecretNotFound):
			printError("%s: Secret not found\n", defaultErrorMessage)
		case errors.Is(err, app.ErrSecretExists):
			printError("%s: Secret already exists\n", defaultErrorMessage)
		default:
			printError("%s: %s\n", defaultErrorMessage, err.Error())
	}
}

func printError(format string, a ...any) {
	color.New(color.FgRed).Fprintf(os.Stderr, format, a...)
}

// func PrintfInfo(format string, a ...any) {
// 	color.New(color.BgHiBlue).Printf(format, a...)
// }

func PrintInfo(message string) {
	color.New(color.BgHiBlue).Print(message)
}

// func PrintfSuccess(format string, a ...any) {
// 	color.New(color.FgGreen).Printf(format, a...)
// }

func PrintSuccess(message string) {
	color.New(color.FgGreen).Print(message)
}