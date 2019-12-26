package cli

import (
	"fmt"
	"snapr/util"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

// RenameCmdRunE runs the rename command
// it is exported for testing
func RenameCmdRunE(ropts *RootCmdOptions, opts *RenameCmdOptions) error {
	funcTag := "rename"
	// logrus.Infof(funcTag)
	var err error

	// validate required arg
	if len(opts.S3SourceKey) == 0 {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "option `--s3-source-key` is required")
	}

	// validate required arg
	if len(opts.S3DestKey) == 0 {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "option `--s3-dest-key` is required")
	}

	// // validate args are not the same
	// if strings.EqualFold(opts.S3SourceKey, opts.S3DestKey) {
	// 	return util.WrapError(fmt.Errorf("validation error"), funcTag, "cannot rename object to the same key")
	// }

	// get a new aws session
	_, s3Client, err := util.NewS3Client(ropts.S3Config)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get new s3 client")
	}

	// track operated object keys
	var operationTracker []*RenameCmdOperationTracker

	if !opts.IsDir {

		// files
		srcObj := util.S3Object{Key: opts.S3SourceKey}
		destObj := util.S3Object{Key: opts.S3DestKey}

		// check if the objct exists
		exists, err := util.CheckS3ObjectExists(s3Client, ropts.Bucket, srcObj.Key)
		if !exists || err != nil {
			return util.WrapError(fmt.Errorf("validation error"), funcTag, fmt.Sprintf("failed to confirm object existence, or object does not exist in bucket: '%s/%s'", ropts.Bucket, srcObj.Key))
		}
		// logrus.Infof("Object exists: %s", file.Key)

		// rename the object
		err = util.RenameS3Object(s3Client, ropts.Bucket, srcObj.Key, destObj.Key)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to rename object: %s", srcObj.Key))
		}

		operationTracker = append(operationTracker, &RenameCmdOperationTracker{
			Source: &srcObj,
			Dest:   &destObj,
		})
		logrus.Infof("Renamed: %s to %s", srcObj.Key, destObj.Key)
	} else {

		// get all the objects in the bucket
		objects, _, err := util.ListS3ObjectsByKey(s3Client, ropts.Bucket, opts.S3SourceKey, false)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to get a list bucket objects for key: %s", opts.S3SourceKey))
		}

		// open a new go errgroup
		eg, _ := util.NewErrGroup()

		// for every object, we want a worker to change the key
		for _, srcObj := range objects {
			destObj := &util.S3Object{Key: strings.ReplaceAll(srcObj.Key, opts.S3SourceKey, opts.S3DestKey)}
			// logrus.Infof("KEY: %s ==> %s", srcObj.Key, destObj.Key)
			eg.Go(RenameObjectWorker(s3Client, ropts.Bucket, srcObj, destObj, &operationTracker))
		}

		// wait on the errgroup and check for error
		err = eg.Wait()
		if err != nil {
			return util.WrapError(err, funcTag, "failed to rename s3 objects in errgroup")
		}

		logrus.Infof("Renamed all objects from %s to %s", opts.S3SourceKey, opts.S3DestKey)
	}

	logrus.Infof("%d objects renamed", len(operationTracker))

	return nil
}

// RenameObjectWorker returns signature of `func() error {}` to satisfy a closure
func RenameObjectWorker(s3Client *s3.S3, bucket string, srcObj *util.S3Object, destObj *util.S3Object, tracker *[]*RenameCmdOperationTracker) func() error {
	funcTag := "DeleteObjectWorker"
	return func() error {

		// rename the object
		err := util.RenameS3Object(s3Client, bucket, srcObj.Key, destObj.Key)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to rename object: %s", srcObj.Key))
		}

		// add to tracker
		*tracker = append(*tracker, &RenameCmdOperationTracker{
			Source: srcObj,
			Dest:   destObj,
		})

		// bail
		return nil
	}
}
