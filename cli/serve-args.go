package cli

import (
	"snapr/util"

	"github.com/spf13/cobra"
)

// ServeCmdOptions options
type ServeCmdOptions struct {
	WorkDir string
	S3Dir   string
	Port    int
	Host    string
}

// serve command
var (
	serveCmdOpts = &ServeCmdOptions{}
	serveCmd     = &cobra.Command{
		Use:   "serve",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			serveCmdOpts = serveCmdOpts.TransformPositionalArgs(args)
			rootCmdOpts = rootCmdOpts.SetupS3ConfigFromRootArgs()
			return ServeCmdRunE(rootCmdOpts, serveCmdOpts)
		},
	}
)

// TransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func (opts *ServeCmdOptions) TransformPositionalArgs(args []string) *ServeCmdOptions {
	// if len(args) > 0 {
	// // can use env vars, too!
	// 	opts.Something = args[0]
	// }
	return opts
}

func init() {
	// add command to root
	rootCmd.AddCommand(serveCmd)

	// this is appended to `dir`if set
	serveCmd.Flags().StringVar(&serveCmdOpts.S3Dir,
		"s3-dir", util.EnvVarString("SERVE_S3_DIR", ""),
		"(Optional) Base S3 Directory Key to browse")

	// this is where the files get written to
	// default to calling user's home directory
	// TODO: default below
	serveCmd.Flags().StringVar(&serveCmdOpts.WorkDir,
		"work-dir", util.EnvVarString("SERVE_WORK_DIR", ""),
		"(Recommended) This will eventually be the Download and Upload directory")

	// file override ... optional
	// TODO: default below
	serveCmd.Flags().IntVar(&serveCmdOpts.Port,
		"port", util.EnvVarInt("SERVE_PORT", 8080),
		"(Override) Serve Port")
}
