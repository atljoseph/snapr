package cli

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"snapr/util"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/sync/errgroup"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

// ProcessedImage ties together all we need
// in order to upload to a bucket
type ProcessedImage struct {
	Bytes []byte
	Key   string
	Size  int
}

// ProcessCmdRunE runs the download command
// it is exported for testing
func ProcessCmdRunE(ropts *RootCmdOptions, opts *ProcessCmdOptions) error {
	funcTag := "process"
	// logrus.Infof(funcTag)
	// var err error

	// ------  DEFAULT -----------------------------------

	// set the object acl to "private"
	acl := "private"
	// unless set to public
	if processCmdOpts.Public {
		acl = "public-read"
	}
	logrus.Infof("With Access ACL: %s", acl)

	// ------  VALIDATE -----------------------------------

	// validate the in dir
	if len(opts.S3InKey) == 0 {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "must provide an s3 key for input directory")
	}

	// validate the out dir
	if len(opts.S3OutKey) == 0 {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "must provide an s3 key for output directory")
	}

	// validate that in and out are not the same
	if strings.EqualFold(opts.S3InKey, opts.S3OutKey) {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, fmt.Sprintf("input and output keys cannot be the same: '%s' vs '%s'", opts.S3InKey, opts.S3OutKey))
	}

	logrus.Infof("IN: %s, OUT: %s, SIZES: %d", opts.S3InKey, opts.S3OutKey, opts.OutSizes)

	// ------  LIST OBJECTS -----------------------------------

	// get a new aws session
	_, s3Client, err := util.NewS3Client(ropts.S3Config)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get new s3 client")
	}

	// list all files recursively
	// for the directory to process from ("originals")
	objects, _, err := util.ListS3ObjectsByKey(s3Client, ropts.Bucket, opts.S3InKey, false)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get s3 object list")
	}

	logrus.Infof("SEARCH RESULTS: %d", len(objects))

	// ------  CLEANUP OUTPUT DIR -----------------------------------

	// build & fire the cli command
	cmdArgs := &DeleteCmdOptions{
		S3Key: opts.S3OutKey,
		IsDir: true,
	}
	// check the error
	err = DeleteCmdRunE(rootCmdOpts, cmdArgs)
	if err != nil {
		return fmt.Errorf("failed running delete command with opts: %+v: %s", cmdArgs, err)
	}

	logrus.Infof("DELETED: %s", opts.S3OutKey)

	// ------  FILTER -----------------------------------

	// errgroupFuncs
	var efs []func() error
	var processed = &[]*string{}
	var errors = &[]*error{}
	matches := 0

	for _, obj := range objects {

		// if partial path matches
		// include it as a match
		// to keep "c" from matching from "candler"
		isMatch := false
		partials := strings.Split(obj.Key, util.S3Delimiter)
		for idx := range partials {
			slice := partials[0:idx]
			if strings.EqualFold(opts.S3InKey, strings.Join(slice, util.S3Delimiter)) {
				isMatch = true
			}
		}

		// is this an image?
		// good compromise for image format determination
		isImage := false
		if isMatch {
			for _, format := range util.SupportedCaptureFormats() {
				if strings.EqualFold(format, obj.Extension) {
					isImage = true
					break
				}
			}
		}

		// if match, put in image slice
		// else file slice
		if isImage && isMatch {
			// incrment
			matches++
			// process all variations
			efs = append(efs, HandleImageProcessWorker(s3Client, ropts.Bucket, obj.Key, opts.S3InKey, opts.S3OutKey, acl, opts.OutSizes, processed, errors))

		}
	}

	// ------  DOWNLOAD ORIGINALS -----------------------------------
	// ------  PROCESS OUTPUTS -----------------------------------
	// ------  UPLOAD OUTPUTS -----------------------------------

	// open a new go errgroup for a parrallel operation
	// batched parallelism
	var eg *errgroup.Group
	counter := 0
	maxPer := 5
	leftovers := maxPer

	logrus.Infof("STARTING")

	// files and images
	for leftovers > 0 {

		// index
		start := counter * maxPer
		end := start + maxPer
		leftovers = len(efs) - len(*errors) - len(*processed)
		if maxPer > leftovers {
			end = start + leftovers
		}
		logrus.Infof("Leftovers %d, High Water %d, Errors %d, Done %d, Start %d, End %d", leftovers, len(efs), len(*errors), len(*processed), start, end)

		// reup the err group
		eg, _ = util.NewErrGroup()

		// upload with worker in errgroup
		// each variation is an object
		// each object has a goroutine
		for i, ef := range efs[start:end] {
			logrus.Infof("Image %d", start+i+1)
			eg.Go(ef)
		}

		// wait on the errgroup and check for error
		err = eg.Wait()
		if err != nil {
			return util.WrapError(err, funcTag, "failed to download files in parallel")
		}

		// next batch
		counter++
	}

	logrus.Infof("PROCESSED: %d", len(*processed))

	return nil
}

// HandleImageProcessWorker handles async processing of images
func HandleImageProcessWorker(s3Client *s3.S3, bucket, origKey, inKey, outKey, acl string, sizes []int, accumulator *[]*string, errorAccumulator *[]*error) func() error {
	funcTag := "HandleImageProcessWorker"
	return func() error {
		logrus.Infof("Starting: (%d) %s", sizes, origKey)

		// logrus.Infof("WORKER: %s", key)
		inBuf, err := util.DownloadS3Object(s3Client, bucket, origKey)
		if err != nil {
			err = util.WrapError(err, funcTag, fmt.Sprintf("failed to download bucket object: %s", inKey))
			*errorAccumulator = append(*errorAccumulator, &err)
			return err
		}

		// convert bytes to image.Image
		img, _, err := image.Decode(bytes.NewReader(inBuf))
		if err != nil {
			err = util.WrapError(err, funcTag, fmt.Sprintf("failed to decode bytes: %s", origKey))
			*errorAccumulator = append(*errorAccumulator, &err)
			return err
		}

		for _, size := range sizes {

			// resize
			imgResized := imaging.Resize(img, size, 0, imaging.Lanczos)

			// convert back to bytes
			outBuf := new(bytes.Buffer)
			err = jpeg.Encode(outBuf, imgResized, nil)

			// key / directory for sizes
			sizeOutKey := strings.ReplaceAll(origKey, inKey, strconv.Itoa(size))
			fullOutKey := util.JoinS3Path(outKey, sizeOutKey)

			// send to AWS
			_, err = util.WriteS3Bytes(s3Client, bucket, acl, fullOutKey, outBuf.Bytes())
			if err != nil {
				err = util.WrapError(err, funcTag, "failed to send bytes to s3")
				*errorAccumulator = append(*errorAccumulator, &err)
				return err
			}
		}

		// append to images slice
		*accumulator = append(*accumulator, &origKey)

		logrus.Infof("WORKER DONE: %s", origKey)
		return nil
	}
}
