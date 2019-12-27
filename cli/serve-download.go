package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"snapr/util"

	"github.com/sirupsen/logrus"
)

// DownloadRequest is the request body for a download request from the browser
type DownloadRequest struct {
	Key   string `json:"key"`
	IsDir bool   `json:"is_dir"`
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

		// build & fire the cli command
		cmdArgs := &DownloadCmdOptions{
			S3Key: body.Key,
			IsDir: body.IsDir,
		}
		// check the error
		err = DownloadCmdRunE(rootCmdOpts, cmdArgs)
		if err != nil {
			err = fmt.Errorf("failed running download command with opts: %+v: %s", cmdArgs, err)
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// logrus.Infof("COPIED: %d", bytesCopied)

		// success message
		resp := DownloadResponse{
			Message: fmt.Sprintf("Object Downloaded: %s", filepath.Join(cmdArgs.OutDir, cmdArgs.S3Key)),
		}

		// write the response
		err = json.NewEncoder(w).Encode(&resp)
		if err != nil {
			err = fmt.Errorf("failed to encode response")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
