package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
			// filter out ".DS_Store" files
			if !strings.EqualFold(info.Name(), ".ds_store") {
				// logrus.Infof("Walker %s", path)
				*files = append(*files, &WalkedFile{
					Path:     path,
					FileInfo: info,
				})
			}
		}
		return nil
	}
}

// WriteFileBytes writes a new file from bytes
func WriteFileBytes(absFilePath string, byteSlice []byte) error {
	funcTag := "WalkAllFilesHelper"

	// ensure dir exists
	mkdir := filepath.Dir(absFilePath)
	// logrus.Infof("Ensuring Directory: %s", mkdir)
	err := os.MkdirAll(mkdir, 0700)
	// err = os.MkdirAll(mkdir, ropts.FileCreateMode)
	if err != nil {
		return WrapError(err, funcTag, fmt.Sprintf("falied to mkdir: %s", mkdir))
	}

	// Create new file
	newFile, err := os.Create(absFilePath)
	if err != nil {
		return WrapError(err, funcTag, fmt.Sprintf("failed to create new file: %s", absFilePath))
	}
	defer newFile.Close()

	// copy the data to the new file
	_, err = newFile.Write(byteSlice)
	if err != nil {
		return WrapError(err, funcTag, fmt.Sprintf("failed to write bytes to file: %s", absFilePath))
	}
	// logrus.Infof("COPIED: %d", bytesCopied)

	return nil
}
