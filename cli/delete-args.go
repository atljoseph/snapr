package cli

import (
	"snapr/util"

	"github.com/spf13/cobra"
)

// TODO: Delete Test AFTER the Upload & Rename Tests

// DeleteCmdOptions options
type DeleteCmdOptions struct {
	S3Key string
	IsDir bool
}

// upload command
var (
	deleteCmdOpts = &DeleteCmdOptions{}
	deleteCmd     = &cobra.Command{
		Use:   "delete",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deleteCmdOpts = deleteCmdOpts.TransformPositionalArgs(args)
			rootCmdOpts = rootCmdOpts.SetupS3ConfigFromRootArgs()
			return DeleteCmdRunE(rootCmdOpts, deleteCmdOpts)
		},
	}
)

// TransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func (opts *DeleteCmdOptions) TransformPositionalArgs(args []string) *DeleteCmdOptions {
	// if len(args) > 0 {
	// // can use env vars, too!
	// 	opts.Something = args[0]
	// }
	return opts
}

func init() {
	// add command to root
	rootCmd.AddCommand(deleteCmd)

	// this is where the files are pulled from
	deleteCmd.Flags().StringVar(&deleteCmdOpts.S3Key,
		"s3-key", util.EnvVarString("DELETE_S3_KEY", ""),
		"(Required) S3 Key or Directory to delete")

	// file override ... optional
	deleteCmd.Flags().BoolVar(&deleteCmdOpts.IsDir,
		"s3-is-dir", util.EnvVarBool("DELETE_S3_IS_DIR", false),
		"(Optional) Set this option to delete an entire S3 directory")
}
