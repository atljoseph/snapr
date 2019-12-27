package cli

import (
	"snapr/util"

	"github.com/spf13/cobra"
)

// TODO: Download Test AFTER the Upload & Rename Tests

// DownloadCmdOptions options
type DownloadCmdOptions struct {
	S3Key  string
	IsDir  bool
	OutDir string
}

// upload command
var (
	downloadCmdOpts = &DownloadCmdOptions{}
	downloadCmd     = &cobra.Command{
		Use:   "download",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deleteCmdOpts = deleteCmdOpts.TransformPositionalArgs(args)
			rootCmdOpts = rootCmdOpts.SetupS3ConfigFromRootArgs()
			return DownloadCmdRunE(rootCmdOpts, downloadCmdOpts)
		},
	}
)

// TransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func (opts *DownloadCmdOptions) TransformPositionalArgs(args []string) *DownloadCmdOptions {
	// if len(args) > 0 {
	// // can use env vars, too!
	// 	opts.Something = args[0]
	// }
	return opts
}

func init() {
	// add command to root
	rootCmd.AddCommand(downloadCmd)

	// this is where the files are written to
	downloadCmd.Flags().StringVar(&downloadCmdOpts.OutDir,
		"dir", util.EnvVarString("DOWNLOAD_DIR", ""),
		"(Optional) Download Directory")

	// this is where the files are pulled from
	downloadCmd.Flags().StringVar(&downloadCmdOpts.S3Key,
		"s3-key", util.EnvVarString("DOWNLOAD_S3_KEY", ""),
		"(Required) S3 Key or Directory to delete")

	// file override ... optional
	downloadCmd.Flags().BoolVar(&downloadCmdOpts.IsDir,
		"is-dir", util.EnvVarBool("DOWNLOAD_IS_DIR", false),
		"(Optional) Set this option to delete an entire S3 directory")
}
