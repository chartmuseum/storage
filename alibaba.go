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
	"net/http"
	"os"
	pathutil "path"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// AlibabaCloudOSSBackend is a storage backend for Alibaba Cloud OSS
type AlibabaCloudOSSBackend struct {
	Bucket *oss.Bucket
	Client *oss.Client
	Prefix string
	SSE    string
}

// NewAlibabaCloudOSSBackend creates a new instance of AlibabaCloudOSSBackend
func NewAlibabaCloudOSSBackend(bucket string, prefix string, endpoint string, sse string) *AlibabaCloudOSSBackend {
	accessKeyId := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")

	if len(accessKeyId) == 0 {
		panic("ALIBABA_CLOUD_ACCESS_KEY_ID environment variable is not set")
	}

	if len(accessKeySecret) == 0 {
		panic("ALIBABA_CLOUD_ACCESS_KEY_SECRET environment variable is not set")
	}

	if len(endpoint) == 0 {
		// Set default endpoint
		endpoint = "oss-cn-hangzhou.aliyuncs.com"
	}

	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)

	if err != nil {
		panic("Failed to create OSS client: " + err.Error())
	}

	ossBucket, err := client.Bucket(bucket)
	if err != nil {
		panic("Failed to get bucket: " + err.Error())
	}

	b := &AlibabaCloudOSSBackend{
		Bucket: ossBucket,
		Client: client,
		Prefix: cleanPrefix(prefix),
		SSE:    sse,
	}
	return b
}


// ListObjects lists all objects in Alibaba Cloud OSS bucket, at prefix
func (b AlibabaCloudOSSBackend) ListObjects(prefix string) ([]Object, error) {
	return getObjects(b, prefix)
}

// ListFolders lists all folders in Alibaba Cloud OSS bucket, at prefix
func (b AlibabaCloudOSSBackend) ListFolders(prefix string) ([]Folder, error) {
	return getFolders(b, prefix)
}

func (b AlibabaCloudOSSBackend) ObjectIter(prefix string) <-chan Item {
	ch := make(chan(Item))
	go func() {
		prefix = pathutil.Join(b.Prefix, prefix)
		ossPrefix := oss.Prefix(prefix)
		marker := oss.Marker("")
		for {
			lor, err := b.Bucket.ListObjects(oss.MaxKeys(50), marker, ossPrefix)
			if err != nil {
				ch <- Item{nil, err}
				close(ch)
				return
			}
			for _, obj := range lor.Objects {
				path := removePrefixFromObjectPath(prefix, obj.Key)
				ch <- Item{
					&Object{
						Path:         path,
						Content:      []byte{},
						LastModified: obj.LastModified,
					}, nil }
			}
			if !lor.IsTruncated {
				break
			}
			ossPrefix = oss.Prefix(lor.Prefix)
			marker = oss.Marker(lor.NextMarker)
		}
		close(ch)
	} ()
	return ch
}

// GetObject retrieves an object from Alibaba Cloud OSS bucket, at prefix
func (b AlibabaCloudOSSBackend) GetObject(path string) (Object, error) {
	var object Object
	object.Path = path
	var content []byte
	key := pathutil.Join(b.Prefix, path)
	body, err := b.Bucket.GetObject(key)

	if err != nil {
		return object, err
	}
	content, err = ioutil.ReadAll(body)
	body.Close()
	if err != nil {
		return object, err
	}
	object.Content = content

	headers, err := b.Bucket.GetObjectMeta(key)
	if err != nil {
		return object, err
	}
	lastModified, _ := http.ParseTime(headers.Get(oss.HTTPHeaderLastModified))
	object.LastModified = lastModified
	return object, nil
}

// PutObject uploads an object to Alibaba Cloud OSS bucket, at prefix
func (b AlibabaCloudOSSBackend) PutObject(path string, content []byte) error {
	key := pathutil.Join(b.Prefix, path)
	var err error
	if b.SSE == "" {
		err = b.Bucket.PutObject(key, bytes.NewReader(content))
	} else {
		sse := oss.ServerSideEncryption(b.SSE)
		err = b.Bucket.PutObject(key, bytes.NewReader(content), sse)
	}
	return err
}

// DeleteObject removes an object from Alibaba Cloud OSS bucket, at prefix
func (b AlibabaCloudOSSBackend) DeleteObject(path string) error {
	key := pathutil.Join(b.Prefix, path)
	err := b.Bucket.DeleteObject(key)
	return err
}
