package util

import (
	"os"
	"strconv"
	"strings"
)

// EnvVarString returns the string input or the default if not set
func EnvVarString(envKey, defaultValue string) string {
	envValue := os.Getenv(envKey)
	if len(envValue) == 0 {
		envValue = defaultValue
	}
	return envValue
}

// EnvVarStringSlice returns the string input or the default if not set
func EnvVarStringSlice(envKey, defaultValue string) []string {
	strValue := EnvVarString(envKey, defaultValue)
	strValue = strings.ReplaceAll(strValue, " ", "")
	return strings.Split(strValue, ",")
}

// EnvVarInt returns the input as an int, or the default if not set
func EnvVarInt(envKey string, defaultValue int) int {
	// start by defaulting to the default
	intValue := defaultValue
	// get the env string value
	strValue := os.Getenv(envKey)
	// if set
	if len(strValue) == 0 {
		// attempt to convert to int
		convValue, err := strconv.Atoi(strValue)
		// if error, default to 0 (if not already done)
		if err != nil {
			convValue = 0
		}
		// assign the int value
		intValue = convValue
	}
	// return the int value
	return intValue
}

// EnvStringToBool returns the default, unless the input parses as true
func EnvStringToBool(stringValue string, defaultValue bool) bool {
	// true === "true" or "1"
	if strings.EqualFold(stringValue, "true") || strings.EqualFold(stringValue, "1") {
		return true
	}
	// else, default
	return defaultValue
}

// BoolToEnvString returns a string that can be parsed accurately by `func convertEnvStringToBool(...)`
func BoolToEnvString(boolValue bool) string {
	// something that parses as true
	if boolValue {
		return "true"
	}
	// else, something that parses as false
	return "false"
}

// EnvVarBool gets an env var and converts it to a bool
func EnvVarBool(envKey string, defaultValue bool) bool {
	stringDefault := BoolToEnvString(defaultValue)
	stringValue := EnvVarString(envKey, stringDefault)
	return EnvStringToBool(stringValue, defaultValue)
}
