package cli

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"path/filepath"
	"snapr/util"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/pieterclaerhout/go-waitgroup"
	"github.com/sirupsen/logrus"
)

// ProcessCmdRunE runs the process command
// it is exported for testing
func ProcessCmdRunE(ropts *RootCmdOptions, opts *ProcessCmdOptions) error {
	funcTag := "process"
	// logrus.Infof(funcTag)
	// var err error

	// ------  DEFAULT -----------------------------------

	// set the object acl to "private"
	acl := "private"
	// unless set to public
	if opts.IsDestPublic {
		acl = "public-read"
	}
	logrus.Infof("With Access ACL: %s", acl)

	// default to RebuildNew if neither is set
	if !opts.RebuildAll && !opts.RebuildNew {
		opts.RebuildAll = false
		opts.RebuildNew = true
	}

	// default to RebuildNew if both are set
	if opts.RebuildAll && opts.RebuildNew {
		opts.RebuildAll = false
		opts.RebuildNew = true
	}

	// ensure ending dir slash for all these
	opts.S3SrcKey = util.EnsureS3DirPath(opts.S3SrcKey)
	opts.S3DestKey = util.EnsureS3DirPath(opts.S3DestKey)

	// ------  VALIDATE -----------------------------------

	// validate the in dir
	if len(opts.S3SrcKey) == 0 {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "must provide a value for `--s3-src-key`")
	}

	// validate the out dir
	if len(opts.S3DestKey) == 0 {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "must provide a value for `--s3-dest-key`")
	}

	// validate that in and out are not the same
	if strings.EqualFold(opts.S3SrcKey, opts.S3DestKey) {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, fmt.Sprintf("input and output keys cannot be the same: '%s' vs '%s'", opts.S3SrcKey, opts.S3DestKey))
	}

	logrus.Infof("IN: %s, OUT: %s, SIZES: %d", opts.S3SrcKey, opts.S3DestKey, opts.Sizes)

	// ------  LIST OBJECTS -----------------------------------

	// get a new aws session
	_, s3Client, err := util.NewS3Client(ropts.S3Config)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get new s3 client")
	}

	// list all SOURCE files recursively
	// for the directory to process from ("originals")
	srcObjects, _, err := util.ListS3ObjectsByKey(s3Client, ropts.Bucket, opts.S3SrcKey, false)
	if err != nil {
		return util.WrapError(err, funcTag, "failed to get src s3 object list")
	}

	logrus.Infof("SOURCE OBJECTS: %d", len(srcObjects))

	// ------  CLEANUP OUTPUT DIR AND GET LIST TO REBUILD -----------------------------------

	objectsToProcess := &[]*util.S3Object{}

	// if rebuilding all files, then remove the entire destination directory
	if opts.RebuildAll {
		// build & fire the cli command
		cmdArgs := &DeleteCmdOptions{
			S3Key: opts.S3DestKey,
			IsDir: true,
		}
		// check the error
		err = DeleteCmdRunE(rootCmdOpts, cmdArgs)
		if err != nil {
			return fmt.Errorf("failed running delete command with opts: %+v: %s", cmdArgs, err)
		}

		logrus.Infof("DELETED: %s", opts.S3DestKey)

		// set all objects in the path to be processed
		objectsToProcess = &srcObjects
	} else {
		// list all DEST files recursively
		// for the directory to process to ("processed")
		destObjects, _, err := util.ListS3ObjectsByKey(s3Client, ropts.Bucket, opts.S3DestKey, false)
		if err != nil {
			return util.WrapError(err, funcTag, "failed to get s3 dest object list")
		}

		// filter objects to process
		// only process new files
		// based on the processed output, what do we expect to see in the originals dir?
		var expects []string
		for _, dobj := range destObjects {

			// strip the base dest key
			destPath := strings.Replace(dobj.Key, util.EnsureS3DirPath(opts.S3DestKey), "", 1)

			for _, size := range opts.Sizes {

				// get the size string
				sizeStr := strconv.Itoa(size)

				// if starts with the specific size prefix
				if strings.Index(destPath, sizeStr) == 0 {
					// strip the size, too
					destPathSize := strings.Replace(destPath, util.EnsureS3DirPath(sizeStr), "", 1)

					// if expects does not already contain, append
					contained := false
					for _, e := range expects {
						if strings.EqualFold(e, destPathSize) {
							contained = true
						}
					}

					if !contained {
						expects = append(expects, destPathSize)
					}
				}

			}
		}

		logrus.Infof("EXPECTING %d IN %s", len(expects), opts.S3SrcKey)

		// look through every original and find what is there that is not in the other place
		for _, sobj := range srcObjects {

			// string the base original path from the src
			path := strings.Replace(sobj.Key, util.EnsureS3DirPath(opts.S3SrcKey), "", 1)

			// look through the expects and find a match
			found := false
			for _, expect := range expects {
				if strings.EqualFold(expect, path) {
					// logrus.Infof("CHECK: '%s'", expect)
					found = true
				}
			}

			// process if not found
			if !found {

				// is this an image?
				// good compromise for image format determination
				isImage := false
				for _, format := range util.SupportedCaptureFormats() {
					if strings.EqualFold(format, strings.ReplaceAll(filepath.Ext(path), ".", "")) {
						isImage = true
						break
					}
				}

				// if image, add it to list for processing
				if isImage {
					logrus.Infof("NEW: '%s'", path)
					*objectsToProcess = append(*objectsToProcess, sobj)
				}
			}
		}
	}

	logrus.Infof("TO PROCESS: %d", len(*objectsToProcess))

	// ------  FILTER FOR IMAGES -----------------------------------

	// waitGroupFuncs
	imagesToProcess := []*util.S3Object{}

	// filter to images only
	for _, obj := range *objectsToProcess {

		// if partial path matches
		// include it as a match
		// to keep "c" from matching from "candler"
		isMatch := false
		partials := strings.Split(obj.Key, util.S3Delimiter)
		for idx := range partials {
			slice := partials[0:idx]
			dirToMatch := util.EnsureS3DirPath(strings.Join(slice, util.S3Delimiter))
			// logrus.Infof("%s ?? %s", opts.S3SrcKey, util.EnsureS3DirPath(strings.Join(slice, util.S3Delimiter)))
			if strings.EqualFold(opts.S3SrcKey, dirToMatch) {
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
			imagesToProcess = append(imagesToProcess, obj)
		}
	}

	// ------ FIRE WAITGROUP -----------------------------------

	// open a new wait group with a maximum number of concurrent workers
	wg := waitgroup.NewWaitGroup(5)

	// track processed and errors
	processed := &[]*string{}
	errors := &[]*error{}

	// loop through all objects and spawn goroutines to wait for
	for _, img := range imagesToProcess {

		// block adding until the next worker has finished
		wg.BlockAdd()

		// on a separate goroutine, do something asyncronous
		// download, process, upload
		go func(origFullKey string, accumulator *[]*string, errorAccumulator *[]*error) {
			funcTag := "ProcessImageWorker"
			defer wg.Done()

			logrus.Infof("WORK: (%d) %s", opts.Sizes, origFullKey)

			// ------  DOWNLOAD ORIGINAL -----------------------------------
			inBuf, err := util.DownloadS3Object(s3Client, ropts.Bucket, origFullKey)
			if err != nil {
				err = util.WrapError(err, funcTag, fmt.Sprintf("failed to download bucket object: %s", opts.S3SrcKey))
				logrus.Warnf(err.Error())
				*errorAccumulator = append(*errorAccumulator, &err)
			}

			// convert bytes to image.Image
			img, _, err := image.Decode(bytes.NewReader(inBuf))
			if err != nil {
				err = util.WrapError(err, funcTag, fmt.Sprintf("failed to decode bytes: %s", origFullKey))
				logrus.Warnf(err.Error())
				*errorAccumulator = append(*errorAccumulator, &err)
			}

			// ------  PROCESS & UPLOAD OUTPUTS -----------------------------------

			// build the output objects
			var outputImages []*ProcessedImage
			for _, size := range opts.Sizes {

				// key / directory for sizes
				// replace the inDirKey with the size, then tack on the outDirKey
				sizeOutKey := strings.ReplaceAll(origFullKey, util.EnsureS3DirPath(opts.S3SrcKey), util.EnsureS3DirPath(strconv.Itoa(size)))
				fullOutKey := util.JoinS3Path(opts.S3DestKey, sizeOutKey)

				// append to list of output images
				outputImages = append(outputImages, &ProcessedImage{
					Size: size,
					Key:  fullOutKey,
				})
			}

			// process and upload
			for _, oi := range outputImages {

				// resize
				imgResized := imaging.Resize(img, oi.Size, 0, imaging.Lanczos)

				// convert back to bytes
				oi.Buffer = new(bytes.Buffer)
				err = jpeg.Encode(oi.Buffer, imgResized, nil)
				oi.Bytes = oi.Buffer.Bytes()

				// send to AWS
				_, err = util.WriteS3Bytes(s3Client, ropts.Bucket, acl, oi.Key, oi.Bytes)
				if err != nil {
					err = util.WrapError(err, funcTag, "failed to send bytes to s3")
					logrus.Warnf(err.Error())
					*errorAccumulator = append(*errorAccumulator, &err)
				}

				logrus.Infof("RESIZED: %s", oi.Key)
			}

			// append to images slice
			*accumulator = append(*accumulator, &origFullKey)

			logrus.Infof("DONE: (%d) %s", opts.Sizes, origFullKey)

			// we need these injected here
		}(img.Key, processed, errors)
	}

	// wait on everything to complete
	wg.Wait()

	logrus.Infof("PROCESSED: %d", len(*processed))

	return nil
}

// ProcessedImage ties together all we need
// in order to upload to a bucket
type ProcessedImage struct {
	Bytes  []byte
	Buffer *bytes.Buffer
	Key    string
	Size   int
}
