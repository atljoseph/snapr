package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"snapr/util"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// SnapCmdOptions options
type SnapCmdOptions struct {
	OutDir            string
	CaptureDeviceAddr string
	OutDirExtra       string
	OutFileOverride   string
	Format            string
	PrependUsers      bool
}

// snap command
var (
	snapCmdOpts = &SnapCmdOptions{}
	snapCmd     = &cobra.Command{
		Use:   "snap",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			snapCmdOpts = snapCmdTransformPositionalArgs(args, snapCmdOpts)
			return SnapCmdRunE(snapCmdOpts)
		},
	}
)

// snapCmdTransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func snapCmdTransformPositionalArgs(args []string, opts *SnapCmdOptions) *SnapCmdOptions {
	// if len(args) > 0 {
	// // can use env vars, too!
	// 	opts.Something = args[0]
	// }
	return opts
}

func init() {
	// add command to root
	rootCmd.AddCommand(snapCmd)

	// capture device address
	snapCmd.Flags().StringVar(&snapCmdOpts.CaptureDeviceAddr,
		"device", util.EnvVarString("SNAP_DEVICE", util.DefaultCaptureDevice()),
		fmt.Sprintf("(Recommended) Capture Device Address - Try it out without this set, or run `snapr list` to discover possibilities"))

	// this is appended to `dir`if set
	snapCmd.Flags().StringVar(&snapCmdOpts.OutDirExtra,
		"dir-extra", util.EnvVarString("SNAP_DIR_EXTRA", ""),
		"(Optional) Output Directory - Appended to the Output Directory")

	// this is where the files get written to
	// default to the directory where the binary exists (pwd)
	pwd, _ := os.Getwd()
	snapCmd.Flags().StringVar(&snapCmdOpts.OutDir,
		"dir", util.EnvVarString("SNAP_DIR", pwd),
		fmt.Sprintf("(Recommended) Output Directory"))

	// file override ... optional
	snapCmd.Flags().StringVar(&snapCmdOpts.OutFileOverride,
		"file", util.EnvVarString("SNAP_FILE", ""),
		"(Override) Output File")

	// format override
	supportedFormats := strings.Join(util.SupportedCaptureFormats(), ",")
	snapCmd.Flags().StringVar(&snapCmdOpts.Format,
		"format", util.EnvVarString("SNAP_FILE_FORMAT", ""),
		fmt.Sprintf("(Optional) Output Format - Ignored if using '--snap-file' - Supported Formats: [%s]", supportedFormats))

	// TODO: make this add an extra dir instead of prepending
	// prepend users logged in to the filename
	snapCmd.Flags().BoolVar(&snapCmdOpts.PrependUsers,
		"users", util.EnvVarBool("SNAP_FILE_USERS", false),
		"(Optional) Prepend Logged in Users to auto-generated filename - Will be ignored if '--snap-file' is used")

	// TODO: Upload flag, or mark for upload
}

