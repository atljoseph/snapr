package cli

import (
	"fmt"
	"snapr/util"
	"strings"

	"github.com/spf13/cobra"
)

// SnapCmdOptions options
type SnapCmdOptions struct {
	OutDir             string
	CaptureDeviceAddr  string
	OutDirExtra        string
	OutFile            string
	Format             string
	OutDirUsers        bool
	UploadAfterSuccess bool
	CleanupAfterUpload bool
}

// snap command
var (
	snapCmdOpts = &SnapCmdOptions{}
	snapCmd     = &cobra.Command{
		Use:   "snap",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			snapCmdOpts = snapCmdOpts.TransformPositionalArgs(args)
			rootCmdOpts = rootCmdOpts.SetupS3ConfigFromRootArgs()
			return SnapCmdRunE(rootCmdOpts, snapCmdOpts)
		},
	}
)

// TransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func (opts *SnapCmdOptions) TransformPositionalArgs(args []string) *SnapCmdOptions {
	// if len(args) > 0 {
	// // can use env vars, too!
	// 	opts.Something = args[0]
	// }
	return opts
}

func init() {
	// add command to root
	rootCmd.AddCommand(snapCmd)

	// capture device address
	snapCmd.Flags().StringVar(&snapCmdOpts.CaptureDeviceAddr,
		"device", util.EnvVarString("SNAP_DEVICE", util.DefaultCaptureDevice()),
		fmt.Sprintf("(Recommended) Capture Device Address - Try it out without this set, or run `snapr list` to discover possibilities"))

	// this is appended to `dir`if set
	snapCmd.Flags().StringVar(&snapCmdOpts.OutDirExtra,
		"extra-dir", util.EnvVarString("SNAP_DIR_EXTRA", ""),
		"(Optional) Output Directory - Appended to the Output Directory")

	// this is where the files get written to
	snapCmd.Flags().StringVar(&snapCmdOpts.OutDir,
		"dir", util.EnvVarString("SNAP_DIR", ""),
		fmt.Sprintf("(Recommended) Output Directory"))

	// file override ... optional
	// `--dir` is ignored if using this
	snapCmd.Flags().StringVar(&snapCmdOpts.OutFile,
		"file", util.EnvVarString("SNAP_FILE", ""),
		"(Override) Output File")

	// format override
	supportedFormats := strings.Join(util.SupportedCaptureFormats(), ",")
	snapCmd.Flags().StringVar(&snapCmdOpts.Format,
		"format", util.EnvVarString("SNAP_FILE_FORMAT", ""),
		fmt.Sprintf("(Optional) Output Format - Ignored if using '--file' - Supported Formats: [%s]", supportedFormats))

	// add extra dir with users logged in to the filename
	snapCmd.Flags().BoolVar(&snapCmdOpts.OutDirUsers,
		"users", util.EnvVarBool("SNAP_FILE_USERS", false),
		"(Optional) Append Logged in Users to Output Directory - Will be ignored if '--file' is used")

	// Upload flag to mark for upload
	snapCmd.Flags().BoolVar(&snapCmdOpts.UploadAfterSuccess,
		"upload", util.EnvVarBool("SNAP_UPLOAD_AFTER_SUCCESS", false),
		"(Optional) Upload the image file after creation")

	// Cleanup flag to mark for upload
	snapCmd.Flags().BoolVar(&snapCmdOpts.CleanupAfterUpload,
		"cleanup", util.EnvVarBool("SNAP_CLEANUP_AFTER_UPLOAD", false),
		"(Optional) Remove the file after successful upload - Ignored unless using `--upload`")

}
