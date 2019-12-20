package util

import (
	"fmt"
)

// WrapError returns another error wrapping the original one
func WrapError(err error, funcTag, message string) error {
	format := "(%s) => [%s] %s"
	if err == nil {
		err = fmt.Errorf("Unspecified Error")
	}
	if len(message) == 0 {
		message = "-"
	}
	return fmt.Errorf(format, err.Error(), funcTag, message)
}
