package cli

import (
	"snapr/util"

	"github.com/spf13/cobra"
)

// TODO: Document this command
// TODO: Test this command

// GrepCmdOptions options
type GrepCmdOptions struct {
	S3Key           string
	S3Dir           string
	OutDir          string
	SearchPattern   string
	SearchIsLiteral bool
	TruncationLimit int
}

// upload command
var (
	grepCmdOpts = &GrepCmdOptions{}
	grepCmd     = &cobra.Command{
		Use:   "grep",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			grepCmdOpts = grepCmdOpts.TransformPositionalArgs(args)
			rootCmdOpts = rootCmdOpts.SetupS3ConfigFromRootArgs()
			return GrepCmdRunE(rootCmdOpts, grepCmdOpts)
		},
	}
)

// TransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func (opts *GrepCmdOptions) TransformPositionalArgs(args []string) *GrepCmdOptions {
	// if len(args) > 0 {
	// // can use env vars, too!
	// 	opts.Something = args[0]
	// }
	return opts
}

func init() {
	// add command to root
	rootCmd.AddCommand(grepCmd)

	// this is where the files are pulled from to search
	grepCmd.Flags().StringVar(&grepCmdOpts.S3Dir,
		"s3-dir", util.EnvVarString("GREP_S3_DIR", "s3gogrep"),
		"(Required) S3 Directory to process. Probably not good to search the entire bucket.")

	// this is where the files are deposited when they match
	grepCmd.Flags().StringVar(&grepCmdOpts.OutDir,
		"out-dir", util.EnvVarString("GREP_OUT_DIR", "results"),
		"(Required) This is where the files are deposited when they match.")

	// how to match files
	// https://golang.org/pkg/regexp/syntax/
	grepCmd.Flags().StringVar(&grepCmdOpts.SearchPattern,
		"pattern", util.EnvVarString("GREP_PATTERN", "order"),
		`(Required) Pattern to search for... Can be a standard string, literal, or regex pattern
		Help is here: 'https://golang.org/pkg/regexp/syntax/'.
		Known Scenarios:
		> Case-Insensitive 		=>> (i?)POst
		> OR 					=>> PUT|POST
		> AND					=>> POST.*?00645618
		> AND / OR Combo		=>> (POST.*?00645618)|(PUT.*?00645618)
		`)

	// use a literal match?
	grepCmd.Flags().BoolVar(&grepCmdOpts.SearchIsLiteral,
		"literal", util.EnvVarBool("GREP_IS_LITERAL", false),
		"(Optional) Conduct a literal search? Surrounds your pattern with '\\b' (same as '\\\\b' escaped).")

	// truncation max length
	grepCmd.Flags().IntVar(&grepCmdOpts.TruncationLimit,
		"truncate", util.EnvVarInt("GREP_TRUNCATION_LIMIT", 400),
		"(Optional) Should the displayed results be truncated other than the default? This is a positive number.")

	// truncation max length
	grepCmd.Flags().StringVar(&grepCmdOpts.S3Key,
		"s3-key", util.EnvVarString("GREP_S3_KEY", ""),
		`(Optional) Set this to search only a specific object (S3 Key). 
		Can be a partial matched key as well.`)
}
