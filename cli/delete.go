package cli

import (
	"fmt"
	"snapr/util"

	"github.com/aws/aws-sdk-go/service/s3"
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
	var operationTracker []*util.S3Object

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

		operationTracker = append(operationTracker, &file)
		logrus.Infof("Deleted: %s", file.Key)
	} else {

		// get all the objects in the bucket
		objects, _, err := util.ListS3ObjectsByKey(s3Client, ropts.Bucket, opts.S3Key, false)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to get a list bucket objects for key: %s", opts.S3Key))
		}

		// open a new go errgroup
		eg, _ := util.NewErrGroup()

		for _, object := range objects {
			// logrus.Infof("KEY: %s", object.Key)
			eg.Go(DeleteObjectWorker(s3Client, ropts.Bucket, object, &operationTracker))
		}

		// wait on the errgroup and check for error
		err = eg.Wait()
		if err != nil {
			return util.WrapError(err, funcTag, "failed to delete s3 objects in errgroup")
		}

		logrus.Infof("Deleted all objects from %s", opts.S3Key)
	}

	logrus.Infof("%d objects deleted", len(operationTracker))

	return nil
}

// DeleteObjectWorker returns signature of `func() error {}` to satisfy a closure
func DeleteObjectWorker(s3Client *s3.S3, bucket string, object *util.S3Object, tracker *[]*util.S3Object) func() error {
	funcTag := "DeleteObjectWorker"
	return func() error {

		// delete the object from storage permanently
		err := util.DeleteS3Object(s3Client, bucket, object.Key)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to delete object: %s", object.Key))
		}

		// add to tracker
		*tracker = append(*tracker, object)

		// bail
		return nil
	}
}

// // HandleImageDownloadWorker handles async download and processing of images
// // closure is needed to retain context
// func HandleImageDownloadWorker(s3Client *s3.S3, bucket, objectKey string, displayKey string, pageImages *[]*Object) func() error {
// 	return func() error {
// 		funcTag := "HandleImageDownloadWorker"

// 		// download the object to byte slice
// 		objBytes, err := util.DownloadS3Object(s3Client, bucket, objectKey)
// 		if err != nil {
// 			return util.WrapError(err, funcTag, "downloading object bucket")
// 		}

// 		// append to images slice
// 		*pageImages = append(*pageImages, &Object{
// 			Bytes:      objBytes,
// 			Base64:     base64.StdEncoding.EncodeToString(objBytes),
// 			DisplayKey: displayKey,
// 			Key:        objectKey,
// 		})

// 		// logrus.Infof("WORKER: (%d files) %s", len(*pageImages), displayKey)
// 		return nil
// 	}
// }
