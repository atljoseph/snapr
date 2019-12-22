package main

import (
	"os"
	"snapr/cli"
	"testing"

	"github.com/sirupsen/logrus"
)

// normally, cobra would handle these, but it cannot since we are testing
// and we have to handle it this way so that we do not show the sensitive data to users
var testRootCmdOpts *cli.RootCmdOptions

func TestMain(m *testing.M) {
	logrus.Infof("Tests Starting")

	// TODO: ? override packr if needed?
	// EnvFilePath = ".tests.env"

	// init root options from env
	testRootCmdOpts = &cli.RootCmdOptions{}
	testRootCmdOpts = testRootCmdOpts.SetupS3ConfigFromRootArgs()

	logrus.Infof("S3: %+v", testRootCmdOpts)

	// run the tests
	exitCode := m.Run()
	logrus.Infof("Tests Done")
	os.Exit(exitCode)
}
