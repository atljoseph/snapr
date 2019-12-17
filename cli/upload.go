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

// UploadCmdOpts options
type UploadCmdOpts struct {
	BaseReadDir string
	InFileName  string
}

// upload command
var (
	uploadCmdOpts = UploadCmdOpts{}
	uploadCmd     = &cobra.Command{
		Use:   "upload",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE:  upload,
	}
)

func init() {
	// add command to root
	rootCmd.AddCommand(uploadCmd)

	// this is where the files are pulled from
	uploadCmd.Flags().StringVar(&uploadCmdOpts.BaseReadDir, "base-dir", "~", "Base Directory")
	uploadCmd.Flags().StringVar(&uploadCmdOpts.InFileName, "in-file", "test.jpg", "Input File")
}

func upload(cmd *cobra.Command, args []string) error {
	funcTag := "upload"
	logrus.Infof("Uploading")

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
	inFilePath := uploadCmdOpts.InFileName
	if len(uploadCmdOpts.BaseReadDir) > 0 {
		inFilePath = uploadCmdOpts.BaseReadDir + "/" + uploadCmdOpts.InFileName
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
