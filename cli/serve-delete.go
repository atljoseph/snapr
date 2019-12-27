package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"snapr/util"

	"github.com/sirupsen/logrus"
)

// DeleteRequest is the request body for a download request from the browser
type DeleteRequest struct {
	Key   string `json:"key"`
	IsDir bool   `json:"is_dir"`
}

// DeleteResponse is sent back to the requester in json format
type DeleteResponse struct {
	Message string `json:"message"`
}

// ServeCmdDeleteHandler is an http handler for downloading files to the work dir from the cli
func ServeCmdDeleteHandler(ropts *RootCmdOptions, opts *ServeCmdOptions) func(w http.ResponseWriter, r *http.Request) {
	funcTag := "ServeCmdDeleteHandler"
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
		var body DeleteRequest
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

		logrus.Infof("KEY: %s", body.Key)

		// build & fire the cli command
		cmdArgs := &DeleteCmdOptions{
			S3Key: body.Key,
			IsDir: body.IsDir,
		}
		// check the error
		err = DeleteCmdRunE(rootCmdOpts, cmdArgs)
		if err != nil {
			err = fmt.Errorf("failed running delete command with opts: %+v: %s", cmdArgs, err)
			logrus.Warnf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// success message
		resp := DeleteResponse{
			Message: fmt.Sprintf("Object Deleted: %s", body.Key),
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
