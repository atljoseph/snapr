package cli

import (
	"os"
	"snapr/util"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// RootCmdOptions are for root flags
type RootCmdOptions struct {
	EnvFilePath    string
	Bucket         string
	Region         string
	Token          string
	Secret         string
	S3Config       *util.S3Accessor
	FileCreateMode os.FileMode
}

// snap command
var (
	rootCmdOpts = &RootCmdOptions{}
	rootCmd     = &cobra.Command{
		Use:   "snapr",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// initialize the s3 config
			rootCmdOpts = rootCmdOpts.SetupS3ConfigFromRootArgs()
			rootCmdOpts.FileCreateMode = 0700
			// exec this when root cmd is exec-ed
			logrus.Infof("Please enter a command or use the `--help` flag.")
			// logrus.Infof(runtime.GOOS)

			return nil
		},
	}
)

func init() {
	// root flags defined here

	// s3 bucket
	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.Bucket,
		"s3-bucket", "",
		"(Optional) S3 Bucket Identifier")

	// s3 region
	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.Region,
		"s3-region", "",
		"(Optional) S3 Region Identifier")

	// s3 token
	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.Token,
		"s3-token", "",
		"(Optional) S3 User Token")

	// s3 secret
	rootCmd.PersistentFlags().StringVar(&rootCmdOpts.Secret,
		"s3-secret", "",
		"(Optional) S3 User Secret")
}

// SetupS3ConfigFromRootArgs initializes the s3 config from env
// this ensures we do not show sensitive data to users, and are still able to use this for testing
func (ropts *RootCmdOptions) SetupS3ConfigFromRootArgs() *RootCmdOptions {
	// set the S3 stuff up from env
	// if cli did not have these set, then default to env with default
	// we don't want to show the defaults to the user
	// in the cli prompts if set in the env from packr build
	if len(ropts.Bucket) == 0 {
		ropts.Bucket = util.EnvVarString("S3_BUCKET", "")
	}
	if len(ropts.Region) == 0 {
		ropts.Region = util.EnvVarString("S3_REGION", "")
	}
	if len(ropts.Token) == 0 {
		ropts.Token = util.EnvVarString("S3_TOKEN", "")
	}
	if len(ropts.Secret) == 0 {
		ropts.Secret = util.EnvVarString("S3_SECRET", "")
	}
	// ctransform / condition
	ropts.S3Config = &util.S3Accessor{
		Bucket: ropts.Bucket,
		Region: ropts.Region,
		Token:  ropts.Token,
		Secret: ropts.Secret,
	}
	// logrus.Infof("S3: %+v", ropts)
	return ropts
}

// Execute starts the cli
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// warn for visibility
		logrus.Warnf(err.Error())
		// exit with status 0 for PAM
		// would normally use code 1
		os.Exit(0)
	}
}
