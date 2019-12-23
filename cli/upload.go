package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"snapr/util"
	"strings"

	"github.com/sirupsen/logrus"
)

// UploadCmdRunE runs the snap command
// it is exported for testing
func UploadCmdRunE(ropts *RootCmdOptions, opts *UploadCmdOptions) error {
	funcTag := "upload"
	logrus.Infof(funcTag)

	// get a new aws session
	_, s3Client, err := util.NewS3Client(ropts.S3Config)
	if err != nil {
		return util.WrapError(err, funcTag, "get new aws session")
	}

	// check limit, is it a crazy high number? if so kick it back
	if opts.UploadLimit > 100 {
		return util.WrapError(fmt.Errorf("Validation Error"), funcTag, "choose an upload limit smaller than 100")
	}

	// default the limit to 1 if 0
	// this situation can happen in testing, where the cobra args arent eval-ed
	if opts.UploadLimit < 1 {
		opts.UploadLimit = 1
	}

	// formats default (weird thing with cobra input slice)
	if len(opts.Formats) == 1 {
		if len(strings.Trim(opts.Formats[0], " ")) == 0 {
			opts.Formats = nil
		}
	}

	// handle the dir and file inputs
	// and get a list of files based on the inputs
	var files []util.WalkedFile
	if len(opts.InFile) > 0 {
		// if the file override is set,
		// ignore the walk and upload limit

		// if dir is also set, join
		if len(opts.InDir) > 0 {
			opts.InFile = filepath.Join(opts.InDir, opts.InFile)
		}

		// get the abs file path
		absPath, err := filepath.Abs(opts.InFile)
		if err != nil {
			logrus.Warnf("cannot convert path to absolute file path: %s", opts.InFile)
		}

		// set these explicitly
		opts.InDir = filepath.Dir(absPath)
		opts.InFile = filepath.Base(absPath)

		// join the override file path with the dir
		fullPath := filepath.Join(opts.InDir, opts.InFile)

		// stat the path
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return util.WrapError(err, funcTag, "cannot stat path")
		}

		// ensure is a file
		if fileInfo.IsDir() {
			return util.WrapError(fmt.Errorf("validation error"), funcTag, "file cannot be a directory")
		}

		// append the walked file struct
		files = append(files, util.WalkedFile{
			Path:     fullPath,
			FileInfo: fileInfo,
		})
	} else {
		// file override is empty

		// default the in dir if empty
		if len(opts.InDir) == 0 {
			// default to the directory where the binary exists (pwd)
			opts.InDir, err = os.Getwd()
			if err != nil {
				return util.WrapError(err, funcTag, "cannot get pwd for InDir")
			}
		}

		// get the abs dir path
		opts.InDir, err = filepath.Abs(opts.InDir)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("cannot convert path to absolute dir path: %s", opts.InDir))
		}

		// stat the path
		fileInfo, err := os.Stat(opts.InDir)
		if err != nil {
			return util.WrapError(err, funcTag, "cannot stat path")
		}

		// ensure is a dir
		if !fileInfo.IsDir() {
			return util.WrapError(fmt.Errorf("validation error"), funcTag, "dir provided is not a directory")
		}

		// get the slice of walkedFiles
		// based on the indir, walk all files
		files, err = util.WalkFiles(opts.InDir)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("walking dir for files to upload: %s", opts.InDir))
		}
	}

	logrus.Infof("Got %d files before filtering", len(files))

	// TODO: order the files with the oldest first and newest last

	// filter out files without specific filename format
	var filteredFiles = files

	// if the option is set, it will filter out files by extension
	if len(opts.Formats) > 0 {
		logrus.Warnf("%s %d", opts.Formats[0], len(opts.Formats))

		// reset and append
		filteredFiles = nil
		for _, file := range files {

			// get the file extension, and replace the dot (weirdness of this lib)
			fileExt := strings.ReplaceAll(filepath.Ext(file.Path), ".", "")

			// filter formats
			for _, format := range opts.Formats {
				// if the format matches
				if strings.EqualFold(fileExt, format) {
					// append the file to the slice
					filteredFiles = append(filteredFiles, file)
				}
			}
		}
		logrus.Infof("Got %d files after filtering", len(filteredFiles))
	}

	// if no files after filtering, error
	if len(filteredFiles) == 0 {
		return util.WrapError(fmt.Errorf("Validation Error"), funcTag, "no files with specified format exist at target")
	}

	// attempt to chop off a slice of these equal to the limit input
	length := len(filteredFiles)
	// if upload limit is greater than 1, take the minimum of the length of files and the limit
	if opts.UploadLimit > 1 {
		length = util.MinInt(opts.UploadLimit, len(filteredFiles))
	}

	// truncate filtered files
	filteredFiles = filteredFiles[0:length]
	logrus.Infof("Got %d files to upload", len(filteredFiles))

	// loop to upload the files
	for _, file := range filteredFiles {

		logrus.Infof("Uploading %s %+v", file.Path, file)

		// get the base s3 dir
		// first, get the key from the end of the filename
		key := strings.ReplaceAll(file.Path, opts.InDir+"/", "")
		if len(opts.S3Dir) > 0 {
			key = opts.S3Dir + util.S3Delimiter + key
		}

		// send to AWS
		key, err := util.SendToS3(s3Client, ropts.Bucket, file, key)
		if err != nil {
			return util.WrapError(err, funcTag, "sending file to aws")
		}

		logrus.Infof("Done uploading key: %s", key)

		// after success, cleanup the files
		if opts.CleanupAfterSuccess {

			// remove the file from the os if desired
			err = os.Remove(file.Path)
			if err != nil {
				return util.WrapError(err, funcTag, "removing the file after upload")
			}

			logrus.Infof("Cleaned up: %s", file.Path)
		}
	}

	return nil
}
