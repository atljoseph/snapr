package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// PidgeonCmdOpts options
type PidgeonCmdOpts struct {
	BaseReadDir string
}

// pidgeon command
var (
	pidgeonCmdOpts = PidgeonCmdOpts{}
	pidgeonCmd     = &cobra.Command{
		Use:   "pidgeon",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE:  upload,
	}
)

func init() {
	// add command to root
	rootCmd.AddCommand(pidgeonCmd)

	// this is where the files are pulled from
	pidgeonCmd.Flags().StringVarP(&pidgeonCmdOpts.BaseReadDir, "base-dir", "d", "~/", "Device Address")
}

func upload(cmd *cobra.Command, args []string) error {
	// funcTag := "pidgeon"
	logrus.Infof("Pidgeoning")

	// TODO: upload to amaazon s3 bucket

	// done
	logrus.Infof("Done")
	return nil
}
