/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package storage

import (
	"bytes"
	"io/ioutil"
	pathutil "path"
	"strings"
	"net/http"
	"crypto/tls"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// AmazonS3Backend is a storage backend for Amazon S3
type AmazonS3Backend struct {
	Bucket     string
	Client     *s3.S3
	Downloader *s3manager.Downloader
	Prefix     string
	Uploader   *s3manager.Uploader
	SSE        string
}

// NewAmazonS3Backend creates a new instance of AmazonS3Backend
func NewAmazonS3Backend(bucket string, prefix string, region string, endpoint string, sse string) *AmazonS3Backend {
	client := http.DefaultClient
	if os.Getenv("AWS_INSECURE_SKIP_VERIFY") == "true" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}
	service := s3.New(session.New(), &aws.Config{
		HTTPClient:       client,
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(strings.HasPrefix(endpoint, "http://")),
		S3ForcePathStyle: aws.Bool(endpoint != ""),
	})
	b := &AmazonS3Backend{
		Bucket:     bucket,
		Client:     service,
		Downloader: s3manager.NewDownloaderWithClient(service),
		Prefix:     cleanPrefix(prefix),
		Uploader:   s3manager.NewUploaderWithClient(service),
		SSE:        sse,
	}
	return b
}

type AmazonS3Options struct {
	S3ForcePathStyle *bool
}

func NewAmazonS3BackendWithOptions(bucket string, prefix string, region string, endpoint string, sse string, options *AmazonS3Options) *AmazonS3Backend {
	client := http.DefaultClient
	if os.Getenv("AWS_INSECURE_SKIP_VERIFY") == "true" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}
	s3ForcePathStyle := endpoint != ""
	if options != nil && options.S3ForcePathStyle != nil {
		s3ForcePathStyle = *options.S3ForcePathStyle
	}
	service := s3.New(session.New(), &aws.Config{
		HTTPClient:       client,
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(strings.HasPrefix(endpoint, "http://")),
		S3ForcePathStyle: aws.Bool(s3ForcePathStyle),
	})
	b := &AmazonS3Backend{
		Bucket:     bucket,
		Client:     service,
		Downloader: s3manager.NewDownloaderWithClient(service),
		Prefix:     cleanPrefix(prefix),
		Uploader:   s3manager.NewUploaderWithClient(service),
		SSE:        sse,
	}
	return b
}

// NewAmazonS3BackendWithCredentials creates a new instance of AmazonS3Backend with credentials
func NewAmazonS3BackendWithCredentials(bucket string, prefix string, region string, endpoint string, sse string, credentials *credentials.Credentials) *AmazonS3Backend {
	client := http.DefaultClient
	if os.Getenv("AWS_INSECURE_SKIP_VERIFY") == "true" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}
	service := s3.New(session.New(), &aws.Config{
		HTTPClient:       client,
		Credentials:      credentials,
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(strings.HasPrefix(endpoint, "http://")),
		S3ForcePathStyle: aws.Bool(endpoint != ""),
	})
	b := &AmazonS3Backend{
		Bucket:     bucket,
		Client:     service,
		Downloader: s3manager.NewDownloaderWithClient(service),
		Prefix:     cleanPrefix(prefix),
		Uploader:   s3manager.NewUploaderWithClient(service),
		SSE:        sse,
	}
	return b
}

// ListObjects lists all objects in Amazon S3 bucket, at prefix
func (b AmazonS3Backend) ListObjects(prefix string) ([]Object, error) {
	var objects []Object
	prefix = pathutil.Join(b.Prefix, prefix)
	s3Input := &s3.ListObjectsInput{
		Bucket: aws.String(b.Bucket),
		Prefix: aws.String(prefix),
	}
	for {
		s3Result, err := b.Client.ListObjects(s3Input)
		if err != nil {
			return objects, err
		}
		for _, obj := range s3Result.Contents {
			path := removePrefixFromObjectPath(prefix, *obj.Key)
			if objectPathIsInvalid(path) {
				continue
			}
			object := Object{
				Path:         path,
				Content:      []byte{},
				LastModified: *obj.LastModified,
			}
			objects = append(objects, object)
		}
		if !*s3Result.IsTruncated {
			break
		}
		s3Input.Marker = s3Result.Contents[len(s3Result.Contents)-1].Key
	}
	return objects, nil
}

// GetObject retrieves an object from Amazon S3 bucket, at prefix
func (b AmazonS3Backend) GetObject(path string) (Object, error) {
	var object Object
	object.Path = path
	var content []byte
	s3Input := &s3.GetObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(pathutil.Join(b.Prefix, path)),
	}
	s3Result, err := b.Client.GetObject(s3Input)
	if err != nil {
		return object, err
	}
	content, err = ioutil.ReadAll(s3Result.Body)
	if err != nil {
		return object, err
	}
	object.Content = content
	object.LastModified = *s3Result.LastModified
	return object, nil
}

// PutObject uploads an object to Amazon S3 bucket, at prefix
func (b AmazonS3Backend) PutObject(path string, content []byte) error {
	s3Input := &s3manager.UploadInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(pathutil.Join(b.Prefix, path)),
		Body:   bytes.NewBuffer(content),
	}

	if b.SSE != "" {
		s3Input.ServerSideEncryption = aws.String(b.SSE)
	}

	_, err := b.Uploader.Upload(s3Input)
	return err
}

// DeleteObject removes an object from Amazon S3 bucket, at prefix
func (b AmazonS3Backend) DeleteObject(path string) error {
	s3Input := &s3.DeleteObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(pathutil.Join(b.Prefix, path)),
	}
	_, err := b.Client.DeleteObject(s3Input)
	return err
}
