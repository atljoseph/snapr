package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"snapr/cli"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log.Println("Tests Starting")
	// EnvFilePath = ".tests.env"
	exitCode := m.Run()
	log.Println("Tests Done")
	os.Exit(exitCode)
}

type snapTest struct {
	description     string
	cmdOpts         cli.SnapCmdOptions
	expectedSuccess bool
}

func TestCommandSnap(t *testing.T) {

	// ensure the temp directory exists
	_, testTempDir, err := createTestDir("temp-test")
	if err != nil {
		t.Errorf("could not create test temp dir: %s", testTempDir)
	}
	// go ahead and clean up, then defer it for later, too
	defer cleanupTestDir(testTempDir)

	// tests
	// OutDir is set on all these to: testTempDir
	tests := []snapTest{
		{"basic with dir",
			cli.SnapCmdOptions{}, true},
		{"unsupported format",
			cli.SnapCmdOptions{
				Format: "xyz",
			}, false},
		{"supported format",
			cli.SnapCmdOptions{
				Format: "png",
			}, false},
		{"invalid device address",
			cli.SnapCmdOptions{
				CaptureDeviceAddr: "fake/device",
			}, false},
		{"basic with extra dir",
			cli.SnapCmdOptions{
				OutDirExtra: "extra",
			}, true},
		{"custom file name",
			cli.SnapCmdOptions{
				OutFileOverride: "test.jpg",
			}, true},
		{"custom file name with extra dir",
			cli.SnapCmdOptions{
				OutDirExtra:     "extra",
				OutFileOverride: "test.jpg",
			}, true},
		{"custom file name with format",
			cli.SnapCmdOptions{
				OutFileOverride: "test.jpg",
				Format:          "png", // should not create with this fmt
			}, true},
	}

	// loop through aand run tests
	for idx, test := range tests {

		logrus.Infof("TEST %d: %s", idx+1, test.description)

		// set args
		test.cmdOpts.OutDir = testTempDir
		cli.SnapCmdOpts = test.cmdOpts

		// run test
		err := cli.SnapCmdRunE([]string{})
		logrus.Infof("Command Ran")
		if err != nil {
			if test.expectedSuccess {
				reportError(t, test, fmt.Sprintf("command failed: %s", err))
			} else {
				logrus.Infof("Failed as expected")
			}
		}

		if err == nil {
			// get the output files
			// add the extra dir if the extra dir is specified
			var files []string
			logrus.Infof("Walking files")
			walkDir := testTempDir
			if len(test.cmdOpts.OutDirExtra) > 0 {
				walkDir += strings.ReplaceAll("/"+test.cmdOpts.OutDirExtra, "//", "/")
			}
			err = filepath.Walk(walkDir, walkFuncHelper(&files))
			if err != nil {
				reportError(t, test, fmt.Sprintf("walking dir %s", walkDir))
			}

			// ensure at least one file
			if len(files) == 0 {
				reportError(t, test, "output file not found. expected one")
			}

			// ensure not too many files
			if len(files) > 1 {
				reportError(t, test, "too many output files. expected one")
			}

			// check for filename
			if len(test.cmdOpts.OutFileOverride) > 0 {
				// if custom file name, check for exact filename
				if files[0] != walkDir+"/"+test.cmdOpts.OutFileOverride {
					reportError(t, test, fmt.Sprintf("output file name incorrect. expected %s, got %s", test.cmdOpts.OutFileOverride, files[0]))
				}
			} else {
				// do nothing for now
				if !strings.Contains(strings.ToLower(files[0]), test.cmdOpts.Format) {
					reportError(t, test, fmt.Sprintf("output file format incorrect. expected %s, got %s", test.cmdOpts.Format, files[0]))
				}
			}
			logrus.Infof("File validated")

			// if file, remove file
			if len(files) > 0 {
				err = os.Remove(files[0])
				if err != nil {
					reportError(t, test, "could not remove test file")
				}
				logrus.Infof("Test File removed")
			}
		}
	}
}

func reportError(t *testing.T, test snapTest, errMsg string) {
	t.Errorf("(%s) [%+v] %s", test.description, test.cmdOpts, errMsg)
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

func createTestDir(relativeDirName string) (pwd string, dir string, err error) {
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

func walkFuncHelper(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if !info.Mode().IsDir() {
			logrus.Infof("Walker %s", path)
			*files = append(*files, path)
		}
		return nil
	}
}
