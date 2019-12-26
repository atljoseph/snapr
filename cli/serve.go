package cli

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"snapr/util"

	"github.com/sirupsen/logrus"
)

// ServeCmdRunE runs the serve command
// it is exported for testing
func ServeCmdRunE(ropts *RootCmdOptions, opts *ServeCmdOptions) error {
	funcTag := "ServeCmdRunE"
	// logrus.Infof(funcTag)
	var err error

	// default the work dir to the pwd
	if len(opts.WorkDir) == 0 {
		opts.WorkDir, err = os.Getwd()
		if err != nil {
			return util.WrapError(err, funcTag, "cannot get pwd for WorkDir")
		}
	} else {
		opts.WorkDir, err = filepath.Abs(opts.WorkDir)
		if err != nil {
			return util.WrapError(err, funcTag, "cannot get abs path for WorkDir")
		}
	}

	// parse templates
	serveCmdTempl, err = ParseTemplates()
	if err != nil {
		return util.WrapError(err, funcTag, "parsing templates")
	}
	logrus.Infof("Templates parsed")

	// set up handlers
	http.HandleFunc("/browse", ServeCmdBrowseHandler(ropts, opts))
	http.HandleFunc("/download", ServeCmdDownloadHandler(ropts, opts))
	http.HandleFunc("/delete", ServeCmdDeleteHandler(ropts, opts))
	http.HandleFunc("/rename", ServeCmdRenameHandler(ropts, opts))
	http.HandleFunc("/", ServeCmd404NotFoundHandler(ropts, opts))
	logrus.Infof("Handlers registered")

	// host and port
	hostNPort := fmt.Sprintf("%s:%d", "localhost", opts.Port)
	logrus.Warnf("Go to `http://%s/browse` in your browser ...", hostNPort)

	// listen and serve
	// blocking for now
	// go (func() {
	err = http.ListenAndServe(hostNPort, nil)
	if err != nil {
		return util.WrapError(err, funcTag, "serving content")
	}
	logrus.Infof("Shutting down")
	// })()

	return nil
}

// ServeCmd404NotFoundHandler catches all 404s
func ServeCmd404NotFoundHandler(ropts *RootCmdOptions, opts *ServeCmdOptions) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Warnf("404 NotFound: %s, %s, %s", r.Method, r.URL, r.RequestURI)
		http.Error(w, "Invalid Request. Endpoint not defined.", http.StatusNotFound)
	}
}
