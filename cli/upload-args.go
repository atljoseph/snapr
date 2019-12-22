package cli

import (
	"fmt"
	"snapr/util"
	"strings"

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
			uploadCmdOpts = uploadCmdOpts.TransformPositionalArgs(args)
			rootCmdOpts = rootCmdOpts.SetupS3ConfigFromRootArgs()
			return UploadCmdRunE(rootCmdOpts, uploadCmdOpts)
		},
	}
)

// TransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func (opts *UploadCmdOptions) TransformPositionalArgs(args []string) *UploadCmdOptions {
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
