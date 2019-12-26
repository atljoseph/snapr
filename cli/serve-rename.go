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

		// get the request directory, based on the base dir key provided in the CLI opts
		// default to the s3 Dir provided by cli interface / env vars
		// if there is a length of string, add a delimiter
		// add the key from request
		srcKey := opts.S3Dir
		if len(srcKey) > 0 {
			srcKey += util.S3Delimiter
		}
		srcKey += body.SrcKey
		// if len(srcKey) > 0 {
		// 	srcKey += util.S3Delimiter
		// }

		// same for the dest key
		destKey := opts.S3Dir
		if len(destKey) > 0 {
			destKey += util.S3Delimiter
		}
		destKey += body.DestKey
		// if len(destKey) > 0 {
		// 	destKey += util.S3Delimiter
		// }

		logrus.Infof("SRC: %s, DEST: %s", srcKey, destKey)

		// build & fire the cli command
		cmdArgs := &RenameCmdOptions{
			S3SourceKey: srcKey,
			S3DestKey:   destKey,
			IsDir:       body.IsDir,
		}
		// check the error
		err = RenameCmdRunE(rootCmdOpts, cmdArgs)
		if err != nil {
			err = fmt.Errorf("failed running rename command with opts: %+v: %s", cmdArgs, err)
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// return success with message
		resp := RenameResponse{
			Message: fmt.Sprintf("Object(s) Renamed: %s to %s", srcKey, destKey),
		}
		err = json.NewEncoder(w).Encode(&resp)
		if err != nil {
			err = fmt.Errorf("could not encode response")
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
