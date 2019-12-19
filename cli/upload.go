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

// UploadCmdOptions options
type UploadCmdOptions struct {
	BaseReadDir         string
	InFileName          string
	CleanupAfterSuccess bool
}

// upload command
var (
	uploadCmdOpts = UploadCmdOptions{}
	uploadCmd     = &cobra.Command{
		Use:   "upload",
		Short: "Snapr is a snapper turtle.",
		Long:  `Do you like turtles?`,
		RunE: func(cmd *cobra.Command, args []string) error {
			uploadCmdOpts = uploadCmdTransformPositionalArgs(args, uploadCmdOpts)
			return RunUploadCmdE(uploadCmdOpts)
		},
	}
)

// uploadCmdTransformPositionalArgs adds the positional string args
// from the command to the options struct (for DI)
// care should be taken to not use the same options here as in flags, etc
func uploadCmdTransformPositionalArgs(args []string, opts UploadCmdOptions) UploadCmdOptions {
	// if len(args) > 0 {
	// // can use env vars, too!
	// 	opts.Something = args[0]
	// }
	return opts
}

func init() {
	// add command to root
	rootCmd.AddCommand(uploadCmd)

	// this is where the files are pulled from
	uploadCmd.Flags().StringVar(&uploadCmdOpts.BaseReadDir, "upload-dir", getEnvVarString("UPLOAD_DIR", "/"), "Base Directory")
	// uploadCmd.Flags().BoolVar(&UploadCmdOpts.CleanupAfterSuccess, "dir", true, "Base Directory")
	uploadCmd.Flags().StringVar(&uploadCmdOpts.InFileName, "upload-file", getEnvVarString("UPLOAD_FILE", "test.jpg"), "Input File")
}

func RunUploadCmdE(opts UploadCmdOptions) error {
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
	inFilePath := opts.InFileName
	if len(opts.BaseReadDir) > 0 {
		inFilePath = opts.BaseReadDir + "/" + opts.InFileName
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
