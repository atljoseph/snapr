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

func Test1SnapCommand(t *testing.T) {

	// ensure the temp directory exists
	_, testTempDir, err := ensureTestDir("test-1")
	if err != nil {
		t.Errorf("could not create test temp dir: %s", testTempDir)
	}
	// clean up on fail
	defer cleanupTestDir(testTempDir)

	// define the tests
	tests := []snapTest{
		{"dir & auto-gen file name & default to webcam image", true,
			&cli.SnapCmdOptions{}},
		{"dir & auto-gen file name & take webcam image", true,
			&cli.SnapCmdOptions{
				UseCamera: true,
			}},
		{"dir & auto-gen file name & take screenshot", true,
			&cli.SnapCmdOptions{
				UseScreenshot: true,
			}},
		{"dir & auto-gen file name & take BOTH screenshot and webcam", true,
			&cli.SnapCmdOptions{
				UseScreenshot: true,
				UseCamera:     true,
			}},
		{"dir & unsupported format", false,
			&cli.SnapCmdOptions{
				Format: "xyz",
			}},
		{"dir & supported format", true,
			&cli.SnapCmdOptions{
				Format: "png",
			}},
		{"dir & invalid device address", false,
			&cli.SnapCmdOptions{
				CaptureDeviceAddr: "fake/device",
			}},
		{"dir & extra dir", true,
			&cli.SnapCmdOptions{
				OutDirExtra: "extra",
			}},
		{"dir & extra & users dir", true,
			&cli.SnapCmdOptions{
				OutDirExtra: "extra",
				OutDirUsers: true,
			}},
		{"dir & custom file name", true,
			&cli.SnapCmdOptions{
				OutFile: "test.jpg",
			}},
		{"dir & extra dir & custom file name", true,
			&cli.SnapCmdOptions{
				OutDirExtra: "extra",
				OutFile:     "test.jpg",
			}},
		{"custom file name with format", true,
			&cli.SnapCmdOptions{
				OutFile: "test.jpg",
				Format:  "png", // should not create with this fmt
			}},
	}

	// tack on the out dir with descriptions
	for _, test := range tests {
		// set the output dir (was lazy)
		test.cmdOpts.OutDir = filepath.Join(testTempDir, test.description)
	}

	// OutDir is set on all these programmatically (below)
	testCommandSnap(t, testTempDir, tests)
}

func testCommandSnap(t *testing.T, testTempDir string, tests []snapTest) {

	// get the list of logged in users
	usersOutput, err := util.OSUsers()
	if err != nil {
		t.Errorf(wrapTestError("init TestCommandSnap", nil, fmt.Sprintf("getting users output")))
	}

	// loop through aand run tests
	for idx, test := range tests {
		logrus.Infof("TEST %d (%s)", idx+1, test.description)

		// don't let these mutate
		testOutDir := test.cmdOpts.OutDir // TODO: this is being ignored when custom file is used
		testOutDirExtra := test.cmdOpts.OutDirExtra
		testOutFile := test.cmdOpts.OutFile
		testUseCamera := test.cmdOpts.UseCamera
		testUseScreenshot := test.cmdOpts.UseScreenshot

		// run test command
		err = cli.SnapCmdRunE(testRootCmdOpts, test.cmdOpts)
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

			// order is important for these folders

			// add the extra dir if the extra dir is specified
			if len(testOutDirExtra) > 0 {
				testOutDir = filepath.Join(testOutDir, testOutDirExtra)
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
				t.Errorf(wrapSnapTestError(test, fmt.Sprintf("walking dir: %s: %s", testOutDir, err)))
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
			if !(testUseCamera && testUseScreenshot) {
				// expect no more than one file
				if len(files) > 1 {
					t.Errorf(wrapSnapTestError(test, "too many output files. expected one"))
				}
			} else {
				// should be at least 2 files
				if len(files) < 2 {
					t.Errorf(wrapSnapTestError(test, "not enough output files. expected at least 2"))
				}
			}

			// TODO: check that 1 output file containing "screen" / "display" in filename
			if testUseScreenshot {

				// check for match in file name
				match := false
				for _, file := range files {

					// see if filename contains "screen"
					if strings.Contains(strings.ToLower(file.FileInfo.Name()), strings.ToLower("screen")) {
						match = true
					}

					// exit loop on match
					if match {
						break
					}
				}

				// error if no match
				if !match {
					t.Errorf(wrapSnapTestError(test, "could not find screenshot file"))
				}
			}

			// check that the output is in the correct folder
			if !strings.EqualFold(filepath.Dir(files[0].Path), testOutDir) {
				t.Errorf(wrapSnapTestError(test, fmt.Sprintf("output file in wrong output location. expected '%s', got '%s'", filepath.Dir(files[0].Path), testOutDir)))
			}

			// check for filename in test output
			if len(testOutFile) > 0 {
				// if custom file name, check for exact filename
				toConfirm := filepath.Join(testOutDir, testOutFile)
				if files[0].Path != toConfirm {
					t.Errorf(wrapSnapTestError(test, fmt.Sprintf("output file name incorrect when explicitly supplied. expected '%s', got '%s'", toConfirm, files[0].Path)))
				}
			} else {
				// get the extension and check it
				ext := filepath.Ext(files[0].FileInfo.Name())
				if !strings.EqualFold(ext, "."+test.cmdOpts.Format) {
					t.Errorf(wrapSnapTestError(test, fmt.Sprintf("output file format incorrect. expected '%s', got '%s'", test.cmdOpts.Format, ext)))
				}
			}

			logrus.Infof("File validated")
		}
	}
}
