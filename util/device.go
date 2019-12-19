package util

import (
	"runtime"
	"strings"
)

// DefaultCaptureDevice returns the default capture device for the runtime platform
func DefaultCaptureDevice() string {
	if strings.EqualFold(runtime.GOOS, "linux") {
		return "/dev/video0"
	}
	if strings.EqualFold(runtime.GOOS, "darwin") {
		return "default"
	}
	return "not specified"
}
