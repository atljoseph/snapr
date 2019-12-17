package cli

import (
	"bytes"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// PidgeonCmdOpts options
type PidgeonCmdOpts struct {
	BaseReadDir string
	InFileName  string
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
	pidgeonCmd.Flags().StringVar(&pidgeonCmdOpts.BaseReadDir, "base-dir", "~/", "Base Directory")
	pidgeonCmd.Flags().StringVar(&pidgeonCmdOpts.InFileName, "in-file", "test.jpg", "Input File")
}

func upload(cmd *cobra.Command, args []string) error {
	funcTag := "pidgeon"
	logrus.Infof("Pidgeoning")

	// TODO: get a list of files to upload to the bucket
	// based on the base dir, etc
	// ignore files with specific filename format
	// if too many files, give up
	// after success, rename the file

	// new AWS config
	cfg := aws.NewConfig().
		WithRegion(os.Getenv("S3_REGION")).
		WithCredentials(credentials.NewStaticCredentials(os.Getenv("S3_TOKEN"), os.Getenv("S3_SECRET"), ""))

	// get a new AWS session
	s, err := session.NewSession(cfg)
	if err != nil {
		return wrapError(err, funcTag, "get new aws session")
	}

	// Open the file for use
	inFilePath := pidgeonCmdOpts.InFileName
	if len(pidgeonCmdOpts.BaseReadDir) > 0 {
		inFilePath = pidgeonCmdOpts.BaseReadDir + "/" + pidgeonCmdOpts.InFileName
	}
	file, err := os.Open(inFilePath)
	if err != nil {
		return wrapError(err, funcTag, "opening file")
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(os.Getenv("S3_BUCKET")),
		Key:                aws.String(fileInfo.Name()),
		ACL:                aws.String("private"),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(size),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
		// ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		return wrapError(err, funcTag, "uploading file")
	}

	// done
	logrus.Infof("Done")
	return nil
}
