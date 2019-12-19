package main

import (
	"fmt"
	"os"
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

func TestCommandSnap(t *testing.T) {

	// ensure the temp directory exists
	_, testTempDir, err := ensureTestDir("test-snap")
	if err != nil {
		t.Errorf("could not create test temp dir: %s", testTempDir)
	}
	// go ahead and clean up, then defer it for later, too
	defer cleanupTestDir(testTempDir)

	// tests
	// OutDir is set on all these to: testTempDir
	tests := []snapTest{
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
	}

	// set the base dir (because i was lazy)
	for _, test := range tests {
		test.cmdOpts.OutDir = testTempDir
	}

	// loop through aand run tests
	for idx, test := range tests {
		logrus.Infof("TEST %d (%s)", idx+1, test.description)

		// run test command
		err := cli.SnapCmdRunE(test.cmdOpts)
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
			// get the output files
			// add the extra dir if the extra dir is specified
			var files []string
			logrus.Infof("Walking files")
			walkDir := testTempDir
			if len(test.cmdOpts.OutDirExtra) > 0 {
				walkDir += strings.ReplaceAll("/"+test.cmdOpts.OutDirExtra, "//", "/")
			}
			err = filepath.Walk(walkDir, util.WalkAllFilesHelper(&files))
			if err != nil {
				t.Errorf(wrapSnapTestError(test, fmt.Sprintf("walking dir %s", walkDir)))
			}

			// ensure at least one file
			if len(files) == 0 {
				t.Errorf(wrapSnapTestError(test, "output file not found. expected one"))
			}

			// ensure not too many files
			if len(files) > 1 {
				t.Errorf(wrapSnapTestError(test, "too many output files. expected one"))
			}

			// check for filename
			if len(test.cmdOpts.OutFileOverride) > 0 {
				// if custom file name, check for exact filename
				if files[0] != walkDir+"/"+test.cmdOpts.OutFileOverride {
					t.Errorf(wrapSnapTestError(test, fmt.Sprintf("output file name incorrect when explicitly supplied. expected %s, got %s", test.cmdOpts.OutFileOverride, files[0])))
				}
			} else {
				// get the extension and check it
				if !strings.EqualFold(filepath.Ext(files[0]), "."+test.cmdOpts.Format) {
					t.Errorf(wrapSnapTestError(test, fmt.Sprintf("output file format incorrect. expected %s, got %s", test.cmdOpts.Format, filepath.Ext(files[0]))))
				}
			}

			logrus.Infof("File validated")

			// if file, remove file
			if len(files) > 0 {
				err = os.Remove(files[0])
				if err != nil {
					t.Errorf(wrapSnapTestError(test, "could not remove test file"))
				}
				logrus.Infof("Test File removed")
			}
		}
	}
}
