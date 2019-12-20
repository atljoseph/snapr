package util

import (
	"bytes"
	"os/exec"
	"strings"
)

// OSUsers gets the list of logged in users for the current device
// if "you" and "i", and "love" are logged in, then this will return "i love you"
func OSUsers() (usersOutput string, err error) {
	funcTag := "OSUsers"

	// build the command and
	// get the logged in users list
	usersExec := exec.Command("/bin/sh", "-c", "users")
	usersBytes, err := usersExec.Output()
	if err != nil {
		err = WrapError(err, funcTag, "exec os cmd to get list of logged in users")
		return
	}

	// get the output
	usersOutput = bytes.NewBuffer(usersBytes).String()
	// remove line breaks
	usersOutput = strings.ReplaceAll(usersOutput, "\n", "")
	// replace spaces ith "-"
	usersOutput = strings.ReplaceAll(usersOutput, " ", "-")

	// done
	return
}
