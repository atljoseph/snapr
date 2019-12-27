package cli

import (
	"snapr/util"

	"github.com/spf13/cobra"
)

// TODO: Rename Test AFTER the Upload Tests

// RenameCmdOptions options
type RenameCmdOptions struct {
	S3SourceKey     string
	S3DestKey       string
	S3DestBucket    string
	SrcIsDir        bool
	IsCopyOperation bool
	IsDestPublic    bool
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
		Use:   "rename",
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
		"s3-src-key", util.EnvVarString("RENAME_S3_SRC_KEY", ""),
		"(Required) S3 Key to move ... the source")

	// this is where the files are pulled from
	renameCmd.Flags().StringVar(&renameCmdOpts.S3DestKey,
		"s3-dest-key", util.EnvVarString("RENAME_S3_DEST_KEY", ""),
		"(Required) S3 Key to move `s3-src-key` to")

	// this is where the files are pulled from
	renameCmd.Flags().StringVar(&renameCmdOpts.S3DestBucket,
		"s3-dest-bucket", util.EnvVarString("RENAME_S3_DEST_BUCKET", ""),
		"(Optional) S3 Bucket to move `s3-src-key` to, otherwise, same bucket")

	// is this a directory?
	renameCmd.Flags().BoolVar(&renameCmdOpts.SrcIsDir,
		"s3-src-is-dir", util.EnvVarBool("RENAME_S3_SRC_IS_DIR", false),
		"(Optional) Set this option to rename an entire S3 file or directory")

	// is this a copy operation
	renameCmd.Flags().BoolVar(&renameCmdOpts.IsCopyOperation,
		"copy", util.EnvVarBool("RENAME_IS_COPY_OPERATION", false),
		"(Optional) Set this option to copy an entire S3 file or directory")

	// is this file public?
	renameCmd.Flags().BoolVar(&renameCmdOpts.IsDestPublic,
		"s3-dest-is-public", util.EnvVarBool("RENAME_S3_DEST_IS_PUBLIC", false),
		"(Optional) Use this to copy as a publicly available file, otherwise its private. Requires a public S3!")
}
