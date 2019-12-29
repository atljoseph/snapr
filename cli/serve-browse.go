package cli

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path/filepath"
	"snapr/util"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

// BrowsePage is the page in a browser
type BrowsePage struct {
	Key     string
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
		qpS3Keys := qp["dir"]

		// get the request directory, based on the base dir key provided in the CLI opts
		// default to the s3 Dir provided by cli interface / env vars
		s3Key := ""
		// get the first value in []string from qp slice value
		if len(qpS3Keys) > 0 {
			// get the value and trim off the last "/"
			s3Key += qpS3Keys[0]
			// ensure the ending char
			s3Key = util.EnsureS3DirPath(s3Key)
		}
		// logrus.Infof("QP: %s", qpS3SubKey)

		// get a new s3 client
		_, s3Client, err := util.NewS3Client(ropts.S3Config)
		if err != nil {
			err = util.WrapError(err, funcTag, "failed to get a new s3 client")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// build the page's images
		p := &BrowsePage{
			Key:    s3Key,
			Images: []*util.S3Object{},
		}

		// get the object list
		var objects []*util.S3Object
		objects, p.Folders, err = util.ListS3ObjectsByKey(s3Client, ropts.Bucket, s3Key, true)
		if err != nil {
			err = util.WrapError(err, funcTag, "failed to get bucket contents info by key")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// open a new go errgroup for a parrallel operation
		eg, _ := util.NewErrGroup()

		// files and images
		for _, obj := range objects {

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
				eg.Go(HandleImageDownloadWorker(s3Client, ropts.Bucket, obj, &p.Images, true))

			} else {

				// append to files slice
				p.Files = append(p.Files, obj)
			}
		}

		// wait on the errgroup and check for error
		err = eg.Wait()
		if err != nil {
			err = util.WrapError(err, funcTag, "failed to download files in parallel")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// reorder the images (due to async gets)
		sort.SliceStable(p.Images, func(a, b int) bool {
			// ascending by filename
			fileA := filepath.Base(p.Images[a].Key)
			fileB := filepath.Base(p.Images[b].Key)
			return fileA < fileB
		})

		// exec the template and data
		serveCmdTempl.ExecuteTemplate(w, "browse", p)
	}
}

// HandleImageDownloadWorker handles async download and conversion of images
func HandleImageDownloadWorker(s3Client *s3.S3, bucket string, obj *util.S3Object, accumulator *[]*util.S3Object, convertBase64 bool) func() error {
	funcTag := "HandleImageDownloadWorker"
	var err error
	return func() error {

		// download the object to byte slice
		obj.Bytes, err = util.DownloadS3Object(s3Client, bucket, obj.Key)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("failed to download bucket object: %s", obj.Key))
		}

		// if specified, conver to base64, too
		if convertBase64 {
			obj.Base64 = base64.StdEncoding.EncodeToString(obj.Bytes)
		}

		// append to images slice
		*accumulator = append(*accumulator, obj)

		// logrus.Infof("WORKER: (%d files) %s", len(*pageImages), displayKey)
		return nil
	}
}
