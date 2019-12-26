package cli

import (
	"snapr/util"

	"github.com/spf13/cobra"
)

// TODO: Rename Test AFTER the Upload Tests

// RenameCmdOptions options
type RenameCmdOptions struct {
	S3SourceKey string
	S3DestKey   string
	IsDir       bool
}

// RenameCmdOperationTracker helps track rename operations
type RenameCmdOperationTracker struct {
	Source *util.S3Object
	Dest   *util.S3Object
}

// upload command
var (
	renameCmdOpts = &RenameCmdOptions{}
	renameCmd     = &cobra.Command{
		Use:   "delete",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			renameCmdOpts = renameCmdOpts.TransformPositionalArgs(args)
			rootCmdOpts = rootCmdOpts.SetupS3ConfigFromRootArgs()
			return RenameCmdRunE(rootCmdOpts, renameCmdOpts)
		},
	}
)

// TransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func (opts *RenameCmdOptions) TransformPositionalArgs(args []string) *RenameCmdOptions {
	// if len(args) > 0 {
	// // can use env vars, too!
	// 	opts.Something = args[0]
	// }
	return opts
}

func init() {
	// add command to root
	rootCmd.AddCommand(renameCmd)

	// this is where the files are pulled from
	renameCmd.Flags().StringVar(&renameCmdOpts.S3SourceKey,
		"s3-source-key", util.EnvVarString("RENAME_S3_SOURCE_KEY", ""),
		"(Required) S3 Key to move")

	// this is where the files are pulled from
	renameCmd.Flags().StringVar(&renameCmdOpts.S3DestKey,
		"s3-dest-key", util.EnvVarString("RENAME_S3_DEST_KEY", ""),
		"(Required) S3 Key to move `s3-source-key` to")

	// file override ... optional
	renameCmd.Flags().BoolVar(&renameCmdOpts.IsDir,
		"is-dir", util.EnvVarBool("RENAME_IS_DIR", false),
		"(Optional) Set this option to delete an entire S3 directory")
}
