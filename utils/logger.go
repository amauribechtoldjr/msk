package utils

import (
	"github.com/fatih/color"
)

var PrintError = color.New(color.FgRed).PrintlnFunc()
var PrintInfo = color.New(color.FgYellow).PrintlnFunc()
var PrintSuccess = color.New(color.FgGreen).PrintlnFunc()

func Panic(e error) {
	if e != nil {
		panic(e)
	}
}

func CheckError(e error) {
	if e != nil {
		PrintError(e)
	}
}

func InfoMessage(e string) {
	PrintInfo(e)
}

func SuccessMessage(m string) {
	PrintSuccess(m)
}

func ErrorMessage(m string) {
	PrintError(m)
}