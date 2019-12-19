package util

import (
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// WalkAllFilesHelper works with filepath.Walk(...)
// to build a slice of only file paths
func WalkAllFilesHelper(files *[]string) filepath.WalkFunc {
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
