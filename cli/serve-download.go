package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"snapr/util"

	"github.com/sirupsen/logrus"
)

// DownloadRequest is the request body for a download request from the browser
type DownloadRequest struct {
	Key string `json:"key"`
}

// DownloadResponse is sent back to the requester in json format
type DownloadResponse struct {
	Message string `json:"message"`
}

// ServeCmdDownloadHandler is an http handler for downloading files to the work dir from the cli
func ServeCmdDownloadHandler(ropts *RootCmdOptions, opts *ServeCmdOptions) func(w http.ResponseWriter, r *http.Request) {
	funcTag := "ServeCmdBrowseHandler"
	var err error
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("REQUEST (%s): %s, %s, %s", funcTag, r.Method, r.URL, r.RequestURI)

		// only respond to post request (from browser)
		if r.Method != http.MethodPost {
			err = fmt.Errorf("incorrect method for this endpoint: %s", r.Method)
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// decode the request body
		var body DownloadRequest
		err = json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			err = util.WrapError(fmt.Errorf("validation error"), funcTag, "failed to parse request body")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// get the key/dir
		// validate it's length / existance
		if len(body.Key) == 0 {
			err = util.WrapError(fmt.Errorf("validation error"), funcTag, "no `key` provided in body")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// get the request directory, based on the base dir key provided in the CLI opts
		// default to the s3 Dir provided by cli interface / env vars
		s3Key := opts.S3Dir

		// if there is a length of string, add a delimiter
		if len(opts.S3Dir) > 0 {
			s3Key += util.S3Delimiter
		}

		// get the value and trim off the last "/"
		s3KeyDisplay := body.Key

		// pick up the qp
		s3Key += s3KeyDisplay
		// get the first value in []string from qp slice value

		logrus.Infof("KEY: %s, DISPLAY: %s", s3Key, s3KeyDisplay)

		// get a new s3 client
		_, s3Client, err := util.NewS3Client(ropts.S3Config)
		if err != nil {
			err = util.WrapError(err, funcTag, "failed to get a new s3 client")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// download the object to byte slice
		objBytes, err := util.DownloadS3Object(s3Client, ropts.Bucket, s3Key)
		if err != nil {
			err = util.WrapError(err, funcTag, "failed to download object: %s")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		logrus.Infof("Downloaded: %s", s3Key)

		// new path
		newFilePath := filepath.Join(opts.WorkDir, s3KeyDisplay)

		// ensure dir exists
		mkdir := filepath.Dir(newFilePath)
		logrus.Infof("Ensuring Directory: %s", mkdir)
		err = os.MkdirAll(mkdir, 0700)
		// err = os.MkdirAll(mkdir, ropts.FileCreateMode)
		if err != nil {
			err = util.WrapError(err, funcTag, fmt.Sprintf("falied to mkdir: %s", mkdir))
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Create new file
		newFile, err := os.Create(newFilePath)
		if err != nil {
			err = util.WrapError(err, funcTag, fmt.Sprintf("failed to create new file: %s", newFilePath))
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer newFile.Close()

		// copy the data to the new file
		_, err = newFile.Write(objBytes)
		if err != nil {
			err = fmt.Errorf("failed to write bytes to file")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// logrus.Infof("COPIED: %d", bytesCopied)

		// return success with message
		resp := DownloadResponse{
			Message: fmt.Sprintf("Object Copied to Disk: %s", newFilePath),
		}
		err = json.NewEncoder(w).Encode(&resp)
		if err != nil {
			err = fmt.Errorf("failed to encode response")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
