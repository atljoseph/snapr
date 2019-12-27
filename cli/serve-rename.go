package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"snapr/util"

	"github.com/sirupsen/logrus"
)

// RenameRequest is the request body for a download request from the browser
type RenameRequest struct {
	SrcKey  string `json:"src_key"`
	DestKey string `json:"dest_key"`
	IsDir   bool   `json:"is_dir"`
	Copy    bool   `json:"copy"`
}

// RenameResponse is sent back to the requester in json format
type RenameResponse struct {
	Message string `json:"message"`
}

// ServeCmdRenameHandler is an http handler for downloading files to the work dir from the cli
func ServeCmdRenameHandler(ropts *RootCmdOptions, opts *ServeCmdOptions) func(w http.ResponseWriter, r *http.Request) {
	funcTag := "ServeCmdRenameHandler"
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
		var body RenameRequest
		err = json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			err = util.WrapError(fmt.Errorf("validation error"), funcTag, "failed to parse request body")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// validate source key
		if len(body.SrcKey) == 0 {
			err = util.WrapError(fmt.Errorf("validation error"), funcTag, "no `src_key` provided in body")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// validate dest key
		if len(body.DestKey) == 0 {
			err = util.WrapError(fmt.Errorf("validation error"), funcTag, "no `dest_key` provided in body")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// build & fire the cli command
		cmdArgs := &RenameCmdOptions{
			S3SourceKey:     body.SrcKey,
			S3DestKey:       body.DestKey,
			SrcIsDir:        body.IsDir,
			IsCopyOperation: body.Copy,
		}
		// check the error
		err = RenameCmdRunE(rootCmdOpts, cmdArgs)
		if err != nil {
			err = fmt.Errorf("failed running rename command with opts: %+v: %s", cmdArgs, err)
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// success message
		resp := RenameResponse{
			Message: fmt.Sprintf("Object(s) Renamed: %s to %s", body.SrcKey, body.DestKey),
		}

		// success message
		err = json.NewEncoder(w).Encode(&resp)
		if err != nil {
			err = fmt.Errorf("could not encode response")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
