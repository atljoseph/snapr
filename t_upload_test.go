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
	{"file name when does not exist", false,
		&cli.UploadCmdOptions{
			InFile: "xyz.jpg",
		}},
	{"file name when exists", true,
		&cli.UploadCmdOptions{
			InFile: "t_test.jpg",
		}},
	{"file name & s3 base dir", true,
		&cli.UploadCmdOptions{
			InFile: "t_test.jpg",
			S3Dir:  "test-base-dir",
		}},
	{"file & mismatched format, should fail", false,
		&cli.UploadCmdOptions{
			InFile:  "t_test.jpg",
			Formats: []string{"png"},
		}},
	{"file & matched format filter", true,
		&cli.UploadCmdOptions{
			InFile:  "t_test.jpg",
			Formats: []string{"jpg"},
		}},
	{"dir & no file, should upload one test file", true,
		&cli.UploadCmdOptions{}},
	{"dir & file & limit, should upload multiple", true,
		&cli.UploadCmdOptions{
			UploadLimit: 3,
		}},
	{"dir & file & limit & base s3 dir", true,
		&cli.UploadCmdOptions{
			UploadLimit: 3,
			S3Dir:       "test-base-dir",
		}},
	{"dir & upload limit too high, should fail", false,
		&cli.UploadCmdOptions{
			UploadLimit: 101,
		}},
	// this should always be the last test
	{"dir & limit more than exists & cleanup after success", true,
		&cli.UploadCmdOptions{
			UploadLimit:         10,
			CleanupAfterSuccess: true,
		}},
}

func Test2CommandUpload(t *testing.T) {

	// ensure the temp directory exists
	_, testTempDir, err := ensureTestDir("test-2")
	if err != nil {
		t.Errorf("could not create test temp dir: %s", testTempDir)
	}
	// clean up on fail
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

		// don't let these mutate
		testInFile := test.cmdOpts.InFile
		testS3Dir := test.cmdOpts.S3Dir
		testUploadLimit := test.cmdOpts.UploadLimit

		// run test command
		err := cli.UploadCmdRunE(testRootCmdOpts, test.cmdOpts)
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
			var keysToConfirm []string
			var filesToConfirm []string

			if len(testInFile) > 0 {
				// get the base s3 dir
				baseS3Key := testInFile
				if len(testS3Dir) > 0 {
					baseS3Key = testS3Dir + util.S3Delimiter + testInFile
				}
				// set the file and key for checking
				filesToConfirm = append(filesToConfirm, testInFile)
				keysToConfirm = append(keysToConfirm, baseS3Key)
			} else {
				// assume anything below 1 is same as 1
				if testUploadLimit < 1 {
					testUploadLimit = 1
				}
				// if more files, append them all, until the milit
				if testUploadLimit > 1 {
					for _, testFile := range testFiles {
						if len(filesToConfirm) < testUploadLimit {
							// get the base s3 dir
							baseS3Key := testFile
							if len(testS3Dir) > 0 {
								baseS3Key = testS3Dir + util.S3Delimiter + testFile
							}
							// set the file and key for checking
							filesToConfirm = append(filesToConfirm, testFile)
							keysToConfirm = append(keysToConfirm, baseS3Key)
						}
					}
				}
			}

			// get a new aws session
			_, s3Client, err := util.NewS3Client(testRootCmdOpts.S3Config)
			if err != nil {
				logrus.Warnf("get aws session: %s", err)
			}

			// check in AWS
			for _, keyToConfirm := range keysToConfirm {

				// get the base file name
				keyToConfirm = filepath.Base(keyToConfirm)

				// check if the file exists in aws
				exists, err := util.CheckS3ObjectExists(s3Client, testRootCmdOpts.Bucket, keyToConfirm)
				if err != nil {
					logrus.Warnf("check file exists in aws: %s", err)
				}

				// report on existance
				if exists {
					logrus.Infof("File Exists in AWS: %s", keyToConfirm)
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
