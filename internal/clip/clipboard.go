package clip

import (
	"errors"

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