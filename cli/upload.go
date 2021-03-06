package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"snapr/util"
	"strings"

	"github.com/pieterclaerhout/go-waitgroup"
	"github.com/sirupsen/logrus"
)

// UploadCmdRunE runs the snap command
// it is exported for testing
func UploadCmdRunE(ropts *RootCmdOptions, opts *UploadCmdOptions) error {
	funcTag := "upload"
	// logrus.Infof(funcTag)
	var err error

	// check limit, is it a crazy high number? if so kick it back
	if opts.UploadLimit > 100 {
		return util.WrapError(fmt.Errorf("Validation Error"), funcTag, "choose an upload limit smaller than 100")
	}

	// default the limit to 1 if 0
	// this situation can happen in testing, where the cobra args arent eval-ed
	if opts.UploadLimit < 1 {
		opts.UploadLimit = 1
	}

	// make sure that it is directory, we add an extra slash
	opts.S3Dir = util.EnsureS3DirPath(opts.S3Dir)

	// handle the dir and file inputs
	// and get a list of files based on the inputs
	var files []*util.WalkedFile
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
			return util.WrapError(fmt.Errorf("validation error"), funcTag, "`--file` cannot be a directory")
		}

		// append the walked file struct
		files = append(files, &util.WalkedFile{
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
				return util.WrapError(err, funcTag, "cannot get pwd for `--dir")
			}
		}

		// get the abs dir path
		opts.InDir, err = filepath.Abs(opts.InDir)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("cannot convert path for `--dir`: %s", opts.InDir))
		}
		logrus.Infof("ABS DIR: %s", opts.InDir)

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
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to walk dir for files to upload: %s", opts.InDir))
		}
	}

	logrus.Infof("Got %d file(s)", len(files))

	// TODO: order the files with the oldest first and newest last

	// filter out files without specific filename format
	var filteredFiles = files

	// if the option is set, it will filter out files by extension
	if len(opts.Formats) > 0 {
		logrus.Infof("Filtering formats: %s", strings.Join(opts.Formats, ","))

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
	uploadLimit := len(filteredFiles)
	// if upload limit is greater than 1, take the minimum of the length of files and the limit
	if opts.UploadLimit > 1 {
		uploadLimit = util.MinInt(opts.UploadLimit, len(filteredFiles))
	}
	logrus.Infof("Upload Limit: %d", uploadLimit)

	// truncate filtered files
	filteredFiles = filteredFiles[0:uploadLimit]
	logrus.Infof("Uploading %d file(s)", len(filteredFiles))

	// get a new aws session
	_, s3Client, err := util.NewS3Client(ropts.S3Config)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get new s3 client")
	}

	// get the base s3 key, if any
	baseS3Key := util.EnsureS3DirPath(opts.S3Dir)
	logrus.Infof("S3 Base Key: %s", baseS3Key)

	// TODO: add backup capability (later) - Get bucket contents by key recursively and check if same key

	// get the keys and determine if they exist
	for _, file := range filteredFiles {
		// logrus.Infof("?????: 1 %s 2 %s 3 %s 4 %s", opts.S3Dir, file.S3Key, file.Path, opts.InDir)
		// get the base s3 dir
		// first, get the key from the end of the filename
		file.S3Key = strings.ReplaceAll(file.Path, opts.InDir+"/", "")
		if len(opts.S3Dir) > 0 {
			file.S3Key = util.JoinS3Path(opts.S3Dir, file.S3Key)
		}
	}

	// set the object acl to "private"
	acl := "private"
	// unless set to public
	if opts.Public {
		acl = "public-read"
	}
	logrus.Infof("With Access ACL: %s", acl)

	// ------  UPLOADING -----------------------------------

	// open a new wait group with a maximum number of concurrent workers
	wg := waitgroup.NewWaitGroup(1)

	// track what is going on
	uploadTracker := &[]*util.WalkedFile{}
	errorTracker := &[]error{}

	// loop through all objects and spawn goroutines to wait for
	for _, waffle := range filteredFiles {

		// block adding until the next worker has finished
		wg.BlockAdd()

		go func(waffle *util.WalkedFile, accumulator *[]*util.WalkedFile, errorAccumulator *[]error) {
			funcTag := "HandleUploadFileWithCleanupWorker"
			defer wg.Done()

			// send to AWS
			_, err := util.WriteS3File(s3Client, ropts.Bucket, acl, waffle.S3Key, waffle)
			if err != nil {
				err = util.WrapError(err, funcTag, "failed to send file to s3: %s")
				*errorAccumulator = append(*errorAccumulator, err)
			}

			logrus.Infof("Uploaded key: %s", waffle.S3Key)

			// after success, cleanup the files
			if opts.CleanupAfterSuccess {

				// remove the file from the os if desired
				err = os.Remove(waffle.Path)
				if err != nil {
					err = util.WrapError(err, funcTag, "failed to remove local file from disk after upload")
					*errorAccumulator = append(*errorAccumulator, err)
				}

				logrus.Infof("Cleaned up file: %s", waffle.Path)
			}

			// append to images slice
			*accumulator = append(*accumulator, waffle)

		}(waffle, uploadTracker, errorTracker)
	}

	// wait on everything to complete
	wg.Wait()

	logrus.Infof("UPLOADED: %d", len(*uploadTracker))

	return nil
}
