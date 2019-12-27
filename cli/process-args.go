package cli

import (
	"snapr/util"

	"github.com/spf13/cobra"
)

// TODO: Download Test AFTER the Upload & Rename Tests

// ProcessCmdOptions options
type ProcessCmdOptions struct {
	S3InKey  string
	S3OutKey string
	OutSizes []int
	Public   bool
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
	processCmd.Flags().StringVar(&processCmdOpts.S3InKey,
		"s3-in-key", util.EnvVarString("PROCESS_S3_IN_KEY", "originals"),
		"(Required) S3 Key or Directory to process")

	// this is where the files are dumped when processing
	processCmd.Flags().StringVar(&processCmdOpts.S3OutKey,
		"s3-out-key", util.EnvVarString("PROCESS_S3_OUT_KEY", "processed"),
		"(Required) S3 Key or Directory to output processing results")

	// each output size will get its own folder
	processCmd.Flags().IntSliceVar(&processCmdOpts.OutSizes,
		"is-dir", util.EnvVarIntSlice("PROCESS_OUT_SIZES", []int{640, 768, 1024}),
		"(Optional) Specify the output sizes of the processing")

	// is this file public?
	processCmd.Flags().BoolVar(&processCmdOpts.Public,
		"public", util.EnvVarBool("PROCESS_PUBLIC", false),
		"(Optional) Use this to upload a publicly available file, otherwise its private. Requires a public S3!")
}
