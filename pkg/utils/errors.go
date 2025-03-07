package utils

import "fmt"

func FormatErrorMessage(action string, err error) error {
	return fmt.Errorf("%s: %v", action, err)
}
