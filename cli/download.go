package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"snapr/util"

	"github.com/pieterclaerhout/go-waitgroup"
	"github.com/sirupsen/logrus"
)

// DownloadCmdRunE runs the download command
// it is exported for testing
func DownloadCmdRunE(ropts *RootCmdOptions, opts *DownloadCmdOptions) error {
	funcTag := "download"
	// logrus.Infof(funcTag)
	var err error

	// not validating the dir here, because you might want to download the entire dir ("")

	// default the out dir if empty
	if len(opts.OutDir) == 0 {
		// default to the directory where the binary exists (pwd)
		opts.OutDir, err = os.Getwd()
		if err != nil {
			return util.WrapError(err, funcTag, "failed to get pwd for output")
		}
	}

	logrus.Infof("KEY: %s, OUT: %s", opts.S3Key, opts.OutDir)

	// get a new aws session
	_, s3Client, err := util.NewS3Client(ropts.S3Config)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get new s3 client")
	}

	// track operated object keys
	operationTracker := &[]*util.S3Object{}

	if !opts.IsDir {

		// file
		absFilePath := filepath.Join(opts.OutDir, opts.S3Key)
		object := util.S3Object{Key: opts.S3Key}

		// check if the objct exists
		exists, err := util.CheckS3ObjectExists(s3Client, ropts.Bucket, object.Key)
		if !exists || err != nil {
			return util.WrapError(fmt.Errorf("validation error"), funcTag, fmt.Sprintf("failed to confirm object existence, or object does not exist in bucket: '%s/%s'", ropts.Bucket, object.Key))
		}
		// logrus.Infof("Object exists: %s", file.Key)

		// delete the object from storage permanently
		byteSlice, err := util.DownloadS3Object(s3Client, ropts.Bucket, object.Key)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to download object: %s", object.Key))
		}

		// write the file
		err = util.WriteFileBytes(absFilePath, byteSlice)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to write file: %s", absFilePath))
		}

		// track
		*operationTracker = append(*operationTracker, &object)
		logrus.Infof("Downloaded %s to %s", object.Key, absFilePath)
	} else {

		// ensure ending dir slash for all these
		opts.S3Key = util.EnsureS3DirPath(opts.S3Key)

		// get all the objects in the bucket
		objects, _, err := util.ListS3ObjectsByKey(s3Client, ropts.Bucket, opts.S3Key, false)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to get a list bucket objects for key: %s", opts.S3Key))
		}

		// open a new wait group with a maximum number of concurrent workers
		wg := waitgroup.NewWaitGroup(50)

		// accumulate errors while awaiting
		errorTracker := &[]error{}

		// loop through all objects and spawn goroutines to wait for
		for _, object := range objects {

			// block adding until the next worker has finished
			wg.BlockAdd()

			// file
			absFilePath := filepath.Join(opts.OutDir, object.Key)

			// logrus.Infof("KEY: %s", object.Key)

			// on a separate goroutine, do something asyncronous
			// download, write, accumulate
			go func(object *util.S3Object, absFilePath string, tracker *[]*util.S3Object, eTracker *[]error) {
				funcTag := "DownloadObjectWorker"
				defer wg.Done()

				// delete the object from storage permanently
				byteSlice, err := util.DownloadS3Object(s3Client, ropts.Bucket, object.Key)
				if err != nil {
					err = util.WrapError(err, funcTag, fmt.Sprintf("failed to download object: %s", object.Key))
					logrus.Warnf(err.Error())
					*eTracker = append(*eTracker, err)
				}

				// write the file
				err = util.WriteFileBytes(absFilePath, byteSlice)
				if err != nil {
					err = util.WrapError(err, funcTag, fmt.Sprintf("failed to write file: %s", absFilePath))
					logrus.Warnf(err.Error())
					*eTracker = append(*eTracker, err)
				}

				// add to tracker
				*tracker = append(*tracker, object)

				// we need these
			}(object, absFilePath, operationTracker, errorTracker)
		}

		// wait on everything to complete
		wg.Wait()

		logrus.Infof("Downloaded all objects from %s", opts.S3Key)
	}

	logrus.Infof("%d objects downloaded", len(*operationTracker))

	return nil
}
