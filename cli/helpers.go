package cli

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func wrapError(err error, funcTag, message string) error {
	format := "(%s) => [%s] %s"
	if err == nil {
		err = fmt.Errorf("Unspecified Error")
	}
	if len(message) == 0 {
		message = "-"
	}
	return fmt.Errorf(format, message, funcTag, err.Error())
}

// getEnvVarString returns the string input or the default if not set
func getEnvVarString(envKey, defaultValue string) string {
	envValue := os.Getenv(envKey)
	if len(envValue) == 0 {
		envValue = defaultValue
	}
	return envValue
}

// convertEnvStringToBool returns the default, unless the input parses as true
func convertEnvStringToBool(stringValue string, defaultValue bool) bool {
	// true === "true" or "1"
	if strings.EqualFold(stringValue, "true") || strings.EqualFold(stringValue, "1") {
		return true
	}
	// else, default
	return defaultValue
}

// convertBoolToEnvString returns a string that can be parsed accurately by `func convertEnvStringToBool(...)`
func convertBoolToEnvString(boolValue bool) string {
	// something that parses as true
	if boolValue {
		return "true"
	}
	// else, something that parses as false
	return "false"
}

// getEnvVarBool gets an env var and converts it to a bool
func getEnvVarBool(envKey string, defaultValue bool) bool {
	stringDefault := convertBoolToEnvString(defaultValue)
	stringValue := getEnvVarString(envKey, stringDefault)
	return convertEnvStringToBool(stringValue, defaultValue)
}

// getDefaultCaptureDevice returns the default capture device for the runtime platform
func getDefaultCaptureDevice() string {
	if strings.EqualFold(runtime.GOOS, "linux") {
		return "/dev/video0"
	}
	if strings.EqualFold(runtime.GOOS, "darwin") {
		return "default"
	}
	return "not specified"
}

func getSupportedCaptureFormats() (list []string) {
	list = append(list, "jpg")
	list = append(list, "png")
	return
}

func getDefaultCaptureFormat() string {
	return "jpg"
}

func isSupportedCaptureFormat(formatInput string) bool {
	for _, format := range getSupportedCaptureFormats() {
		if strings.EqualFold(format, formatInput) {
			return true
		}
	}
	return false
}
