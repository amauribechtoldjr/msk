package cli

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/validator"
	"golang.org/x/term"
)

var ErrInvalidValue = errors.New("Invalid master key.")
var ErrConfirmationMatch = errors.New("Invalid master key confirmation.")

func PromptSafeValue(label string) ([]byte, error) {
	logger.PrintInfo(label)
	safeValue, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	if err != nil {
		return nil, err
	}

	if len(safeValue) == 0 {
		return nil, ErrInvalidValue
	}

	return safeValue, nil
}

func PromptMasterPassword(shouldConfirm bool) ([]byte, error) {
	pass, err := PromptSafeValue("Enter master password:")
	if err != nil {
		return nil, err
	}

	if err := validator.ValidateMasterPass(string(pass)); err != nil {
		return nil, err
	}

	if shouldConfirm {
		passConfirmation, err := PromptSafeValue("Enter master password again to confirm operation:")
		if err != nil {
			return nil, err
		}

		if !reflect.DeepEqual(pass, passConfirmation) {
			return nil, ErrConfirmationMatch
		}
	}

	return pass, nil
}
