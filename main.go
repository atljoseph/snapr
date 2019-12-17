package main

import (
	"snapr/cli"
	"strings"

	"github.com/gobuffalo/packr"
	"github.com/sirupsen/logrus"
	"github.com/subosito/gotenv"
)

var (
	// EnvFilePath is the path to `.env` file
	// this can be set by `-ldflags` upon running `go build -ldflags "-X main.EnvFilePath=.other.env" snapr`
	EnvFilePath string
)

func main() {

	if len(EnvFilePath) == 0 {
		EnvFilePath = ".env"
	}

	// load env
	// used this lib to be able to compile env into binary
	logrus.Infof("Loading environment")
	box := packr.NewBox("./")
	envString, err := box.FindString(EnvFilePath)
	if err != nil {
		logrus.Warnf(err.Error())
	}
	// logrus.Infof(envString)

	// apply env
	// used this lib because it is loadable from string
	logrus.Infof("Applying environment")
	gotenv.Apply(strings.NewReader(envString))
	// logrus.Infof(os.Getenv("S3_BUCKET"))

	// start the cli
	cli.Execute()
}
