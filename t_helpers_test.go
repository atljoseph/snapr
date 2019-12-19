package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func wrapSnapTestError(test snapTest, errMsg string) string {
	return wrapTestError(test.description, test.cmdOpts, errMsg)
}

func wrapUploadTestError(test uploadTest, errMsg string) string {
	return wrapTestError(test.description, test.cmdOpts, errMsg)
}

func wrapTestError(description string, cmdOpts interface{}, errMsg string) string {
	msg := fmt.Sprintf("(%s) [%+v] %s", description, cmdOpts, errMsg)
	logrus.Warnf(msg)
	return msg
}

func cleanupTestDir(testDir string) (err error) {
	// remove test dir
	logrus.Infof("Cleaning up")
	err = os.RemoveAll(testDir)
	if err != nil {
		err = fmt.Errorf("could not remove test dir: %s", err)
		return
	}
	logrus.Infof("Test Dir removed")
	return
}

// ensureTestDir ensures that a test directory exists
// relativeDirName is relative the directory in which the go file exists
func ensureTestDir(relativeDirName string) (pwd string, dir string, err error) {
	// get the working directory
	pwd, err = os.Getwd()
	if err != nil {
		err = fmt.Errorf("could not get pwd: %s", pwd)
		return
	}
	logrus.Infof("PWD %s", pwd)

	// what is the directory?
	dir = fmt.Sprintf("%s/%s", pwd, relativeDirName)
	logrus.Infof("TMP DIR %s", dir)

	// remove the directory if it exists
	err = cleanupTestDir(dir)
	if err != nil {
		err = fmt.Errorf("could not create test temp dir (%s): %s", dir, err)
		return
	}

	// ensure the temp directory exists
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		err = fmt.Errorf("could not create test temp dir (%s): %s", dir, err)
		return
	}

	logrus.Infof("TMP DIR Created")
	return
}
