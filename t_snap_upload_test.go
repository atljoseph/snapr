package main

import (
	"path/filepath"
	"testing"

	"snapr/cli"
)

func Test3SnapUploadCommand(t *testing.T) {

	// ensure the temp directory exists
	_, testTempDir, err := ensureTestDir("test-3")
	if err != nil {
		t.Errorf("could not create test temp dir: %s", testTempDir)
	}
	// clean up on fail
	defer cleanupTestDir(testTempDir)

	// define the tests
	tests := []snapTest{
		{"dir & file & no cleanup", true,
			&cli.SnapCmdOptions{
				OutFile:            "toUpload.jpg",
				UploadAfterSuccess: true,
			}},
		{"dir & file & cleanup", true,
			&cli.SnapCmdOptions{
				OutFile:            "toUploadAndRemove.jpg",
				UploadAfterSuccess: true,
				CleanupAfterUpload: true,
			}},
		{"dir & users", true,
			&cli.SnapCmdOptions{
				OutDirUsers:        true,
				UploadAfterSuccess: true,
			}},
		{"dir & extra dir & users", true,
			&cli.SnapCmdOptions{
				OutDirExtra:        "extra",
				OutDirUsers:        true,
				UploadAfterSuccess: true,
			}},
		{"dir & webcam image", true,
			&cli.SnapCmdOptions{
				UseCamera:          true,
				UploadAfterSuccess: true,
			}},
		{"dir & screenshot", true,
			&cli.SnapCmdOptions{
				UseScreenshot:      true,
				UploadAfterSuccess: true,
			}},
		{"dir & BOTH screenshot and webcam", true,
			&cli.SnapCmdOptions{
				UseScreenshot:      true,
				UseCamera:          true,
				UploadAfterSuccess: true,
			}},
		{"dir & BOTH screenshot and webcam & cleanup", true,
			&cli.SnapCmdOptions{
				UseScreenshot:      true,
				UseCamera:          true,
				UploadAfterSuccess: true,
				CleanupAfterUpload: true,
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
