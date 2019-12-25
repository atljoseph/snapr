package cli

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"snapr/util"

	"github.com/sirupsen/logrus"
)

// DownloadPage shows when visiting /download
type DownloadPage struct {
	Message string
}

// ServeCmdDownloadHandler is an http handler for downloading files to the work dir from the cli
func ServeCmdDownloadHandler(ropts *RootCmdOptions, opts *ServeCmdOptions) func(w http.ResponseWriter, r *http.Request) {
	funcTag := "ServeCmdBrowseHandler"
	return func(w http.ResponseWriter, r *http.Request) {
		// logrus.Infof("REQUEST (%s): %s, %s, %s", funcTag, r.Method, r.URL, r.RequestURI)

		// only respond to get request (from browser)
		// TODO: Change to POST and use XHR template
		if r.Method == http.MethodGet {

			// get the key/dir from the url
			qp := r.URL.Query()
			qpS3Keys := qp["key"]

			// get the request directory, based on the base dir key provided in the CLI opts
			// default to the s3 Dir provided by cli interface / env vars
			qpS3Key := opts.S3Dir
			qpS3KeyDisplay := ""
			// get the first value in []string from qp slice value
			if len(qpS3Keys) > 0 {
				// get the value and trim off the last "/"
				qpS3KeyDisplay = qpS3Keys[0]
				// if there is a length of string, add a delimiter
				if len(opts.S3Dir) > 0 {
					qpS3Key += util.S3Delimiter
				}
				// pick up the qp
				qpS3Key += qpS3KeyDisplay
			}

			logrus.Infof("KEY: %s, DISPLAY: %s", qpS3Key, qpS3KeyDisplay)

			// get a new s3 client
			awsSesh, _, err := util.NewS3Client(ropts.S3Config)
			if err != nil {
				err = util.WrapError(err, funcTag, "get a new s3 client")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// download the object to byte slice
			objBytes, err := util.DownloadS3Object(awsSesh, ropts.Bucket, qpS3Key)
			if err != nil {
				err = util.WrapError(err, funcTag, "downloading object")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			logrus.Infof("Downloaded: %s", qpS3Key)

			// new path
			newFilePath := filepath.Join(opts.WorkDir, qpS3KeyDisplay)

			// ensure dir exists
			mkdir := filepath.Dir(newFilePath)
			logrus.Infof("Ensuring Directory: %s", mkdir)
			err = os.MkdirAll(mkdir, 0700)
			// err = os.MkdirAll(mkdir, ropts.FileCreateMode)
			if err != nil {
				err = util.WrapError(err, funcTag, "mkdir "+mkdir)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Create new file
			newFile, err := os.Create(newFilePath)
			if err != nil {
				err = util.WrapError(err, funcTag, fmt.Sprintf("could not create new file: %s", newFilePath))
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer newFile.Close()

			// copy the data to the new file
			_, err = newFile.Write(objBytes)
			if err != nil {
				err = fmt.Errorf("could not write bytes to file")
				return
			}

			// logrus.Infof("COPIED: %d", bytesCopied)

			// execute with message
			message := fmt.Sprintf("File Copied: %s", newFilePath)
			serveCmdTempl.ExecuteTemplate(w, "download", &DownloadPage{Message: message})

		}
	}
}
