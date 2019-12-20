package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"snapr/cli"
	"snapr/util"

	"github.com/sirupsen/logrus"
)

type snapTest struct {
	description     string
	expectedSuccess bool
	cmdOpts         *cli.SnapCmdOptions
}

// define the tests
// OutDir is set on all these programmatically (below)
var snapCommandTests = []snapTest{
	{"basic with dir and auto-generate file name", true,
		&cli.SnapCmdOptions{}},
	{"unsupported format", false,
		&cli.SnapCmdOptions{
			Format: "xyz",
		}},
	{"supported format", true,
		&cli.SnapCmdOptions{
			Format: "png",
		}},
	{"invalid device address", false,
		&cli.SnapCmdOptions{
			CaptureDeviceAddr: "fake/device",
		}},
	{"basic with extra dir", true,
		&cli.SnapCmdOptions{
			OutDirExtra: "extra",
		}},
	{"users dir with extra dir", true,
		&cli.SnapCmdOptions{
			OutDirExtra: "extra",
			OutDirUsers: true,
		}},
	{"custom file name", true,
		&cli.SnapCmdOptions{
			OutFileOverride: "test.jpg",
		}},
	{"custom file name with extra dir", true,
		&cli.SnapCmdOptions{
			OutDirExtra:     "extra",
			OutFileOverride: "test.jpg",
		}},
	{"custom file name with format", true,
		&cli.SnapCmdOptions{
			OutFileOverride: "test.jpg",
			Format:          "png", // should not create with this fmt
		}},
	{"snap with upload", true,
		&cli.SnapCmdOptions{
			OutFileOverride:    "toUpload.jpg",
			UploadAfterSuccess: true,
		}},
	{"snap with upload, then cleanup", true,
		&cli.SnapCmdOptions{
			OutFileOverride:    "toUploadAndRemove.jpg",
			UploadAfterSuccess: true,
			CleanupAfterUpload: true,
		}},
}

func TestCommandSnap(t *testing.T) {

	// ensure the temp directory exists
	_, testTempDir, err := ensureTestDir("test-snap")
	if err != nil {
		t.Errorf("could not create test temp dir: %s", testTempDir)
	}
	// go ahead and clean up, then defer it for later, too
	defer cleanupTestDir(testTempDir)

	// get the list of logged in users
	usersOutput, err := util.OSUsers()
	if err != nil {
		t.Errorf(wrapTestError("init TestCommandSnap", nil, fmt.Sprintf("getting users output")))
	}

	// loop through aand run tests
	for idx, test := range snapCommandTests {
		logrus.Infof("TEST %d (%s)", idx+1, test.description)

		// set the output dir (was lazy)
		test.cmdOpts.OutDir = filepath.Join(testTempDir, test.description)
		testOutDir := test.cmdOpts.OutDir

		// run test command
		err = cli.SnapCmdRunE(&cli.RootCmdOptions{}, test.cmdOpts)
		logrus.Infof("Command Ran")

		// what was expected vs. what was got?
		if err == nil && test.expectedSuccess {
			logrus.Infof("Command succeeded")
		} else if err != nil && test.expectedSuccess {
			t.Errorf(wrapSnapTestError(test, fmt.Sprintf("command failed: %s", err)))
		} else if err != nil && !test.expectedSuccess {
			logrus.Infof("Failed as ecxpected")
		} else if err == nil && !test.expectedSuccess {
			t.Errorf(wrapSnapTestError(test, fmt.Sprintf("command succeeded when expected to fail: %s", err)))
		} else {
			t.Errorf(wrapSnapTestError(test, fmt.Sprintf("unexpected error: %s", err)))
		}

		// continue, process, and reset for next test
		// if we make it into this block, then there should be a file
		if err == nil {
			// add the extra dir if the extra dir is specified
			if len(test.cmdOpts.OutDirExtra) > 0 {
				testOutDir = filepath.Join(testOutDir, test.cmdOpts.OutDirExtra)
			}

			// if users params is supplied, check that, too
			if test.cmdOpts.OutDirUsers {
				testOutDir = filepath.Join(testOutDir, usersOutput)
			}

			logrus.Infof("Walking files: %s", testOutDir)

			// get the output files
			files, err := util.WalkFiles(testOutDir)
			// err = filepath.Walk(walkDir, util.WalkAllFilesHelper(&files))
			if err != nil {
				t.Errorf(wrapSnapTestError(test, fmt.Sprintf("walking dir %s", testOutDir)))
			}

			// ensure at least one file
			if len(files) == 0 {
				// ignore this error if we are uploading
				if test.cmdOpts.CleanupAfterUpload {
					continue
				}
				t.Errorf(wrapSnapTestError(test, "output file not found. expected one"))
			}

			// ensure not too many files
			if len(files) > 1 {
				t.Errorf(wrapSnapTestError(test, "too many output files. expected one"))
			}

			// check that the output is in the correct folder
			if !strings.EqualFold(filepath.Dir(files[0].Path), testOutDir) {
				t.Errorf(wrapSnapTestError(test, fmt.Sprintf("output file in wrong output location. expected '%s', got '%s'", filepath.Dir(files[0].Path), testOutDir)))
			}

			// check for filename in test output
			if len(test.cmdOpts.OutFileOverride) > 0 {
				// if custom file name, check for exact filename
				if files[0].FileInfo.Name() != test.cmdOpts.OutFileOverride {
					t.Errorf(wrapSnapTestError(test, fmt.Sprintf("output file name incorrect when explicitly supplied. expected '%s', got '%s'", test.cmdOpts.OutFileOverride, files[0].FileInfo.Name())))
				}
			} else {
				// get the extension and check it
				if !strings.EqualFold(filepath.Ext(files[0].FileInfo.Name()), "."+test.cmdOpts.Format) {
					t.Errorf(wrapSnapTestError(test, fmt.Sprintf("output file format incorrect. expected '%s', got '%s'", test.cmdOpts.Format, filepath.Ext(files[0].FileInfo.Name()))))
				}
			}

			logrus.Infof("File validated")
		}
	}
}
