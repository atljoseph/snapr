package cli

import (
	"fmt"
	"snapr/util"

	"github.com/pieterclaerhout/go-waitgroup"
	"github.com/sirupsen/logrus"
)

// DeleteCmdRunE runs the delete command
// it is exported for testing
func DeleteCmdRunE(ropts *RootCmdOptions, opts *DeleteCmdOptions) error {
	funcTag := "delete"
	// logrus.Infof(funcTag)
	var err error

	// validate required arg
	if len(opts.S3Key) == 0 {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "option `--s3-key` is required")
	}

	// get a new aws session
	_, s3Client, err := util.NewS3Client(ropts.S3Config)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get new s3 client")
	}

	// track operated object keys
	operationTracker := &[]*util.S3Object{}

	if !opts.IsDir {

		// file
		file := util.S3Object{Key: opts.S3Key}

		// check if the objct exists
		exists, err := util.CheckS3ObjectExists(s3Client, ropts.Bucket, file.Key)
		if !exists || err != nil {
			return util.WrapError(fmt.Errorf("validation error"), funcTag, fmt.Sprintf("failed to confirm object existence, or object does not exist in bucket: '%s/%s'", ropts.Bucket, file.Key))
		}
		// logrus.Infof("Object exists: %s", file.Key)

		// delete the object from storage permanently
		err = util.DeleteS3Object(s3Client, ropts.Bucket, file.Key)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to delete object: %s", file.Key))
		}

		*operationTracker = append(*operationTracker, &file)
		logrus.Infof("Deleted: %s", file.Key)
	} else {

		// ensure ending dir slash for all these
		opts.S3Key = util.EnsureS3DirPath(opts.S3Key)

		// get all the objects in the bucket
		objects, _, err := util.ListS3ObjectsByKey(s3Client, ropts.Bucket, opts.S3Key, false)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to get a list bucket objects for key: %s", opts.S3Key))
		}

		// open a new wait group with a maximum number of concurrent workers
		wg := waitgroup.NewWaitGroup(100)

		// accumulate errors while awaiting
		errorTracker := &[]error{}

		// loop through all objects and spawn goroutines to wait for
		for _, object := range objects {

			// block adding until the next worker has finished
			wg.BlockAdd()

			// logrus.Infof("KEY: %s", object.Key)

			// on a separate goroutine, do something asyncronous
			// delete, accumulate
			go func(object *util.S3Object, tracker *[]*util.S3Object, etracker *[]error) {
				funcTag := "DeleteObjectWorker"
				defer wg.Done()

				// delete the object from storage permanently
				err := util.DeleteS3Object(s3Client, ropts.Bucket, object.Key)
				if err != nil {
					// log error, if any, and accumulate it
					err = util.WrapError(err, funcTag, fmt.Sprintf("failed to delete object: %s", object.Key))
					logrus.Warnf(err.Error())
					*etracker = append(*etracker, err)
				}

				// add to tracker
				*tracker = append(*tracker, object)

				// we need these injected here
			}(object, operationTracker, errorTracker)
		}

		// wait on everything to complete
		wg.Wait()

		logrus.Infof("Deleted all objects from %s", opts.S3Key)
	}

	logrus.Infof("%d objects deleted", len(*operationTracker))

	return nil
}
