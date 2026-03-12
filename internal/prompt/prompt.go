package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/amauribechtoldjr/msk/internal/logger"
	"github.com/amauribechtoldjr/msk/internal/validator"
	"github.com/amauribechtoldjr/msk/internal/wipe"
	"golang.org/x/term"
)

var ErrEmptyInput = errors.New("input cannot be empty")
var ErrConfirmationMatch = errors.New("invalid master key confirmation")

func ReadString(label string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	logger.PrintInfo(label)
	str, err := reader.ReadString('\n')
	if err != nil {
		logger.Lb()
		return "", err
	}

	return string(str), nil
}

func ReadBoolean(label string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	logger.PrintInfo(label)

	str, err := reader.ReadByte()
	if err != nil {
		logger.Lb()
		return false, err
	}

	answer := strings.TrimSpace(strings.ToLower(string(str)))

	return answer == "y", nil
}

func ReadSafeValue(label string) ([]byte, error) {
	logger.PrintInfo(label)
	safeValue, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)

	if err != nil {
		return nil, err
	}

	if len(safeValue) == 0 {
		return nil, ErrEmptyInput
	}

	return safeValue, nil
}

func ReadMasterPassword(shouldConfirm bool) ([]byte, error) {
	pass, err := ReadSafeValue("Enter master password:")
	if err != nil {
		wipe.Bytes(pass)
		return nil, err
	}

	if err := validator.ValidateMasterPass(pass); err != nil {
		wipe.Bytes(pass)
		return nil, err
	}

	if shouldConfirm {
		passConfirmation, err := ReadSafeValue("Enter master password again to confirm operation:")
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
