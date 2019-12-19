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
	InFileOverride      string
	CleanupAfterSuccess bool
	Formats             string
	UploadLimit         int
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
			return UploadCmdRunE(uploadCmdOpts)
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
	// default to the directory where the binary exists (pwd)
	pwd, _ := os.Getwd()
	uploadCmd.Flags().StringVar(&uploadCmdOpts.InDir,
		"dir", util.EnvVarString("UPLOAD_DIR", pwd),
		"(Optional) Upload Directory")

	// file override ... optional
	uploadCmd.Flags().StringVar(&uploadCmdOpts.InFileOverride,
		"file", util.EnvVarString("UPLOAD_FILE", ""),
		"(Optional) Upload File Path")

	// delete all uploaded files after success
	uploadCmd.Flags().BoolVar(&uploadCmdOpts.CleanupAfterSuccess,
		"cleanup", util.EnvVarBool("UPLOAD_CLEANUP_AFTER_SUCCESS", false),
		"(Optional) Delete file after uploading")

	// upload format filter
	supportedFormats := strings.Join(util.SupportedCaptureFormats(), ",")
	snapCmd.Flags().StringVar(&uploadCmdOpts.Formats,
		"formats", util.EnvVarString("UPLOAD_FORMATS", ""),
		fmt.Sprintf("(Optional) Upload Formats (comma delimited) - Ignored if using '--snap-file' - Supported Formats: [%s]", supportedFormats))

	// upload file limit
	snapCmd.Flags().IntVar(&uploadCmdOpts.UploadLimit,
		"limit", util.EnvVarInt("UPLOAD_LIMIT", 1),
		"(Optional) Limit the number of files to upload in any one operation - Ignored if using '--file'")

}

// UploadCmdRunE runs the snap command
// it is exported for testing
func UploadCmdRunE(opts *UploadCmdOptions) error {
	funcTag := "upload"
	logrus.Infof("Upload")

	// get a new aws session
	s, err := util.NewAwsSession()
	if err != nil {
		return util.WrapError(err, funcTag, "get new aws session")
	}

	// TODO: get a list of files to upload to the bucket
	// based on the base dir, etc
	// ignore files without specific filename format
	// if too many files, give up
	// after success, rename the file

	// get the abs dir path
	opts.InDir, err = filepath.Abs(opts.InDir)
	if err != nil {
		logrus.Warnf("cannot convert path to absolute dir path: %s", opts.InDir)
	}

	// only do this if indir is empty
	if len(opts.InDir) == 0 {
		// get the abs file path
		opts.InFileOverride, err = filepath.Abs(opts.InFileOverride)
		if err != nil {
			logrus.Warnf("cannot convert path to absolute file path: %s", opts.InFileOverride)
		}
	}

	// build the file path
	inFilePath := opts.InFileOverride
	if len(opts.InDir) > 0 {
		inFilePath = filepath.Join(opts.InDir, opts.InFileOverride)
	}

	// stat the file
	fileInfo, err := os.Stat(inFilePath)
	if err != nil {
		return util.WrapError(err, funcTag, "cannot stat file")
	}

	// ensure is a file
	if fileInfo.IsDir() {
		return util.WrapError(fmt.Errorf("validation error"), funcTag, "file cannot be a directory")
	}

	logrus.Infof("Uploading %s", inFilePath)

	// send to AWS
	key, err := util.SendToAws(s, inFilePath)
	if err != nil {
		return util.WrapError(err, funcTag, "sending file to aws")
	}

	// done
	logrus.Infof("Done uploading key: %s", key)
	return nil
}
