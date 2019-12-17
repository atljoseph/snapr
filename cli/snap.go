package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// SnapCmdOpts options
type SnapCmdOpts struct {
	BaseOutDir        string
	CaptureDeviceAddr string
	OutFilePathParam  string
}

// snap command
var (
	snapCmOpts = SnapCmdOpts{}
	snapCmd    = &cobra.Command{
		Use:   "snap",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE:  snap,
	}
)

func init() {
	// add command to root
	rootCmd.AddCommand(snapCmd)

	// "default" device is the default
	snapCmd.Flags().StringVarP(&snapCmOpts.CaptureDeviceAddr, "capture-device", "c", "default", "Device Address")

	// if this is not set, the base dir is the output directory
	snapCmd.Flags().StringVarP(&snapCmOpts.OutFilePathParam, "extra-dir", "e", "", "Output File Path Parameter")

	// this is where the files go
	snapCmd.Flags().StringVarP(&snapCmOpts.BaseOutDir, "base-dir", "d", "~/", "Device Address")
}

func snap(cmd *cobra.Command, args []string) error {
	funcTag := "snap"
	logrus.Infof("Snapping")

	// get the logged in users list
	usersExec := exec.Command("/bin/sh", "-c", "users")
	usersBytes, err := usersExec.Output()
	if err != nil {
		return fmt.Errorf("(%s) => (%s) %s", funcTag, "get users list", err)
	}
	usersOutput := bytes.NewBuffer(usersBytes).String()
	// remove line breaks
	usersOutput = strings.ReplaceAll(usersOutput, "\n", "")
	// replace spaces ith "-"
	usersOutput = strings.ReplaceAll(usersOutput, " ", "-")

	// build the os cmd to execute
	timeFormat := "2006-01-02T15-04-05"
	outFileExt := "jpg"
	outFileName := fmt.Sprintf("%s-%s.%s", usersOutput, time.Now().Format(timeFormat), outFileExt)
	outFileDir := snapCmOpts.BaseOutDir
	if len(snapCmOpts.OutFilePathParam) > 0 {
		outFileDir += "/" + snapCmOpts.OutFilePathParam
	}
	outFilePath := fmt.Sprintf("%s/%s", outFileDir, outFileName)

	// ensure output dir exists
	err = os.MkdirAll(outFileDir, os.ModeDir)
	if err != nil {
		return fmt.Errorf("(%s) => (%s) %s", funcTag, "mkdir for outFileDir", err)
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
	if len(snapCmOpts.CaptureDeviceAddr) > 0 {
		webcamAddr += snapCmOpts.CaptureDeviceAddr
	} else {
		if strings.EqualFold(runtime.GOOS, "linux") {
			webcamAddr += "/dev/video0"
		}
		if strings.EqualFold(runtime.GOOS, "darwin") {
			webcamAddr += "default"
		}
	}

	// capture command with ffmpeg
	// overwrite existing file if any
	ffmpegExecString := fmt.Sprintf("ffmpeg %s %s %s %s %s %s -y", driverType, framerate, resolution, webcamAddr, vframes, outFilePath)
	// logrus.Infof("Command: %s", ffmpegExecString)
	ffmpegExec := exec.Command("/bin/sh", "-c", ffmpegExecString)

	// execute and wait
	err = ffmpegExec.Run()
	if err != nil {
		return fmt.Errorf("(%s) => (%s) %s", funcTag, "running command", err)
	}

	// done
	logrus.Infof("Snapped %s", outFilePath)
	return nil
}
