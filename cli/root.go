package cli

import (
	"os"
	"snapr/util"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// RootCmdOptions are for root flags
type RootCmdOptions struct {
	EnvFilePath string
	Bucket      string
	Region      string
	Token       string
	Secret      string
	S3Config    *util.S3Accessor
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
	newRopts := &RootCmdOptions{
		Bucket: util.EnvVarString("S3_BUCKET", ""),
		Region: util.EnvVarString("S3_REGION", ""),
		Token:  util.EnvVarString("S3_TOKEN", ""),
		Secret: util.EnvVarString("S3_SECRET", ""),
	}
	// if cli had this set, then override
	// we didn't want to show the defaults to the user
	if len(ropts.Bucket) > 0 {
		newRopts.Bucket = ropts.Bucket
	}
	if len(ropts.Region) > 0 {
		newRopts.Region = ropts.Region
	}
	if len(ropts.Token) > 0 {
		newRopts.Token = ropts.Token
	}
	if len(ropts.Secret) > 0 {
		newRopts.Secret = ropts.Secret
	}
	// ctransform / condition
	newRopts.S3Config = &util.S3Accessor{
		Bucket: newRopts.Bucket,
		Region: newRopts.Region,
		Token:  newRopts.Token,
		Secret: newRopts.Secret,
	}
	// logrus.Infof("S3: %+v", newRopts)
	return newRopts
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
