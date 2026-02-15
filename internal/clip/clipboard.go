package clip

import (
	"errors"
	"fmt"
	"time"

	"github.com/amauribechtoldjr/msk/internal/logger"
	"golang.design/x/clipboard"
)

var (
	ErrClipboardInit = errors.New("failed to initialize clipboard")
)

func Init() error {
	err := clipboard.Init()
	if err != nil {
		return ErrClipboardInit
	}
	return nil
}

func CopyText(text []byte) error {
	_ = clipboard.Write(clipboard.FmtText, text)
	return nil
}

func Clear() {
	timer := 15
	logger.PrintSuccessf("Password will be cleared from clipboard in %v seconds: ", timer)

	for timer > 0 {
		logger.PrintSuccess(".")
		time.Sleep(1 * time.Second)
		timer -= 1
	}
	CopyText([]byte{})

	fmt.Println()
	logger.PrintSuccess("Clipboard cleared.\n")
}
