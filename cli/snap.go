package cli

import (
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
	OutDir             string
	CaptureDeviceAddr  string
	OutDirExtra        string
	OutFile            string
	Format             string
	OutDirUsers        bool
	UploadAfterSuccess bool
	CleanupAfterUpload bool
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
			return SnapCmdRunE(rootCmdOpts, snapCmdOpts)
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
		"extra-dir", util.EnvVarString("SNAP_DIR_EXTRA", ""),
		"(Optional) Output Directory - Appended to the Output Directory")

	// this is where the files get written to
	snapCmd.Flags().StringVar(&snapCmdOpts.OutDir,
		"dir", util.EnvVarString("SNAP_DIR", ""),
		fmt.Sprintf("(Recommended) Output Directory"))

	// file override ... optional
	// `--dir` is ignored if using this
	snapCmd.Flags().StringVar(&snapCmdOpts.OutFile,
		"file", util.EnvVarString("SNAP_FILE", ""),
		"(Override) Output File")

	// format override
	supportedFormats := strings.Join(util.SupportedCaptureFormats(), ",")
	snapCmd.Flags().StringVar(&snapCmdOpts.Format,
		"format", util.EnvVarString("SNAP_FILE_FORMAT", ""),
		fmt.Sprintf("(Optional) Output Format - Ignored if using '--file' - Supported Formats: [%s]", supportedFormats))

	// add extra dir with users logged in to the filename
	snapCmd.Flags().BoolVar(&snapCmdOpts.OutDirUsers,
		"users", util.EnvVarBool("SNAP_FILE_USERS", false),
		"(Optional) Append Logged in Users to Output Directory - Will be ignored if '--file' is used")

	// Upload flag to mark for upload
	snapCmd.Flags().BoolVar(&snapCmdOpts.UploadAfterSuccess,
		"upload", util.EnvVarBool("SNAP_UPLOAD_AFTER_SUCCESS", false),
		"(Optional) Upload the image file after creation")

	// Cleanup flag to mark for upload
	snapCmd.Flags().BoolVar(&snapCmdOpts.CleanupAfterUpload,
		"cleanup", util.EnvVarBool("SNAP_CLEANUP_AFTER_UPLOAD", false),
		"(Optional) Remove the file after successful upload - Ignored unless using `--upload`")
}

// SnapCmdRunE runs the snap command
// it is exported for testing
func SnapCmdRunE(ropts *RootCmdOptions, opts *SnapCmdOptions) error {
	funcTag := "SnapCmdRunE"
	logrus.Infof(funcTag)

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

	fmt.Printf("%+v", opts)

	// handle the dir and file inputs
	var err error
	if len(opts.OutFile) > 0 {

		// if dir is also set, join
		if len(opts.OutDir) > 0 {
			opts.OutFile = filepath.Join(opts.OutDir, opts.OutFile)
		}

		// get the abs file path
		absPath, err := filepath.Abs(opts.OutFile)
		if err != nil {
			logrus.Warnf("cannot convert path to absolute file path: %s", opts.OutFile)
		}

		// set these explicitly
		opts.OutDir = filepath.Dir(absPath)
		opts.OutFile = filepath.Base(absPath)

		// // stat the path
		// fileInfo, err := os.Stat(fullPath)
		// if err != nil {
		// 	return util.WrapError(err, funcTag, "cannot stat path")
		// }

		// // ensure is a file
		// if fileInfo.IsDir() {
		// 	return util.WrapError(fmt.Errorf("validation error"), funcTag, "file cannot be a directory")
		// }
	} else {
		// file override is empty

		// default the in dir if empty
		if len(opts.OutDir) == 0 {
			// default to the directory where the binary exists (pwd)
			opts.OutDir, err = os.Getwd()
			if err != nil {
				return util.WrapError(err, funcTag, "cannot get pwd for OutDir")
			}
		}

		// get the abs dir path
		opts.OutDir, err = filepath.Abs(opts.OutDir)
		if err != nil {
			return util.WrapError(err, funcTag, fmt.Sprintf("cannot convert path to absolute dir path: %s", opts.OutDir))
		}

		// // stat the path
		// fileInfo, err := os.Stat(opts.OutDir)
		// if err != nil {
		// 	return util.WrapError(err, funcTag, "cannot stat path")
		// }

		// // ensure is a dir
		// if !fileInfo.IsDir() {
		// 	return util.WrapError(fmt.Errorf("validation error"), funcTag, "dir provided is not a directory")
		// }

		// build the filename

		// add the time format
		opts.OutFile += time.Now().Format(outFileNameTimeFormat)

		// add the extension
		opts.OutFile = opts.OutFile + "." + opts.Format
	}

	// extra sub-directory for the users output
	// get the users if specified
	if opts.OutDirUsers {
		// get the list of logged in users
		usersOutput, err := util.OSUsers()
		if err != nil {
			return util.WrapError(err, funcTag, "getting users output")
		}

		// add the users output
		opts.OutFile = filepath.Join(usersOutput, opts.OutFile)
	}

	// extra sub-directory where the output file will go
	// goes in directory path BEFORE users output (above)
	if len(opts.OutDirExtra) > 0 {
		opts.OutFile = filepath.Join(opts.OutDirExtra, opts.OutFile)
	}

	// the complete out dir and file path
	outFilePath := fmt.Sprintf("%s/%s", opts.OutDir, opts.OutFile)
	logrus.Infof("Snapping %s", outFilePath)

	// ensure output dir exists
	mkdir := filepath.Dir(outFilePath)
	logrus.Infof("Ensuring Directory: %s", mkdir)
	err = os.MkdirAll(mkdir, 0700)
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
	ffmpegExecString := fmt.Sprintf("ffmpeg %s %s %s %s %s \"%s\" -y", driverType, framerate, resolution, webcamAddr, vframes, outFilePath)
	logrus.Infof("Command: %s", ffmpegExecString)
	ffmpegExec := exec.Command("/bin/sh", "-c", ffmpegExecString)

	// execute and wait
	err = ffmpegExec.Run()
	if err != nil {
		return util.WrapError(err, funcTag, "running command")
	}
	logrus.Infof("Snapped %s", outFilePath)

	// if upload required, call the upload command!
	if opts.UploadAfterSuccess {
		uOpts := &UploadCmdOptions{
			InDir:               opts.OutDir,
			InFile:              outFileName,
			CleanupAfterSuccess: opts.CleanupAfterUpload,
		}
		logrus.Infof("Running Upload Command %+v", uOpts)
		err = UploadCmdRunE(ropts, uOpts)
		if err != nil {
			return util.WrapError(err, funcTag, "uploading after success")
		}
	}

	// done
	return nil
}
