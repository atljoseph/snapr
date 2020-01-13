package main

import (
	"io/ioutil"
	"snapr/util"
	"strings"

	"github.com/markbates/pkger"
	"github.com/sirupsen/logrus"
	"github.com/subosito/gotenv"
)

// LoadEnv loads an env file with packr and gotenv
func LoadEnv() (err error) {
	funcTag := "LoadEnv"
	envFilePath := "/.env"

	// load env
	// used this lib to be able to compile env into binary
	// do this first
	// logrus.Infof("PKGER: %s", envFilePath)
	f, err := pkger.Open(envFilePath)
	if err != nil {
		return util.WrapError(err, funcTag, "cannot load file from pgker")
	}
	defer f.Close()

	// then can stat
	// info, err := f.Stat()
	// if err != nil {
	// 	return util.WrapError(err, funcTag, "cannot stat file from pgker")
	// }
	// logrus.Infof("SIZE: %d", info.Size())

	// read the file
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return util.WrapError(err, funcTag, "cannot stat file from pgker")
	}
	envString := string(b)
	logrus.Infof("DATA: %s", envString)

	// apply env
	// used this lib because it is loadable from string
	logrus.Infof("Applying environment")
	gotenv.Apply(strings.NewReader(envString))
	// logrus.Infof(os.Getenv("SNAPR_VERSION"))

	return nil
}
