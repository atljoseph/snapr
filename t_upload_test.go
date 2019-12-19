package main

import (
	"fmt"
	"path/filepath"
	"snapr/cli"
	"snapr/util"
	"testing"

	"github.com/sirupsen/logrus"
)

type uploadTest struct {
	description     string
	expectedSuccess bool
	cmdOpts         *cli.UploadCmdOptions
}

func TestCommandUpload(t *testing.T) {

	// ensure the temp directory exists
	_, testTempDir, err := ensureTestDir("test-upload")
	if err != nil {
		t.Errorf("could not create test temp dir: %s", testTempDir)
	}
	// go ahead and clean up, then defer it for later, too
	defer cleanupTestDir(testTempDir)

	// TODO: ensure a test file exists

	tests := []uploadTest{
		{"no file defined should fail", false,
			&cli.UploadCmdOptions{}},
		{"explicitly defined file when does not exist", false,
			&cli.UploadCmdOptions{
				InFileOverride: "xyz.jpg",
			}},
		{"explicitly defined file when exists", true,
			&cli.UploadCmdOptions{
				InFileOverride: "../t_test.jpg",
			}},
	}

	// set the base dir (because i was lazy)
	for _, test := range tests {
		test.cmdOpts.InDir = testTempDir
	}

	// loop through aand run tests
	for idx, test := range tests {
		logrus.Infof("TEST %d (%s)", idx+1, test.description)

		// run test command
		err := cli.UploadCmdRunE(test.cmdOpts)
		logrus.Infof("Command Ran")

		// what was expected vs. what was got?
		if err == nil && test.expectedSuccess {
			logrus.Infof("Command succeeded")
		} else if err != nil && test.expectedSuccess {
			t.Errorf(wrapUploadTestError(test, fmt.Sprintf("command failed: %s", err)))
		} else if err != nil && !test.expectedSuccess {
			logrus.Infof("Failed as ecxpected")
		} else if err == nil && !test.expectedSuccess {
			t.Errorf(wrapUploadTestError(test, fmt.Sprintf("command succeeded when expected to fail")))
		} else {
			t.Errorf(wrapUploadTestError(test, fmt.Sprintf("unexpected error: %s", err)))
		}

		// continue, process, and reset for next test
		// if we make it into this block, then the file most likely made it to aws
		if err == nil {

			// TODO: check all uploaded files (when multiples)
			var filesToConfirm []string
			filesToConfirm = append(filesToConfirm, test.cmdOpts.InFileOverride)

			// get a new aws session
			s, err := util.NewAwsSession()
			if err != nil {
				logrus.Warnf("get aws session: %s", err)
			}

			for _, fileToConfirm := range filesToConfirm {

				// get the base file path
				fileToConfirm = filepath.Base(fileToConfirm)

				// check if the file exists in aws
				exists, err := util.CheckAwsFileExists(s, fileToConfirm)
				if err != nil {
					logrus.Warnf("check aws file exists: %s", err)
				}

				// report on existance
				if exists {
					logrus.Infof("File Exists: %s", fileToConfirm)
				} else {
					t.Errorf(wrapUploadTestError(test, fmt.Sprintf("file not found: %s", err)))
				}
			}

		}

	}
}
