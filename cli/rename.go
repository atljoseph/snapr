package cli

import (
	"fmt"
	"snapr/util"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

// TODO: serve command - add move/rename capability (GLOB)

// RenameCmdRunE runs the rename command
// it is exported for testing
func RenameCmdRunE(ropts *RootCmdOptions, opts *RenameCmdOptions) error {
	funcTag := "rename"
	// logrus.Infof(funcTag)
	var err error

	// validate required arg
	if len(opts.S3SourceKey) == 0 {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "option `--s3-src-key` is required")
	}

	// validate required arg
	if len(opts.S3DestKey) == 0 {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "option `--s3-dest-key` is required")
	}

	// default dest bucket to current s3 bucket if not already done
	if len(opts.S3DestBucket) == 0 {
		opts.S3DestBucket = ropts.Bucket
	}

	// get a new aws session
	_, s3Client, err := util.NewS3Client(ropts.S3Config)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get new s3 client")
	}

	// set the object acl to "private"
	destAcl := "private"
	// unless set to public
	if opts.IsDestPublic {
		destAcl = "public-read"
	}
	logrus.Infof("With DESTINATION Access ACL: %s", destAcl)

	// track operated object keys
	operationTracker := &[]*RenameCmdOperationTracker{}
	errorTracker := &[]*error{}

	if !opts.SrcIsDir {

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
		err = util.CopyS3Object(s3Client, ropts.Bucket, srcObj.Key, opts.S3DestBucket, destObj.Key, destAcl)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to rename object: %s", srcObj.Key))
		}

		*operationTracker = append(*operationTracker, &RenameCmdOperationTracker{
			Source: &srcObj,
			Dest:   &destObj,
		})
		logrus.Infof("Renamed: %s to %s", srcObj.Key, destObj.Key)
	} else {

		// make sure that it is directory, we add an extra slash
		opts.S3SourceKey = util.EnsureS3DirPath(opts.S3SourceKey)
		opts.S3DestKey = util.EnsureS3DirPath(opts.S3DestKey)

		logrus.Infof("SRC: %s, DEST: %s", opts.S3SourceKey, opts.S3DestKey)

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
			logrus.Infof("KEY: %s ==> %s", srcObj.Key, destObj.Key)
			eg.Go(RenameObjectWorker(s3Client, ropts.Bucket, srcObj, opts.S3DestBucket, destObj, operationTracker, errorTracker, opts.IsCopyOperation, destAcl))
		}

		// wait on the errgroup and check for error
		err = eg.Wait()
		if err != nil {
			return util.WrapError(err, funcTag, "failed to rename s3 objects in errgroup")
		}

		logrus.Infof("Renamed all objects from %s to %s", opts.S3SourceKey, opts.S3DestKey)
	}

	logrus.Infof("%d objects renamed", len(*operationTracker))

	return nil
}

// RenameObjectWorker returns signature of `func() error {}` to satisfy a closure
// TODO: add error tracking *[]*error
func RenameObjectWorker(s3Client *s3.S3, srcBucket string, srcObj *util.S3Object, destBucket string, destObj *util.S3Object, accumulator *[]*RenameCmdOperationTracker, errorAccumulator *[]*error, isCopyOperation bool, acl string) func() error {
	funcTag := "DeleteObjectWorker"
	var err error
	return func() error {

		// same or different buckets?
		differentBuckets := !strings.EqualFold(srcBucket, destBucket)

		// to copy or rename (copy and delete) ?
		if isCopyOperation || differentBuckets {
			// copy the object
			err = util.CopyS3Object(s3Client, srcBucket, srcObj.Key, destBucket, destObj.Key, acl)
			if err != nil {
				err = util.WrapError(err, funcTag, fmt.Sprintf("failed to rename object: %s", srcObj.Key))
				*errorAccumulator = append(*errorAccumulator, &err)
				return err
			}
		} else {
			// rename the object
			err = util.RenameS3Object(s3Client, srcBucket, srcObj.Key, destBucket, destObj.Key, acl)
			if err != nil {
				err = util.WrapError(err, funcTag, fmt.Sprintf("failed to rename object: %s", srcObj.Key))
				*errorAccumulator = append(*errorAccumulator, &err)
				return err
			}
		}

		// add to tracker
		*accumulator = append(*accumulator, &RenameCmdOperationTracker{
			Source: srcObj,
			Dest:   destObj,
		})

		// bail
		return nil
	}
}
