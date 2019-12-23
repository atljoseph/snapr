package cli

import (
	"encoding/base64"
	"net/http"
	"path/filepath"
	"snapr/util"
	"strings"
	"text/template"
)

// BrowsePage is the page in a browser
type BrowsePage struct {
	Folders []Folder
	Files   []Object
	Images  []Object
}

// Object is a wrapper for an aws object
type Object struct {
	Base64     string
	Key        string
	DisplayKey string
}

// Folder is a wrapper for an aws folder
type Folder struct {
	Key        string
	DisplayKey string
}

// BrowsePageTemplate describes how the page should look
var BrowsePageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	<link rel="icon" href="data:,">
</head>
<body>
	{{range .Folders}}
	<p>
		<a href="browse?dir={{.Key}}">{{.DisplayKey}}</a>
	</p>
	{{end}}
	{{range .Files}}
	<p>
		<span>
			{{.DisplayKey}}
			&nbsp;<a href="download?key={{.DisplayKey}}">Download</a>
		</span>
	</p>
	{{end}}
	{{range .Images}}
	<span>
		<p>{{.DisplayKey}}</p>
		<img src="data:image/jpg;base64,{{.Base64}}">
	</span>
	{{end}}
</body></html>`

// ServeCmdBrowseHandler is a proving ground right meow
func ServeCmdBrowseHandler(ropts *RootCmdOptions, opts *ServeCmdOptions) func(w http.ResponseWriter, r *http.Request) {
	funcTag := "ServeCmdBrowseHandler"
	return func(w http.ResponseWriter, r *http.Request) {
		// logrus.Infof("REQUEST (%s): %s, %s, %s", funcTag, r.Method, r.URL, r.RequestURI)

		// only respond to get request (from browser)
		if r.Method == http.MethodGet {

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
			awsSesh, s3Client, err := util.NewS3Client(ropts.S3Config)
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

			// parse the html template into a go object
			tmpl, err := template.New("browse").Parse(BrowsePageTemplate)
			if err != nil {
				err = util.WrapError(err, funcTag, "parse object image into html template")
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			// build the page's images
			p := &BrowsePage{}

			// folders
			// for each sub-directory (common key)
			for _, commonKey := range commonKeys {

				// smash together the cli input s3-dir with the object key
				cliInputDir := opts.S3Dir
				if len(cliInputDir) > 0 {
					cliInputDir += util.S3Delimiter
				}
				// href query param key
				linkKey := strings.ReplaceAll(commonKey, cliInputDir, "")

				// get the last folder in the key
				keysSlice := strings.Split(commonKey, util.S3Delimiter)
				displayKey := keysSlice[len(keysSlice)-2]

				// logrus.Infof("LINK KEY: %s (%s), %s", commonKey, cliInputDir, displayKey)
				p.Folders = append(p.Folders, Folder{Key: linkKey, DisplayKey: displayKey})
			}

			// files and images
			for _, obj := range objects {

				// smash together the cli input s3-dir with the object key
				cliInputDir := opts.S3Dir
				if len(cliInputDir) > 0 {
					cliInputDir += util.S3Delimiter
				}

				// object key in aws
				objKey := strings.ReplaceAll(*obj.Key, cliInputDir, "")

				// determine if file or image
				ext := strings.ReplaceAll(filepath.Ext(objKey), ".", "")

				// is this an image?
				// good compromise for image format determination
				isImage := false
				for _, format := range util.SupportedCaptureFormats() {
					if strings.EqualFold(format, ext) {
						isImage = true
						break
					}
				}

				// if match, put in image slice
				// else file slice
				if isImage {

					// download the object to byte slice
					objBytes, err := util.DownloadS3Object(awsSesh, *obj.Key)
					if err != nil {
						err = util.WrapError(err, funcTag, "downloading object")
						http.Error(w, err.Error(), http.StatusBadRequest)
					}

					// append to images slice
					p.Images = append(p.Images, Object{
						Base64:     base64.StdEncoding.EncodeToString(objBytes),
						DisplayKey: objKey,
						Key:        *obj.Key,
					})
				} else {

					// append to files slice
					p.Files = append(p.Files, Object{
						DisplayKey: objKey,
						Key:        *obj.Key,
					})
				}

			}

			// exec the template and data
			tmpl.Execute(w, p)
		}
	}
}
