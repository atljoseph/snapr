package cli

import (
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// RootCmdOptions are for root flags
type RootCmdOptions struct {
	EnvFilePath string
}

// snap command
var (
	rootCmdOpts = &RootCmdOptions{}
	rootCmd     = &cobra.Command{
		Use:   "snapr",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logrus.Infof("Please enter a command or use the `--help` flag.")
			logrus.Infof(runtime.GOOS)
			return nil
		},
	}
)

func init() {
	// root flags defined here
}

// Execute starts the cli
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// warn for visibility
		logrus.Warnf(err.Error())
		// exit with status 0 for PAM
		os.Exit(0)
	}
}
