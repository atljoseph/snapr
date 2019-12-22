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

// S3Accessor describes how to access a bucket in aws s3
type S3Accessor struct {
	Bucket string
	Region string
	Token  string
	Secret string
}

// NewS3Client gets a new AWS session in a structured way
func NewS3Client(config *S3Accessor) (*session.Session, *s3.S3, error) {
	funcTag := "NewS3Client"

	// new AWS config
	cfg := aws.NewConfig().
		WithRegion(config.Region).
		WithCredentials(credentials.NewStaticCredentials(config.Token, config.Secret, ""))

	// get a new AWS session
	sesh, err := session.NewSession(cfg)
	if err != nil {
		return nil, nil, WrapError(err, funcTag, "opening aws session")
	}

	return sesh, s3.New(sesh), nil
}

// CheckS3FileExists confirms that a file exists in an AWS S3
func CheckS3FileExists(s3Client *s3.S3, bucket, key string) (bool, error) {
	funcTag := "CheckS3FileExists"

	// logrus.Infof("Check Key: %s", key)

	_, err := s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, WrapError(err, funcTag, "checking aws file")
	}

	return true, nil
}

// SendToS3 sends a single file to an AWS S3 bucket
func SendToS3(s3Client *s3.S3, bucket, baseDirPth string, waffle WalkedFile) (string, error) {
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
		Bucket:             aws.String(bucket),
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

// S3Delimiter is the folder delimiter (for us) in AWS S3
var S3Delimiter = "/"

// S3ObjectsByKey sends a single file to an AWS S3 bucket
func S3ObjectsByKey(s3Client *s3.S3, key string) ([]*s3.Object, []string, error) {
	funcTag := "S3ObjectsByKey"

	// build the input
	query := &s3.ListObjectsV2Input{
		Bucket:    aws.String(EnvVarString("S3_BUCKET", "")),
		Prefix:    aws.String(key),
		Delimiter: aws.String("/"),
	}

	logrus.Infof("Fetching from: %s::%s", *query.Bucket, *query.Prefix)

	var files []*s3.Object
	var folders []string

	// syncronously cycle through until all are returned (not truncated response)
	for {

		// get the list of contents
		response, err := s3Client.ListObjectsV2(query)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case s3.ErrCodeNoSuchBucket:
					return files, folders, WrapError(aerr, funcTag, s3.ErrCodeNoSuchBucket)
				default:
					return files, folders, WrapError(aerr, funcTag, "unspecified error; ok")
				}
			} else {
				return files, folders, WrapError(aerr, funcTag, "unspecified error; not ok")
			}
		}
		// logrus.Infof("Fetched %+v", response)

		// yank the files
		for _, file := range response.Contents {
			files = append(files, file)
		}

		// yank the directories
		for _, dir := range response.CommonPrefixes {
			folders = append(folders, *dir.Prefix)
		}

		logrus.Infof("Fetched %d results", len(response.Contents))

		// set the continuation token and handle break on truncation
		query.ContinuationToken = response.NextContinuationToken

		// if truncated, break with all the results
		if !*response.IsTruncated {
			logrus.Infof("Done fetching. %d total results", len(files))
			break
		}
	}

	// return if no error
	return files, folders, nil
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
