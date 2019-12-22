package util

import (
	"bytes"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
)

// NewS3Client gets a new AWS session in a structured way
func NewS3Client() (*session.Session, *s3.S3, error) {
	funcTag := "NewS3Client"

	// new AWS config
	cfg := aws.NewConfig().
		WithRegion(EnvVarString("S3_REGION", "")).
		WithCredentials(credentials.NewStaticCredentials(
			EnvVarString("S3_TOKEN", ""), EnvVarString("S3_SECRET", ""), ""))

	// get a new AWS session
	sesh, err := session.NewSession(cfg)
	if err != nil {
		return nil, nil, WrapError(err, funcTag, "opening aws session")
	}

	return sesh, s3.New(sesh), nil
}

// CheckS3FileExists confirms that a file exists in an AWS S3
func CheckS3FileExists(s3Client *s3.S3, key string) (bool, error) {
	funcTag := "CheckS3FileExists"

	// logrus.Infof("Check Key: %s", key)

	_, err := s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(EnvVarString("S3_BUCKET", "")),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, WrapError(err, funcTag, "checking aws file")
	}

	return true, nil
}

// SendToS3 sends a single file to an AWS S3 bucket
func SendToS3(s3Client *s3.S3, baseDirPth string, waffle WalkedFile) (string, error) {
	funcTag := "SendToS3"

	// Open the file for use
	file, err := os.Open(waffle.Path)
	if err != nil {
		return "", WrapError(err, funcTag, "opening file")
	}
	defer file.Close()

	// determine the aws path key
	key := strings.ReplaceAll(waffle.Path, baseDirPth+"/", "")

	// Get file size and read the file content into a buffer
	fileSize := waffle.FileInfo.Size()
	buffer := make([]byte, fileSize)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(EnvVarString("S3_BUCKET", "")),
		Key:                aws.String(key),
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
	return key, nil
}

// func DownloadBucket

// S3ObjectsByKey sends a single file to an AWS S3 bucket
func S3ObjectsByKey(s3Client *s3.S3, key string) ([]*s3.Object, error) {
	funcTag := "S3ObjectsByKey"

	// build the input
	query := &s3.ListObjectsV2Input{
		Bucket: aws.String(EnvVarString("S3_BUCKET", "")),
		Prefix: aws.String(key),
	}

	var results []*s3.Object

	// syncronously cycle through until all are returned (not truncated response)
	for {

		// get the list of contents
		response, err := s3Client.ListObjectsV2(query)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case s3.ErrCodeNoSuchBucket:
					return results, WrapError(aerr, funcTag, s3.ErrCodeNoSuchBucket)
				default:
					return results, WrapError(aerr, funcTag, "unspecified error; ok")
				}
			} else {
				return results, WrapError(aerr, funcTag, "unspecified error; not ok")
			}
		}

		// append the results
		for _, o := range response.Contents {
			results = append(results, o)
		}

		logrus.Infof("Fetched %d results", len(response.Contents))

		// set the continuation token and handle break on truncation
		query.ContinuationToken = response.NextContinuationToken

		// if truncated, break with all the results
		if !*response.IsTruncated {
			logrus.Infof("Done fetching. %d total results", len(results))
			break
		}
	}

	// return if no error
	return results, nil
}

// DownloadS3Object downaloads a single object from aws s3 bucket
func DownloadS3Object(sesh *session.Session, key string) ([]byte, error) {
	funcTag := "DownloadS3Object"
	buff := &aws.WriteAtBuffer{}
	downloader := s3manager.NewDownloader(sesh)
	query := &s3.GetObjectInput{
		Bucket: aws.String(EnvVarString("S3_BUCKET", "")),
		Key:    aws.String(key),
	}
	_, err := downloader.Download(buff, query)
	if err != nil {
		return []byte{}, WrapError(err, funcTag, "dowloading to buffer")
	}

	return buff.Bytes(), nil
}
