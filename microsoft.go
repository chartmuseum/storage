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
	"errors"
	"io/ioutil"
	pathutil "path"
	"time"

	"os"

	microsoft_storage "github.com/Azure/azure-sdk-for-go/storage"
)

// MicrosoftBlobBackend is a storage backend for Microsoft Azure Blob Storage
type MicrosoftBlobBackend struct {
	Prefix    string
	Container *microsoft_storage.Container
}

// NewMicrosoftBlobBackend creates a new instance of MicrosoftBlobBackend
func NewMicrosoftBlobBackend(container string, prefix string) *MicrosoftBlobBackend {

	// From the Azure portal, get your storage account name and key and set environment variables.
	accountName, accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT"), os.Getenv("AZURE_STORAGE_ACCESS_KEY")
	var serviceBaseURL, apiVersion string
	if serviceBaseURL = os.Getenv("AZURE_BASE_URL"); serviceBaseURL == "" {
		serviceBaseURL = microsoft_storage.DefaultBaseURL
	}
	if apiVersion = os.Getenv("AZURE_API_VERSION"); apiVersion == "" {
		apiVersion = microsoft_storage.DefaultAPIVersion
	}
	if len(accountName) == 0 || len(accountKey) == 0 {
		panic("Either the AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY environment variable is not set")
	}

	client, err := microsoft_storage.NewClient(accountName, accountKey, serviceBaseURL, apiVersion, true)
	if err != nil {
		panic(err)
	}

	blobClient := client.GetBlobService()
	containerRef := blobClient.GetContainerReference(container)

	b := &MicrosoftBlobBackend{
		Prefix:    prefix,
		Container: containerRef,
	}

	return b
}

// ListObjects lists all objects in Microsoft Azure Blob Storage container
func (b MicrosoftBlobBackend) ListObjects(prefix string) ([]Object, error) {
	var objects []Object

	if b.Container == nil {
		return objects, errors.New("Unable to obtain a container reference.")
	}

	var params microsoft_storage.ListBlobsParameters
	prefix = pathutil.Join(b.Prefix, prefix)
	params.Prefix = prefix

	for {
		response, err := b.Container.ListBlobs(params)
		if err != nil {
			return objects, err
		}

		for _, blob := range response.Blobs {
			path := removePrefixFromObjectPath(prefix, blob.Name)
			if objectPathIsInvalid(path) {
				continue
			}

			object := Object{
				Path:         path,
				Content:      []byte{},
				LastModified: time.Time(blob.Properties.LastModified),
			}

			objects = append(objects, object)
		}

		if response.NextMarker == "" {
			break
		}

		params.Marker = response.NextMarker
	}

	return objects, nil
}

// GetObject retrieves an object from Microsoft Azure Blob Storage, at path
func (b MicrosoftBlobBackend) GetObject(path string) (Object, error) {
	var object Object
	object.Path = path

	if b.Container == nil {
		return object, errors.New("Unable to obtain a container reference.")
	}

	var content []byte

	blobReference := b.Container.GetBlobReference(pathutil.Join(b.Prefix, path))
	exists, err := blobReference.Exists()
	if err != nil {
		return object, err
	}

	if !exists {
		return object, errors.New("Object does not exist.")
	}

	readCloser, err := blobReference.Get(nil)
	if err != nil {
		return object, err
	}

	content, err = ioutil.ReadAll(readCloser)
	if err != nil {
		return object, err
	}

	object.Content = content
	err = blobReference.GetProperties(nil)
	object.LastModified = time.Time(blobReference.Properties.LastModified)
	return object, nil
}

// PutObject uploads an object to Microsoft Azure Blob Storage container, at path
func (b MicrosoftBlobBackend) PutObject(path string, content []byte) error {
	if b.Container == nil {
		return errors.New("Unable to obtain a container reference.")
	}

	blobReference := b.Container.GetBlobReference(pathutil.Join(b.Prefix, path))

	err := blobReference.PutAppendBlob(nil)
	if err == nil {
		err = blobReference.AppendBlock(content, nil)
	}

	return err
}

// DeleteObject removes an object from Microsoft Azure Blob Storage container, at path
func (b MicrosoftBlobBackend) DeleteObject(path string) error {
	if b.Container == nil {
		return errors.New("Unable to obtain a container reference.")
	}

	blobReference := b.Container.GetBlobReference(pathutil.Join(b.Prefix, path))
	_, err := blobReference.DeleteIfExists(nil)
	return err
}
