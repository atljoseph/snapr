package cli

import (
	"context"
	"encoding/base64"
	"net/http"
	"path/filepath"
	"snapr/util"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"

	"golang.org/x/sync/errgroup"
)

// BrowsePage is the page in a browser
type BrowsePage struct {
	Folders []Folder
	Files   []Object
	Images  []*Object
}

// Object is a wrapper for an aws object
type Object struct {
	Bytes      []byte
	Base64     string
	Key        string
	DisplayKey string
}

// Folder is a wrapper for an aws folder
type Folder struct {
	Key        string
	DisplayKey string
}

// ServeCmdBrowseHandler is a proving ground right meow
func ServeCmdBrowseHandler(ropts *RootCmdOptions, opts *ServeCmdOptions) func(w http.ResponseWriter, r *http.Request) {
	funcTag := "ServeCmdBrowseHandler"
	return func(w http.ResponseWriter, r *http.Request) {
		// logrus.Infof("REQUEST (%s): %s, %s, %s", funcTag, r.Method, r.URL, r.RequestURI)

		// only respond to get request (from browser)
		if r.Method == http.MethodGet {

			// get the key/dir from the url
			qp := r.URL.Query()
			qpS3SubKeys := qp["dir"]

			// get the request directory, based on the base dir key provided in the CLI opts
			// default to the s3 Dir provided by cli interface / env vars
			qpS3SubKey := opts.S3Dir
			// get the first value in []string from qp slice value
			if len(qpS3SubKeys) > 0 {
				// get the value and trim off the last "/"
				qpVal := qpS3SubKeys[0]
				// if there is a length of string, add a delimiter
				if len(opts.S3Dir) > 0 {
					qpS3SubKey += util.S3Delimiter
				}
				// pick up the qp
				qpS3SubKey += qpVal
				if len(qpVal) > 0 {
					// find out if it ends with a delimiter
					lastCharIsDelimiter := strings.EqualFold(string(qpS3SubKey[len(qpS3SubKey)-1]), util.S3Delimiter)
					// if there is a length of string, add a delimiter
					if !lastCharIsDelimiter {
						qpS3SubKey += util.S3Delimiter
					}
				}
			}
			// logrus.Infof("QP: %s", qpS3SubKey)

			// get a new s3 client
			awsSesh, s3Client, err := util.NewS3Client(ropts.S3Config)
			if err != nil {
				err = util.WrapError(err, funcTag, "get a new s3 client")
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			// get the object list
			objects, commonKeys, err := util.ListS3ObjectsByKey(s3Client, ropts.Bucket, qpS3SubKey, true)
			if err != nil {
				err = util.WrapError(err, funcTag, "get bucket contents info by key")
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			// build the page's images
			p := &BrowsePage{}

			// folders
			// for each sub-directory (common key)
			for _, commonKey := range commonKeys {

				// smash together the cli input s3-dir with the object key
				cliInputDir := opts.S3Dir
				if len(cliInputDir) > 0 {
					cliInputDir += util.S3Delimiter
				}
				// href query param key
				linkKey := strings.ReplaceAll(commonKey, cliInputDir, "")

				// get the last folder in the key
				keysSlice := strings.Split(commonKey, util.S3Delimiter)
				displayKey := keysSlice[len(keysSlice)-2]

				// logrus.Infof("LINK KEY: %s (%s), %s", commonKey, cliInputDir, displayKey)
				p.Folders = append(p.Folders, Folder{Key: linkKey, DisplayKey: displayKey})
			}

			// get a new err group to wait on goroutine group completion
			// and catch errors
			// different than wait groups!
			ctx := context.Background()
			// ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
			eg, ctx := errgroup.WithContext(ctx)

			// files and images
			for _, obj := range objects {

				// smash together the cli input s3-dir with the object key
				cliInputDir := opts.S3Dir
				if len(cliInputDir) > 0 {
					cliInputDir += util.S3Delimiter
				}

				// object key in aws
				displayKey := strings.ReplaceAll(*obj.Key, cliInputDir, "")

				// determine if file or image
				ext := strings.ReplaceAll(filepath.Ext(displayKey), ".", "")

				// is this an image?
				// good compromise for image format determination
				isImage := false
				for _, format := range util.SupportedCaptureFormats() {
					if strings.EqualFold(format, ext) {
						isImage = true
						break
					}
				}

				// if match, put in image slice
				// else file slice
				if isImage {

					// errgroup: closure is needed
					eg.Go(HandleImageDownloadWorker(awsSesh, ropts.Bucket, *obj.Key, displayKey, &p.Images))

				} else {

					// append to files slice
					p.Files = append(p.Files, Object{
						DisplayKey: displayKey,
						Key:        *obj.Key,
					})
				}
			}

			// wait on the errgroup and check for error
			err = eg.Wait()
			if err != nil {
				err = util.WrapError(err, funcTag, "downloading files in errgroup")
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			// reorder the images (due to async gets)
			sort.SliceStable(p.Images, func(a, b int) bool {
				// ascending by "DisplayKey"
				return p.Images[a].DisplayKey < p.Images[b].DisplayKey
			})

			// exec the template and data
			serveCmdTempl.ExecuteTemplate(w, "browse", p)
		}
	}
}

// HandleImageDownloadWorker handles async download and processing of images
// closure is needed to retain context
func HandleImageDownloadWorker(awsSesh *session.Session, bucket, objectKey string, displayKey string, pageImages *[]*Object) func() error {
	return func() error {
		funcTag := "HandleImageDownloadWorker"

		// download the object to byte slice
		objBytes, err := util.DownloadS3Object(awsSesh, bucket, objectKey)
		if err != nil {
			return util.WrapError(err, funcTag, "downloading object bucket")
		}

		// append to images slice
		*pageImages = append(*pageImages, &Object{
			Bytes:      objBytes,
			Base64:     base64.StdEncoding.EncodeToString(objBytes),
			DisplayKey: displayKey,
			Key:        objectKey,
		})

		// logrus.Infof("WORKER: (%d files) %s", len(*pageImages), displayKey)
		return nil
	}
}
