package validator

import (
	"fmt"
	"regexp"
)

func Validate(pass string) error {
	fmt.Print(pass)
	onlyLettersAndNumbers, err := regexp.Compile(`^[A-Za-z0-9]+$`)
	if err != nil {
		return err
	}
	fmt.Print(onlyLettersAndNumbers.MatchString(pass))

	return nil
}
