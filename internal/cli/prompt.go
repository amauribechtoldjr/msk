package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/amauribechtoldjr/msk/internal/logger"
	"golang.org/x/term"
)

func PromptPassword(label string) ([]byte, error) {
	logger.PrintInfo(label)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	if err != nil {
		return nil, err
	}

	if len(bytePassword) == 0 {
		return nil, errors.New("Invalid master key.")
	}

	return bytePassword, nil
}