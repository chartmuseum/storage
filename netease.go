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
	"os"
	pathutil "path"

	"time"

	"github.com/NetEase-Object-Storage/nos-golang-sdk/config"
	"github.com/NetEase-Object-Storage/nos-golang-sdk/logger"
	"github.com/NetEase-Object-Storage/nos-golang-sdk/model"
	"github.com/NetEase-Object-Storage/nos-golang-sdk/nosclient"
)

// NeteaseNOSBackend is a storage backend for Netease Cloud NOS
type NeteaseNOSBackend struct {
	Client nosclient.NosClient
	Bucket string
	Prefix string
}

// NewNeteaseNOSBackend creates a new instance of NeteaseNOSBackend
func NewNeteaseNOSBackend(bucket string, prefix string, endpoint string) *NeteaseNOSBackend {
	accessKeyId := os.Getenv("NETEASE_CLOUD_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("NETEASE_CLOUD_ACCESS_KEY_SECRET")

	if len(accessKeyId) == 0 {
		panic("NETEASE_CLOUD_ACCESS_KEY_ID environment variable is not set")
	}

	if len(accessKeySecret) == 0 {
		panic("NETEASE_CLOUD_ACCESS_KEY_SECRET environment variable is not set")
	}

	if len(endpoint) == 0 {
		// Set default endpoint
		endpoint = "nos-eastchina1.126.net"
	}

	conf := &config.Config{
		Endpoint:                    endpoint,
		AccessKey:                   accessKeyId,
		SecretKey:                   accessKeySecret,
		NosServiceConnectTimeout:    3,
		NosServiceReadWriteTimeout:  5,
		NosServiceMaxIdleConnection: 15,
		LogLevel:                    logger.LogLevel(logger.DEBUG),
		Logger:                      logger.NewDefaultLogger(),
	}

	client, err := nosclient.New(conf)
	if err != nil {
		panic("Failed to create NOS client: " + err.Error())
	}

	b := &NeteaseNOSBackend{
		Client: *client,
		Bucket: bucket,
		Prefix: prefix,
	}
	return b
}

// ListObjects lists all objects in Netease Cloud NOS bucket, at prefix
func (b NeteaseNOSBackend) ListObjects(prefix string) ([]Object, error) {
	var objects []Object

	prefix = pathutil.Join(b.Prefix, prefix)

	listRequest := &model.ListObjectsRequest{
		Bucket:    b.Bucket,
		Prefix:    prefix,
		Delimiter: "",
		Marker:    "",
		MaxKeys:   100,
	}

	for {
		var lor *model.ListObjectsResult
		lor, err := b.Client.ListObjects(listRequest)
		if err != nil {
			return objects, nil
		}

		for _, obj := range lor.Contents {
			path := removePrefixFromObjectPath(prefix, obj.Key)
			if objectPathIsInvalid(path) {
				continue
			}

			local, _ := time.LoadLocation("Local")
			// LastModified time layout in NOS is 2006-01-02T15:04:05 -0700
			t, _ := time.ParseInLocation("2006-01-02T15:04:05 -0700", obj.LastModified, local)
			object := Object{
				Path:         path,
				Content:      []byte{},
				LastModified: t,
			}
			objects = append(objects, object)
		}
		if !lor.IsTruncated {
			break
		}

	}

	return objects, nil
}

// GetObject retrieves an object from Netease Cloud NOS bucket, at prefix
func (b NeteaseNOSBackend) GetObject(path string) (Object, error) {
	var object Object
	object.Path = path
	var content []byte
	key := pathutil.Join(b.Prefix, path)

	objectRequest := &model.GetObjectRequest{
		Bucket: b.Bucket,
		Object: key,
	}

	var nosObject *model.NOSObject
	nosObject, err := b.Client.GetObject(objectRequest)
	if err != nil {
		return object, err
	}

	body := nosObject.Body
	content, err = ioutil.ReadAll(body)
	defer body.Close()
	if err != nil {
		return object, err
	}

	object.Content = content
	objectMetaRequest := &model.ObjectRequest{
		Bucket: b.Bucket,
		Object: key,
	}

	var meta *model.ObjectMetadata
	meta, err = b.Client.GetObjectMetaData(objectMetaRequest)
	if err != nil {
		return object, err
	}

	m := meta.Metadata
	// 	"Last-Modified" 是从nos获取的存储 最后修改时间 的key
	if t, ok := m["Last-Modified"]; ok {

		local, _ := time.LoadLocation("Local")
		// NOS的LastModified格式为 2019-04-18T16:55:39 +0800
		lastModified, _ := time.ParseInLocation("2006-01-02T15:04:05 -0700", t, local)
		object.LastModified = lastModified
	}

	return object, nil
}

// PutObject uploads an object to Netease Cloud NOS bucket, at prefix
func (b NeteaseNOSBackend) PutObject(path string, content []byte) error {
	key := pathutil.Join(b.Prefix, path)
	var err error

	metadata := &model.ObjectMetadata{
		Metadata:      map[string]string{},
		ContentLength: int64(len(content)),
	}

	putObjectRequest := &model.PutObjectRequest{
		Bucket:   b.Bucket,
		Object:   key,
		Body:     bytes.NewReader(content),
		Metadata: metadata,
	}
	_, err = b.Client.PutObjectByStream(putObjectRequest)
	return err
}

// DeleteObject removes an object from Netease Cloud NOS bucket, at prefix
func (b NeteaseNOSBackend) DeleteObject(path string) error {
	key := pathutil.Join(b.Prefix, path)

	objectRequest := &model.ObjectRequest{
		Bucket: b.Bucket,
		Object: key,
	}

	err := b.Client.DeleteObject(objectRequest)
	return err
}
