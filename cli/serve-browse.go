package cli

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"snapr/util"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

// BrowsePage is the page in a browser
type BrowsePage struct {
	Folders []*util.S3Directory
	Files   []*util.S3Object
	Images  []*util.S3Object
}

// ServeCmdBrowseHandler is a handler that serve up an s3 bucket in file folders
func ServeCmdBrowseHandler(ropts *RootCmdOptions, opts *ServeCmdOptions) func(w http.ResponseWriter, r *http.Request) {
	funcTag := "ServeCmdBrowseHandler"
	var err error
	return func(w http.ResponseWriter, r *http.Request) {
		// logrus.Infof("REQUEST (%s): %s, %s, %s", funcTag, r.Method, r.URL, r.RequestURI)

		// only respond to post request (from browser)
		if r.Method != http.MethodGet {
			err = fmt.Errorf("incorrect method for this endpoint: %s", r.Method)
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

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
		_, s3Client, err := util.NewS3Client(ropts.S3Config)
		if err != nil {
			err = util.WrapError(err, funcTag, "get a new s3 client")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// get the object list
		objects, dirs, err := util.ListS3ObjectsByKey(s3Client, ropts.Bucket, qpS3SubKey, true)
		if err != nil {
			err = util.WrapError(err, funcTag, "get bucket contents info by key")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// build the page's images
		p := &BrowsePage{}

		// folders
		// for each sub-directory (common key)
		for _, dir := range dirs {

			// smash together the cli input s3-dir with the object key
			cliInputDir := opts.S3Dir
			if len(cliInputDir) > 0 {
				cliInputDir += util.S3Delimiter
			}
			// href query param key
			linkKey := strings.ReplaceAll(dir.Key, cliInputDir, "")

			// get the last folder in the key
			keysSlice := strings.Split(dir.Key, util.S3Delimiter)
			displayKey := keysSlice[len(keysSlice)-2]

			// logrus.Infof("LINK KEY: %s (%s), %s", commonKey, cliInputDir, displayKey)
			p.Folders = append(p.Folders, &util.S3Directory{Key: linkKey, DisplayKey: displayKey})
		}

		// open a new go errgroup for a parrallel operation
		eg, _ := util.NewErrGroup()

		// files and images
		for _, obj := range objects {

			// smash together the cli input s3-dir with the object key
			cliInputDir := opts.S3Dir
			if len(cliInputDir) > 0 {
				cliInputDir += util.S3Delimiter
			}

			// object key in aws
			obj.DisplayKey = strings.ReplaceAll(obj.Key, cliInputDir, "")

			// is this an image?
			// good compromise for image format determination
			isImage := false
			for _, format := range util.SupportedCaptureFormats() {
				if strings.EqualFold(format, obj.Extension) {
					isImage = true
					break
				}
			}

			// if match, put in image slice
			// else file slice
			if isImage {

				// errgroup: closure is needed
				eg.Go(HandleImageDownloadWorker(s3Client, ropts.Bucket, obj, &p.Images))

			} else {

				// append to files slice
				p.Files = append(p.Files, obj)
			}
		}

		// wait on the errgroup and check for error
		err = eg.Wait()
		if err != nil {
			err = util.WrapError(err, funcTag, "downloading files in errgroup")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
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

// HandleImageDownloadWorker handles async download and processing of images
// closure is needed to retain context
func HandleImageDownloadWorker(s3Client *s3.S3, bucket string, obj *util.S3Object, pageImages *[]*util.S3Object) func() error {
	funcTag := "HandleImageDownloadWorker"
	return func() error {

		// download the object to byte slice
		objBytes, err := util.DownloadS3Object(s3Client, bucket, obj.Key)
		if err != nil {
			return util.WrapError(err, funcTag, "downloading object bucket")
		}

		// append to images slice
		obj.Base64 = base64.StdEncoding.EncodeToString(objBytes)
		*pageImages = append(*pageImages, obj)

		// logrus.Infof("WORKER: (%d files) %s", len(*pageImages), displayKey)
		return nil
	}
}
