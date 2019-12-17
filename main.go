package main

import (
	"snapr/cli"
	"strings"

	"github.com/gobuffalo/packr"
	"github.com/sirupsen/logrus"
	"github.com/subosito/gotenv"
)

func main() {

	// load env
	// used this lib to be able to compile env into binary
	logrus.Infof("Loading environment")
	box := packr.NewBox("./")
	envString, err := box.FindString(".env")
	if err != nil {
		logrus.Warnf(err.Error())
	}

	// apply env
	// used this lib because it is loadable from string
	logrus.Infof("Applying environment")
	gotenv.Apply(strings.NewReader(envString))

	// logrus.Infof(envString)

	cli.Execute()
}
