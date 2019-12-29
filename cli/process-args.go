package cli

import (
	"snapr/util"

	"github.com/spf13/cobra"
)

// TODO: Document this command
// TODO: Remove orphans flag
// TODO: Default when rebuild-all not set (only new photos)
// TODO: Process Test AFTER the Upload & Rename Tests

// ProcessCmdOptions options
type ProcessCmdOptions struct {
	S3SrcKey     string
	S3DestKey    string
	Sizes        []int
	IsDestPublic bool
	RebuildAll   bool
	RebuildNew   bool
}

// upload command
var (
	processCmdOpts = &ProcessCmdOptions{}
	processCmd     = &cobra.Command{
		Use:   "process",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deleteCmdOpts = deleteCmdOpts.TransformPositionalArgs(args)
			rootCmdOpts = rootCmdOpts.SetupS3ConfigFromRootArgs()
			return ProcessCmdRunE(rootCmdOpts, processCmdOpts)
		},
	}
)

// TransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func (opts *ProcessCmdOptions) TransformPositionalArgs(args []string) *ProcessCmdOptions {
	// if len(args) > 0 {
	// // can use env vars, too!
	// 	opts.Something = args[0]
	// }
	return opts
}

func init() {
	// add command to root
	rootCmd.AddCommand(processCmd)

	// this is where the files are pulled from to process
	processCmd.Flags().StringVar(&processCmdOpts.S3SrcKey,
		"s3-src-key", util.EnvVarString("PROCESS_S3_SRC_KEY", "originals"),
		"(Required) S3 Key or Directory to process")

	// this is where the files are dumped when processing
	processCmd.Flags().StringVar(&processCmdOpts.S3DestKey,
		"s3-dest-key", util.EnvVarString("PROCESS_S3_DEST_KEY", "processed"),
		"(Required) S3 Key or Directory to output processing results")

	// each output size will get its own folder
	processCmd.Flags().IntSliceVar(&processCmdOpts.Sizes,
		"sizes", util.EnvVarIntSlice("PROCESS_SIZES", []int{640, 768, 1024}),
		"(Optional) Specify the output sizes of the processing")

	// is this file public?
	processCmd.Flags().BoolVar(&processCmdOpts.IsDestPublic,
		"s3-is-public", util.EnvVarBool("PROCESS_IS_DEST_PUBLIC", false),
		"(Optional) Use this to upload a publicly available file, otherwise its private. Requires a public S3!")

	// rebuild all files?
	processCmd.Flags().BoolVar(&processCmdOpts.RebuildAll,
		"rebuild-all", util.EnvVarBool("PROCESS_REBUILD_ALL", false),
		"(Optional) Remove the processed directory before processing so that all files are re-processed")

	// rebuild only new files?
	processCmd.Flags().BoolVar(&processCmdOpts.RebuildNew,
		"rebuild-new", util.EnvVarBool("PROCESS_REBUILD_NEW", false),
		"(Optional) Process files that exist in the src which do not exist in the dest")

}
