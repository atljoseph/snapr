package main

import (
	"fmt"
	"os"
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

// define the tests
// InDir is set on all these programmatically (below)
var uploadCommandTests = []uploadTest{
	{"explicitly defined file when does not exist", false,
		&cli.UploadCmdOptions{
			InFile: "xyz.jpg",
		}},
	{"explicitly defined file when exists", true,
		&cli.UploadCmdOptions{
			InFile: "t_test.jpg",
		}},
	{"mismatched format, should fail", false,
		&cli.UploadCmdOptions{
			InFile:  "t_test.jpg",
			Formats: []string{"png"},
		}},
	{"specific format", true,
		&cli.UploadCmdOptions{
			InFile:  "t_test.jpg",
			Formats: []string{"jpg"},
		}},
	{"directory without file, should upload one test file", true,
		&cli.UploadCmdOptions{}},
	{"directory without file and with limit, should upload multiple", true,
		&cli.UploadCmdOptions{
			UploadLimit: 3,
		}},
	// this should always be the last test
	{"cleanup after success", true,
		&cli.UploadCmdOptions{
			UploadLimit:         3,
			CleanupAfterSuccess: true,
		}},
}

func TestCommandUpload(t *testing.T) {

	// ensure the temp directory exists
	_, testTempDir, err := ensureTestDir("test-upload")
	if err != nil {
		t.Errorf("could not create test temp dir: %s", testTempDir)
	}
	// go ahead and clean up, then defer it for later, too
	defer cleanupTestDir(testTempDir)

	// ensure a test file exists
	testFile1 := "t_test.jpg"
	_, err = copyFile(testFile1, filepath.Join(testTempDir, testFile1))
	if err != nil {
		t.Errorf("could not copy test image file")
	}

	// ensure another test file exists
	testFile2 := "t_testy.jpg"
	_, err = copyFile(testFile1, filepath.Join(testTempDir, testFile2))
	if err != nil {
		t.Errorf("could not copy test image file")
	}

	// keep these for use below when confirming
	testFiles := []string{
		testFile1,
		testFile2,
	}

	// loop through aand run tests
	for idx, test := range uploadCommandTests {
		logrus.Infof("TEST %d (%s)", idx+1, test.description)

		test.cmdOpts.InDir = testTempDir // filepath.Join(testTempDir, test.description)

		// run test command
		err := cli.UploadCmdRunE(&cli.RootCmdOptions{}, test.cmdOpts)
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

			if len(test.cmdOpts.InFile) > 0 {
				filesToConfirm = append(filesToConfirm, test.cmdOpts.InFile)
			} else {
				// assume anything below 1 is same as 1
				// if more files, append them all, until the milit
				if test.cmdOpts.UploadLimit > 1 {
					for _, testFile := range testFiles {
						if len(filesToConfirm) < test.cmdOpts.UploadLimit {
							filesToConfirm = append(filesToConfirm, testFile)
						}
					}
				}
			}

			// get a new aws session
			s, err := util.NewS3Client()
			if err != nil {
				logrus.Warnf("get aws session: %s", err)
			}

			// check in AWS
			for _, fileToConfirm := range filesToConfirm {

				// get the base file name
				fileToConfirm = filepath.Base(fileToConfirm)

				// check if the file exists in aws
				exists, err := util.CheckS3FileExists(s, fileToConfirm)
				if err != nil {
					logrus.Warnf("check file exists in aws: %s", err)
				}

				// report on existance
				if exists {
					logrus.Infof("File Exists in AWS: %s", fileToConfirm)
				} else {
					t.Errorf(wrapUploadTestError(test, fmt.Sprintf("file not found in aws: %s", err)))
				}
			}

			// check in OS
			// if cleanup, then make sure they aren't there (or there if not)
			for _, fileToConfirm := range filesToConfirm {
				// stat each file
				// if no error, then there is a problem
				fileToConfirm := filepath.Join(testTempDir, fileToConfirm)
				_, err := os.Stat(fileToConfirm)
				if err != nil {
					// ignore this error if we are cleaning up after uploading
					if test.cmdOpts.CleanupAfterSuccess {
						continue
					}
					// if not cleaning up, then should still be there
					t.Errorf(wrapUploadTestError(test, fmt.Sprintf("file found in os when expected to be deleted: %s", err)))
				}
			}

			logrus.Infof("Cleanup validated")

		}

	}
}
