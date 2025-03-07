package utils

import "fmt"

func FormatErrorMessage(action string, err error) string {
	return fmt.Sprintf("%s: %v", action, err)
}
