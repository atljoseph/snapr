package main

import (
	"os"
	"runtime"
	"snapr/cli"

	"github.com/sirupsen/logrus"
)

func main() {

	// log the runtime OS code
	logrus.Infof("OS: %s", runtime.GOOS)

	// load the environment
	err := LoadEnv()
	if err != nil {
		logrus.Warnf(err.Error())
	}

	// start the cli and react if there is an error
	err = cli.Execute()
	if err != nil {
		// warn for visibility
		logrus.Warnf(err.Error())
		// exit with status 0 for PAM
		// would normally use code 1
		os.Exit(0)
	}
}
