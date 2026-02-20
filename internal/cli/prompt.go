package cli

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/validator"
	"github.com/amauribechtoldjr/msk/internal/wipe"
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
		wipe.Bytes(pass)
		return nil, err
	}

	if err := validator.ValidateMasterPass(pass); err != nil {
		wipe.Bytes(pass)
		return nil, err
	}

	if shouldConfirm {
		passConfirmation, err := PromptSafeValue("Enter master password again to confirm operation:")
		if err != nil {
			wipe.Bytes(pass)
			return nil, err
		}

		if !reflect.DeepEqual(pass, passConfirmation) {
			wipe.Bytes(pass)
			wipe.Bytes(passConfirmation)
			return nil, ErrConfirmationMatch
		}

		wipe.Bytes(passConfirmation)
	}

	return pass, nil
}
