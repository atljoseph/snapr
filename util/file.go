package util

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type WalkedFile struct {
	Path     string
	FileInfo os.FileInfo
}

// WalkFiles walks over a directory for files recursively
func WalkFiles(walkDir string) (files []WalkedFile, err error) {
	funcTag := "WalkFiles"
	err = filepath.Walk(walkDir, WalkAllFilesHelper(&files))
	if err != nil {
		err = WrapError(err, funcTag, "walking files")
		return
	}
	return
}

// WalkAllFilesHelper works with filepath.Walk(...)
// to build a slice of only file paths
func WalkAllFilesHelper(files *[]WalkedFile) filepath.WalkFunc {
	funcTag := "WalkAllFilesHelper"
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return WrapError(err, funcTag, "walking helper error")
		} else if !info.Mode().IsDir() {
			logrus.Infof("Walker %s", path)
			*files = append(*files, WalkedFile{
				Path:     path,
				FileInfo: info,
			})
		}
		return nil
	}
}
