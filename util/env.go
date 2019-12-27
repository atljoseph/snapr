package util

import (
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
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
func EnvVarStringSlice(envKey string, defaultValue []string) []string {
	// get the env value as string
	// including default if applicable
	strValue := EnvVarString(envKey, strings.Join(defaultValue, ","))

	// trim all spaces
	strValue = strings.ReplaceAll(strValue, " ", "")

	// split into string slice
	result := strings.Split(strValue, ",")

	// formats default (weird thing with cobra input slice)
	if len(result) == 1 {
		if len(strings.Trim(result[0], " ")) == 0 {
			result = nil
		}
	}
	return result
}

// EnvVarIntSlice returns an int slice of a string env variable, or its default
func EnvVarIntSlice(envKey string, defaultValue []int) []int {

	// convert the default to parsable string
	strDefault := ""
	for idx, i := range defaultValue {
		strDefault += strconv.Itoa(i)
		if idx < len(defaultValue) {
			strDefault += ","
		}
	}

	// get the env value as string
	// including default if applicable
	strValue := EnvVarString(envKey, strDefault)

	// trim all spaces
	strValue = strings.ReplaceAll(strValue, " ", "")

	// split into string slice
	strSlice := strings.Split(strValue, ",")

	// // formats default (weird thing with cobra input slice)
	// if len(strSlice) == 1 {
	// 	if len(strings.Trim(strSlice[0], " ")) == 0 {
	// 		strSlice = nil
	// 	}
	// }

	// convert output
	var result []int
	var err error
	for _, s := range strSlice {

		// convert to int
		convInt, err := strconv.Atoi(s)
		if err != nil {
			// warn and break with the error
			logrus.Warnf("Found unparsable input in int slice: %s", strSlice)
			continue
		}

		// append the result
		result = append(result, convInt)
	}

	// revert to the default value if conversion error
	if err != nil {
		result = defaultValue
	}

	return result
}

// EnvVarInt returns the input as an int, or the default if not set
func EnvVarInt(envKey string, defaultValue int) int {
	// start by defaulting to the default
	intValue := defaultValue
	// get the env string value
	strValue := os.Getenv(envKey)
	// if set
	if len(strValue) > 0 {
		// attempt to convert to int
		convValue, err := strconv.Atoi(strValue)
		// if error, default to 0 (if not already done)
		if err != nil {
			logrus.Warnf(err.Error())
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
