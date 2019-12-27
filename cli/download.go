package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"snapr/util"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

// DownloadCmdRunE runs the download command
// it is exported for testing
func DownloadCmdRunE(ropts *RootCmdOptions, opts *DownloadCmdOptions) error {
	funcTag := "download"
	// logrus.Infof(funcTag)
	var err error

	// default the out dir if empty
	if len(opts.OutDir) == 0 {
		// default to the directory where the binary exists (pwd)
		opts.OutDir, err = os.Getwd()
		if err != nil {
			return util.WrapError(err, funcTag, "failed to get pwd for OutDir")
		}
	}

	logrus.Infof("KEY: %s, OUT: %s", opts.S3Key, opts.OutDir)

	// get a new aws session
	_, s3Client, err := util.NewS3Client(ropts.S3Config)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get new s3 client")
	}

	// track operated object keys
	var operationTracker []*util.S3Object

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
		operationTracker = append(operationTracker, &object)
		logrus.Infof("Downloaded %s to %s", object.Key, absFilePath)
	} else {

		// get all the objects in the bucket
		objects, _, err := util.ListS3ObjectsByKey(s3Client, ropts.Bucket, opts.S3Key, false)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to get a list bucket objects for key: %s", opts.S3Key))
		}

		// open a new go errgroup
		eg, _ := util.NewErrGroup()

		for _, object := range objects {

			// file
			absFilePath := filepath.Join(opts.OutDir, object.Key)

			// logrus.Infof("KEY: %s", object.Key)
			eg.Go(DownloadObjectWorker(s3Client, ropts.Bucket, object, absFilePath, &operationTracker))
		}

		// wait on the errgroup and check for error
		err = eg.Wait()
		if err != nil {
			return util.WrapError(err, funcTag, "failed to download s3 objects in errgroup")
		}

		logrus.Infof("Downloaded all objects from %s", opts.S3Key)
	}

	logrus.Infof("%d objects downloaded", len(operationTracker))

	return nil
}

// DownloadObjectWorker returns signature of `func() error {}` to satisfy a closure
func DownloadObjectWorker(s3Client *s3.S3, bucket string, object *util.S3Object, absFilePath string, tracker *[]*util.S3Object) func() error {
	funcTag := "DownloadObjectWorker"
	return func() error {

		// delete the object from storage permanently
		byteSlice, err := util.DownloadS3Object(s3Client, bucket, object.Key)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to delete object: %s", object.Key))
		}

		// write the file
		err = util.WriteFileBytes(absFilePath, byteSlice)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to write file: %s", absFilePath))
		}

		// add to tracker
		*tracker = append(*tracker, object)

		// bail
		return nil
	}
}
