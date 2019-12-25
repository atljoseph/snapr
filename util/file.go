package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type WalkedFile struct {
	// main attributes
	Path     string
	FileInfo os.FileInfo

	// used for aws s3
	S3Key string
}

// WalkFiles walks over a directory for files recursively
func WalkFiles(walkDir string) (files []*WalkedFile, err error) {
	funcTag := "WalkFiles"
	err = filepath.Walk(walkDir, WalkAllFilesHelper(&files))
	if err != nil {
		err = WrapError(err, funcTag, fmt.Sprintf("walking files in %s", walkDir))
		logrus.Warnf("walking helper error: %s", err)
		return
	}
	return
}

// WalkAllFilesHelper works with filepath.Walk(...)
// to build a slice of only file paths
func WalkAllFilesHelper(files *[]*WalkedFile) filepath.WalkFunc {
	funcTag := "WalkAllFilesHelper"
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return WrapError(err, funcTag, "walking helper error")
		} else if !info.Mode().IsDir() {
			// logrus.Infof("Walker %s", path)
			*files = append(*files, &WalkedFile{
				Path:     path,
				FileInfo: info,
			})
		}
		return nil
	}
}
