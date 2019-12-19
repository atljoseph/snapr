package util

import (
	"bytes"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

// NewAwsSession gets a new AWS session in a structured way
func NewAwsSession() (*session.Session, error) {
	// new AWS config
	cfg := aws.NewConfig().
		WithRegion(os.Getenv("S3_REGION")).
		WithCredentials(credentials.NewStaticCredentials(os.Getenv("S3_TOKEN"), os.Getenv("S3_SECRET"), ""))

	// get a new AWS session
	return session.NewSession(cfg)
}

// CheckAwsFileExists confirms that a file exists in an AWS S3
func CheckAwsFileExists(s *session.Session, key string) (bool, error) {
	funcTag := "CheckAwsFileExists"

	logrus.Infof("Check Key: %s", key)

	_, err := s3.New(s).HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, WrapError(err, funcTag, "checking aws file")
	}

	return true, nil
}

// SendToAws sends a single file to an AWS S3 bucket
func SendToAws(s *session.Session, filePath string) (string, error) {
	funcTag := "SendToAws"

	// Open the file for use
	file, err := os.Open(filePath)
	if err != nil {
		return "", WrapError(err, funcTag, "opening file")
	}
	defer file.Close()

	// stat the file for some info
	fileInfo, err := file.Stat()
	if err != nil {
		return "", WrapError(err, funcTag, "stat file")
	}

	// Get file size and read the file content into a buffer
	fileSize := fileInfo.Size()
	buffer := make([]byte, fileSize)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(os.Getenv("S3_BUCKET")),
		Key:                aws.String(fileInfo.Name()),
		ACL:                aws.String("private"),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(fileSize),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
		// ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		return "", WrapError(err, funcTag, "uploading file")
	}

	// return if no error
	return fileInfo.Name(), nil
}