// SnapCmdRunE runs the snap command
// it is exported for testing
func SnapCmdRunE(opts *SnapCmdOptions) error {
	funcTag := "SnapCmdRunE"
	logrus.Infof("Snap")

	// validate the format override
	// TODO: add the format check for the file name override
	if len(opts.Format) > 0 {
		if !util.IsSupportedCaptureFormat(opts.Format) {
			return util.WrapError(fmt.Errorf("Validation Error"), funcTag, "get users list")
		}
	}

	// build the out file name
	outFileName := ""
	outFileNameTimeFormat := "2006-01-02T15-04-05"
	if len(opts.Format) == 0 {
		opts.Format = util.DefaultCaptureFormat()
	}

	// if not overridden filename
	if len(opts.OutFileOverride) == 0 {

		// get the users if specified
		if opts.PrependUsers {
			// get the logged in users list
			usersExec := exec.Command("/bin/sh", "-c", "users")
			usersBytes, err := usersExec.Output()
			if err != nil {
				return util.WrapError(err, funcTag, "get users list")
			}
			usersOutput := bytes.NewBuffer(usersBytes).String()
			// remove line breaks
			usersOutput = strings.ReplaceAll(usersOutput, "\n", "")
			// replace spaces ith "-"
			usersOutput = strings.ReplaceAll(usersOutput, " ", "-")

			// add the users output
			outFileName += usersOutput + "-"
		}

		// add the time format
		outFileName = time.Now().Format(outFileNameTimeFormat)

		// add the extension
		outFileName = outFileName + "." + opts.Format
	} else {

		// overriding filename
		outFileName = opts.OutFileOverride
	}

	var err error

	// get the abs dir path
	opts.OutDir, err = filepath.Abs(opts.OutDir)
	if err != nil {
		logrus.Warnf("cannot convert path to absolute dir path: %s", opts.OutDir)
	}

	// only do this if indir is empty
	if len(opts.OutDir) == 0 {
		// get the abs file path
		opts.OutFileOverride, err = filepath.Abs(opts.OutFileOverride)
		if err != nil {
			logrus.Warnf("cannot convert path to absolute file path: %s", opts.OutFileOverride)
		}
	}

	// directory where the output file will go
	outFileDir := opts.OutDir
	if len(opts.OutDirExtra) > 0 {
		outFileDir = filepath.Join(outFileDir, opts.OutDirExtra)
	}

	// the complete out dir and file path
	outFilePath := fmt.Sprintf("%s/%s", outFileDir, outFileName)
	// outFilePath = strings.ReplaceAll(outFilePath, "//", "/")
	logrus.Infof("Snapping %s", outFilePath)

	// ensure output dir exists
	err = os.MkdirAll(outFileDir, 0700)
	if err != nil {
		return util.WrapError(err, funcTag, "mkdir for outFileDir")
	}

	// input device driver
	driverType := "-f "
	if strings.EqualFold(runtime.GOOS, "linux") {
		driverType += "video4linux2"
	}
	if strings.EqualFold(runtime.GOOS, "darwin") {
		driverType += "avfoundation"
	}

	// resolution
	resolution := "-s 640x480"

	// vframes are video frames
	vframes := "-vframes 1"

	// framerate only on mac
	framerate := ""
	if strings.EqualFold(runtime.GOOS, "darwin") {
		framerate += "-framerate 30"
	}

	// capture address
	webcamAddr := "-i "
	if len(opts.CaptureDeviceAddr) > 0 {
		webcamAddr += opts.CaptureDeviceAddr
	} else {
		webcamAddr += util.DefaultCaptureDevice()
	}

	// check if the ffmpeg command exists
	_, err = exec.LookPath("ffmpeg")
	if err != nil {
		suggestion := "dependency issue ... please install ffmpeg: `%s`"
		cmdSuggestion := ""
		if strings.EqualFold(runtime.GOOS, "linux") {
			cmdSuggestion = "sudo apt install ffmpeg"
		}
		if strings.EqualFold(runtime.GOOS, "darwin") {
			cmdSuggestion = "brew install ffmpeg"
		}
		suggestion = fmt.Sprintf(suggestion, cmdSuggestion)
		logrus.Warnf(suggestion)
		return util.WrapError(err, funcTag, suggestion)
	}

	// build the os cmd to execute
	// capture command with ffmpeg
	// overwrite existing file if any
	ffmpegExecString := fmt.Sprintf("ffmpeg %s %s %s %s %s %s -y", driverType, framerate, resolution, webcamAddr, vframes, outFilePath)
	logrus.Infof("Command: %s", ffmpegExecString)
	ffmpegExec := exec.Command("/bin/sh", "-c", ffmpegExecString)

	// execute and wait
	err = ffmpegExec.Run()
	if err != nil {
		return util.WrapError(err, funcTag, "running command")
	}

	// done
	logrus.Infof("Snapped %s", outFilePath)
	return nil
}
