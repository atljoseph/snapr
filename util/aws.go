package util

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
		return nil, nil, WrapError(err, funcTag, "failed to open aws session")
	}

	return sesh, s3.New(sesh), nil
}

// CheckS3ObjectExists confirms that a file exists in an AWS S3
func CheckS3ObjectExists(s3Client *s3.S3, bucket, key string) (bool, error) {
	funcTag := "CheckS3ObjectExists"

	// logrus.Infof("Check Key: %s", key)

	_, err := s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, WrapError(err, funcTag, "failed check s3 object")
	}

	return true, nil
}

// WriteS3File sends a single file to an AWS S3 bucket
func WriteS3File(s3Client *s3.S3, bucket, acl, targetKey string, waffle *WalkedFile) (string, error) {
	funcTag := "WriteS3File"

	// Open the file for use
	file, err := os.Open(waffle.Path)
	if err != nil {
		return "", WrapError(err, funcTag, "failed to open file")
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileSize := waffle.FileInfo.Size()
	buffer := make([]byte, fileSize)
	file.Read(buffer)

	return WriteS3Bytes(s3Client, bucket, acl, targetKey, buffer)
}

// WriteS3Bytes sends a single file to an AWS S3 bucket
func WriteS3Bytes(s3Client *s3.S3, bucket, acl, targetKey string, buffer []byte) (string, error) {
	funcTag := "WriteS3Bytes"

	// default the acl
	if len(acl) == 0 {
		acl = "private"
	}

	// content length bneeds to be int64
	contentLength := int64(len(buffer))

	// build the query
	query := &s3.PutObjectInput{
		Bucket:             aws.String(bucket),
		Key:                aws.String(targetKey),
		ACL:                aws.String(acl),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(contentLength),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
		// ServerSideEncryption: aws.String("AES256"),
	}

	// logrus.Infof("QUERY: %+v", *query)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err := s3Client.PutObject(query)
	if err != nil {
		return "", WrapError(err, funcTag, "failed to upload file")
	}

	// return if no error
	return targetKey, nil
}

// S3Delimiter is the folder delimiter (for us) in AWS S3
var S3Delimiter = "/"

// ListS3ObjectsByKey sends a single file to an AWS S3 bucket
// objects are "files"
// commonKeys are "directories"
func ListS3ObjectsByKey(s3Client *s3.S3, bucket, key string, useDelimiter bool) ([]*S3Object, []*S3Directory, error) {
	funcTag := "ListS3ObjectsByKey"

	// build the input
	query := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	}

	// recursive search ? use delimiter
	if useDelimiter {
		query.Delimiter = aws.String(S3Delimiter)
	}

	logrus.Infof("Fetching from: %s::%s", *query.Bucket, *query.Prefix)

	var files []*S3Object
	var folders []*S3Directory

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

		// yank the files with extension
		for _, file := range response.Contents {
			files = append(files, &S3Object{
				Key:       *file.Key,
				Extension: strings.ReplaceAll(filepath.Ext(*file.Key), ".", ""),
			})
		}

		// yank the directories
		for _, dir := range response.CommonPrefixes {
			folders = append(folders, &S3Directory{
				Key: *dir.Prefix,
			})
		}

		logrus.Infof("Fetched %d results", len(response.Contents))

		// set the continuation token and handle break on truncation
		query.ContinuationToken = response.NextContinuationToken

		// logrus.Warnf("%+v, %s", response, err)

		// if truncated, break with all the results
		if !*response.IsTruncated {
			msg := fmt.Sprintf("Done fetching. %d files", len(files))
			if useDelimiter {
				fmt.Sprintf("%s, %d folders", msg, len(folders))
			}
			logrus.Infof(msg)
			break
		}
	}

	// return if no error
	return files, folders, nil
}

// DownloadS3Object downaloads a single object from aws s3 bucket
func DownloadS3Object(s3Client *s3.S3, bucket, key string) ([]byte, error) {
	funcTag := "DownloadS3Object"

	// basoically, get a new byte slice to write to
	buff := &aws.WriteAtBuffer{}

	// get a new s3manager downloader
	downloader := s3manager.NewDownloaderWithClient(s3Client)

	// build the query
	query := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	// download the object
	_, err := downloader.Download(buff, query)
	if err != nil {
		return []byte{}, WrapError(err, funcTag, "failed to download to buffer")
	}

	return buff.Bytes(), nil
}

// DeleteS3Object deletes an object from S3 and returns an error, if any
func DeleteS3Object(s3Client *s3.S3, bucket, key string) error {
	funcTag := "DownloadS3Object"

	// build the query
	query := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	// remove the object from the bucket
	_, err := s3Client.DeleteObject(query)
	if err != nil {
		return WrapError(err, funcTag, "failed to delete object")
	}

	return nil
}

// RenameS3Object renames an object in S3 and returns an error, if any
// This operation is made on the same bucket, although it could be made between 2 buckets
func RenameS3Object(s3Client *s3.S3, bucket, sourceKey, destKey string) error {
	funcTag := "RenameS3Object"

	// validate the rename
	if strings.EqualFold(sourceKey, destKey) {
		return WrapError(fmt.Errorf("validation error"), funcTag, "cannot rename object to the same key")
	}

	// build the query
	query := &s3.CopyObjectInput{
		Bucket:            aws.String(bucket),
		Key:               aws.String(destKey),
		CopySource:        aws.String(fmt.Sprintf("%s/%s", bucket, sourceKey)),
		MetadataDirective: aws.String("REPLACE"),
	}

	// copy the original object to a new key
	_, err := s3Client.CopyObject(query)
	if err != nil {
		return WrapError(err, funcTag, "failed to delete object")
	}

	// remove the original object from the bucket
	err = DeleteS3Object(s3Client, bucket, sourceKey)
	if err != nil {
		return WrapError(err, funcTag, "failed to delete original object after copying during rename operation")
	}

	return nil
}

// S3Object is a wrapper for an aws object
type S3Object struct {
	Bytes      []byte
	Base64     string
	Key        string
	Extension  string
	DisplayKey string
}

// S3Directory is a wrapper for an aws folder
type S3Directory struct {
	Key        string
	DisplayKey string
}
