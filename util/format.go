package util

import (
	"strings"
)

// SupportedCaptureFormats returns a slice of supported image capture formats
func SupportedCaptureFormats() []string {
	var list []string
	list = append(list, "jpg")
	list = append(list, "png")
	return list
}

// DefaultCaptureFormat returns the default capture format
func DefaultCaptureFormat() string {
	return "jpg"
}

// IsSupportedCaptureFormat returns true of the format input is supported
func IsSupportedCaptureFormat(formatInput string) bool {
	for _, format := range SupportedCaptureFormats() {
		if strings.EqualFold(format, formatInput) {
			return true
		}
	}
	return false
}
