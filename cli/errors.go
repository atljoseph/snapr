package cli

import "fmt"

func wrapError(err error, funcTag, message string) error {
	fs := "(%s) => [%s] %s"
	if err == nil {
		err = fmt.Errorf("Unspecified Error")
	}
	if len(message) == 0 {
		message = "-"
	}
	return fmt.Errorf(fs, funcTag, message, err.Error())
}
