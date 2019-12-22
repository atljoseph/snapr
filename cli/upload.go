package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"snapr/util"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// UploadCmdOptions options
type UploadCmdOptions struct {
	InDir               string
	InFile              string
	CleanupAfterSuccess bool
	Formats             []string
	UploadLimit         int
	S3Dir               string
	S3Bucket            string
	S3Region            string
}

// upload command
var (
	uploadCmdOpts = &UploadCmdOptions{}
	uploadCmd     = &cobra.Command{
		Use:   "upload",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			uploadCmdOpts = uploadCmdTransformPositionalArgs(args, uploadCmdOpts)
			return UploadCmdRunE(rootCmdOpts, uploadCmdOpts)
		},
	}
)

// uploadCmdTransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func uploadCmdTransformPositionalArgs(args []string, opts *UploadCmdOptions) *UploadCmdOptions {
	// if len(args) > 0 {
	// // can use env vars, too!
	// 	opts.Something = args[0]
	// }
	return opts
}

func init() {
	// add command to root
	rootCmd.AddCommand(uploadCmd)

	// this is where the files are pulled from
	uploadCmd.Flags().StringVar(&uploadCmdOpts.InDir,
		"dir", util.EnvVarString("UPLOAD_DIR", ""),
		"(Optional) Upload Directory")

	// file override ... optional
	uploadCmd.Flags().StringVar(&uploadCmdOpts.InFile,
		"file", util.EnvVarString("UPLOAD_FILE", ""),
		"(Optional) Upload File Path")

	// delete all uploaded files after success
	uploadCmd.Flags().BoolVar(&uploadCmdOpts.CleanupAfterSuccess,
		"cleanup", util.EnvVarBool("UPLOAD_CLEANUP_AFTER_SUCCESS", false),
		"(Optional) Delete file after uploading")

	// upload format filter
	supportedFormats := strings.Join(util.SupportedCaptureFormats(), ",")
	uploadCmd.Flags().StringSliceVar(&uploadCmdOpts.Formats,
		"formats", util.EnvVarStringSlice("UPLOAD_FORMATS", ""),
		fmt.Sprintf("(Optional) Upload Formats (comma delimited) - Ignored if using '--file' - Supported Formats: [%s]", supportedFormats))

	// upload file limit
	uploadCmd.Flags().IntVar(&uploadCmdOpts.UploadLimit,
		"limit", util.EnvVarInt("UPLOAD_LIMIT", 1),
		"(Optional) Limit the number of files to upload in any one operation - Ignored if using '--file'")

}

// UploadCmdRunE runs the snap command
// it is exported for testing
func UploadCmdRunE(ropts *RootCmdOptions, opts *UploadCmdOptions) error {
	funcTag := "upload"
	logrus.Infof(funcTag)

	// get a new aws session
	_, s3Client, err := util.NewS3Client()
	if err != nil {
		return util.WrapError(err, funcTag, "get new aws session")
	}

	// check limit, is it a crazy high number? if so kick it back
	if opts.UploadLimit > 100 {
		return util.WrapError(fmt.Errorf("Validation Error"), funcTag, "choose an upload limit smaller than 100")
	}

	// handle the dir and file inputs
	// and get a list of files based on the inputs
	var files []util.WalkedFile
	if len(opts.InFile) > 0 {
		// if the file override is set,
		// ignore the walk and upload limit

		// if dir is also set, join
		if len(opts.InDir) > 0 {
			opts.InFile = filepath.Join(opts.InDir, opts.InFile)
		}

		// get the abs file path
		absPath, err := filepath.Abs(opts.InFile)
		if err != nil {
			logrus.Warnf("cannot convert path to absolute file path: %s", opts.InFile)
		}

		// set these explicitly
		opts.InDir = filepath.Dir(absPath)
		opts.InFile = filepath.Base(absPath)

		// join the override file path with the dir
		fullPath := filepath.Join(opts.InDir, opts.InFile)

		// stat the path
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return util.WrapError(err, funcTag, "cannot stat path")
		}

		// ensure is a file
		if fileInfo.IsDir() {
			return util.WrapError(fmt.Errorf("validation error"), funcTag, "file cannot be a directory")
		}

		// append the walked file struct
		files = append(files, util.WalkedFile{
			Path:     fullPath,
			FileInfo: fileInfo,
		})
	} else {
		// file override is empty

		// default the in dir if empty
		if len(opts.InDir) == 0 {
			// default to the directory where the binary exists (pwd)
			opts.InDir, err = os.Getwd()
			if err != nil {
				return util.WrapError(err, funcTag, "cannot get pwd for InDir")
			}
		}

		// get the abs dir path
		opts.InDir, err = filepath.Abs(opts.InDir)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("cannot convert path to absolute dir path: %s", opts.InDir))
		}

		// stat the path
		fileInfo, err := os.Stat(opts.InDir)
		if err != nil {
			return util.WrapError(err, funcTag, "cannot stat path")
		}

		// ensure is a dir
		if !fileInfo.IsDir() {
			return util.WrapError(fmt.Errorf("validation error"), funcTag, "dir provided is not a directory")
		}

		// get the slice of walkedFiles
		// based on the indir, walk all files
		files, err = util.WalkFiles(opts.InDir)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("walking dir for files to upload: %s", opts.InDir))
		}
	}

	logrus.Infof("Got %d files before filtering", len(files))

	// TODO: order the files with the oldest first and newest last

	// validate the formats input
	// weirdness with the cobra lib and the []string var
	if len(opts.Formats) == 0 || len(opts.Formats[0]) == 0 {
		opts.Formats = util.SupportedCaptureFormats()
	}

	// filter out files without specific filename format
	var filteredFiles []util.WalkedFile
	for _, file := range files {

		// get the file extension, and replace the dot (weirdness of this lib)
		fileExt := strings.ReplaceAll(filepath.Ext(file.Path), ".", "")

		// filter formats
		for _, format := range opts.Formats {
			// if the format matches
			if strings.EqualFold(fileExt, format) {
				// append the file to the slice
				filteredFiles = append(filteredFiles, file)
			}
		}
	}

	logrus.Infof("Got %d files after filtering", len(filteredFiles))

	// if no files after filtering, error
	if len(filteredFiles) == 0 {
		return util.WrapError(fmt.Errorf("Validation Error"), funcTag, "no files with specified format exist at target")
	}

	// attempt to chop off a slice of these equal to the limit input
	length := len(filteredFiles)
	// if upload limit is greater than 1, take the minimum of the length of files and the limit
	if opts.UploadLimit > 1 {
		length = util.MinInt(opts.UploadLimit, len(filteredFiles))
	}

	// truncate filtered files
	filteredFiles = filteredFiles[0:length]
	logrus.Infof("Got %d files to upload", len(filteredFiles))

	// loop to upload the files
	for _, file := range filteredFiles {

		logrus.Infof("Uploading %s %+v", file.Path, file)

		// send to AWS
		key, err := util.SendToS3(s3Client, opts.InDir, file)
		if err != nil {
			return util.WrapError(err, funcTag, "sending file to aws")
		}

		logrus.Infof("Done uploading key: %s", key)

		// after success, cleanup the files
		if opts.CleanupAfterSuccess {

			// remove the file from the os if desired
			err = os.Remove(file.Path)
			if err != nil {
				return util.WrapError(err, funcTag, "removing the file after upload")
			}

			logrus.Infof("Cleaned up: %s", file.Path)
		}
	}

	return nil
}
