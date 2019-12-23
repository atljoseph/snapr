package main

import (
	"testing"

	"snapr/cli"
)

func Test3SnapUploadCommand(t *testing.T) {
	// define the tests
	// OutDir is set on all these programmatically (below)
	testCommandSnap(t, []snapTest{
		{"snap with upload and no cleanup", true,
			&cli.SnapCmdOptions{
				OutFile:            "toUpload.jpg",
				UploadAfterSuccess: true,
			}},
		{"snap with upload, then cleanup", true,
			&cli.SnapCmdOptions{
				OutFile:            "toUploadAndRemove.jpg",
				UploadAfterSuccess: true,
				CleanupAfterUpload: true,
			}},
		{"snap with upload and users dir", true,
			&cli.SnapCmdOptions{
				OutDirUsers:        true,
				UploadAfterSuccess: true,
			}},
		{"snap with upload and users dir and extra dir", true,
			&cli.SnapCmdOptions{
				OutDirExtra:        "extra",
				OutDirUsers:        true,
				UploadAfterSuccess: true,
			}},
	})
}
