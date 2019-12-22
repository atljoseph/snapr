package cli

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"snapr/util"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ServeCmdOptions options
type ServeCmdOptions struct {
	WorkDir string
	S3Dir   string
	Port    int
	Host    string
	Formats []string
}

// serve command
var (
	serveCmdOpts = &ServeCmdOptions{}
	serveCmd     = &cobra.Command{
		Use:   "serve",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			serveCmdOpts = serveCmdTransformPositionalArgs(args, serveCmdOpts)
			return ServeCmdRunE(rootCmdOpts, serveCmdOpts)
		},
	}
)

// serveCmdTransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func serveCmdTransformPositionalArgs(args []string, opts *ServeCmdOptions) *ServeCmdOptions {
	// if len(args) > 0 {
	// // can use env vars, too!
	// 	opts.Something = args[0]
	// }
	return opts
}

func init() {
	// add command to root
	rootCmd.AddCommand(serveCmd)

	// this is appended to `dir`if set
	serveCmd.Flags().StringVar(&serveCmdOpts.S3Dir,
		"s3-dir", util.EnvVarString("SERVE_S3_DIR", ""),
		"(Optional) Base S3 Directory Key to browse")

	// this is where the files get written to
	// default to calling user's home directory
	// TODO: default below
	serveCmd.Flags().StringVar(&serveCmdOpts.WorkDir,
		"dir", util.EnvVarString("SERVE_DIR", ""),
		"(Recommended) This will eventually be the Download and Upload directory")

	// file override ... optional
	// TODO: default below
	serveCmd.Flags().IntVar(&serveCmdOpts.Port,
		"port", util.EnvVarInt("SERVE_PORT", 8080),
		"(Override) Serve Port")

	// format override
	supportedFormats := strings.Join(util.SupportedCaptureFormats(), ",")
	serveCmd.Flags().StringSliceVar(&serveCmdOpts.Formats,
		"format", util.EnvVarStringSlice("SERVE_FILE_FORMATS", ""),
		fmt.Sprintf("(Override) Serve Browsing Formats - Supported Formats: [%s]", supportedFormats))
}

// Page is the page in a browser
type Page struct {
	Folders []Folder
	Images  []Image
}

// Image is a wrapper for an aws image
type Image struct {
	Base64 string
	Key    string
}

// Folder is a wrapper for an aws folder
type Folder struct {
	Key        string
	DisplayKey string
}

// PageTemplate describes how the page should look
var PageTemplate string = `<!DOCTYPE html>
<html lang="en">
<head>
	<link rel="icon" href="data:,">
</head>
<body>
	{{range .Folders}}
	<p>
		<a href="?dir={{.Key}}">{{.DisplayKey}}</a>
	</p>
	{{end}}
	{{range .Images}}
	<span>
		<p>{{.Key}}</p>
		<img src="data:image/jpg;base64,{{.Base64}}">
	</span>
	{{end}}
</body></html>`

// ServeCmdGetHandler is a proving ground right meow
func ServeCmdGetHandler(opts *ServeCmdOptions) func(w http.ResponseWriter, r *http.Request) {
	funcTag := "ServeCmdGetHandler"
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("REQUEST: %s, %s, %s", r.Method, r.URL, r.RequestURI)

		if r.Method == http.MethodGet {

			// TODO: get the key/dir from the url
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
			logrus.Infof("QP: %s", qpS3SubKey)

			// get a new s3 client
			awsSesh, s3Client, err := util.NewS3Client()
			if err != nil {
				err = util.WrapError(err, funcTag, "get a new s3 client")
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			// get the object list
			objects, commonKeys, err := util.S3ObjectsByKey(s3Client, qpS3SubKey)
			if err != nil {
				err = util.WrapError(err, funcTag, "get bucket contents info by key")
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			tmpl, err := template.New("image").Parse(PageTemplate)
			if err != nil {
				err = util.WrapError(err, funcTag, "parse object image into html template")
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			// build the page's images
			p := &Page{}

			for _, commonKey := range commonKeys {
				cliInputDir := opts.S3Dir
				if len(cliInputDir) > 0 {
					cliInputDir += util.S3Delimiter
				}
				linkKey := strings.ReplaceAll(commonKey, cliInputDir, "")
				keysSlice := strings.Split(commonKey, util.S3Delimiter)
				displayKey := keysSlice[len(keysSlice)-2]
				logrus.Infof("LINK KEY: %s (%s), %s", commonKey, cliInputDir, displayKey)
				p.Folders = append(p.Folders, Folder{Key: linkKey, DisplayKey: displayKey})
			}

			for _, obj := range objects {

				imgBytes, err := util.DownloadS3Object(awsSesh, *obj.Key)
				if err != nil {
					err = util.WrapError(err, funcTag, "downloading object")
					http.Error(w, err.Error(), http.StatusBadRequest)
				}

				cliInputDir := opts.S3Dir
				if len(cliInputDir) > 0 {
					cliInputDir += util.S3Delimiter
				}
				imgKey := strings.ReplaceAll(*obj.Key, cliInputDir, "")
				p.Images = append(p.Images, Image{
					Base64: base64.StdEncoding.EncodeToString(imgBytes),
					Key:    imgKey,
				})
			}

			// exec the template and data
			tmpl.Execute(w, p)

			// for _, itm := range list.Contents {
			// 	logrus.Infof("%+v", itm)
			// }
		}
	}
}

func corsHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			//handle preflight in here
		} else {
			h.ServeHTTP(w, r)
		}
	}
}

// ServeCmdRunE runs the serve command
// it is exported for testing
func ServeCmdRunE(ropts *RootCmdOptions, opts *ServeCmdOptions) error {
	funcTag := "ServeCmdRunE"
	logrus.Infof(funcTag)

	http.HandleFunc("/", ServeCmdGetHandler(opts))

	hostNPort := fmt.Sprintf("%s:%d", "localhost", opts.Port)
	logrus.Warnf("Go to `http://%s` in your browser ...", hostNPort)

	// go (func() {
	err := http.ListenAndServe(hostNPort, nil)
	if err != nil {
		logrus.Warnf(util.WrapError(err, funcTag, "serving content").Error())
	}
	// })()

	return nil
}
