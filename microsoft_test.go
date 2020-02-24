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
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MicrosoftTestSuite struct {
	suite.Suite
	BrokenAzureBlobBackend   *MicrosoftBlobBackend
	NoPrefixAzureBlobBackend *MicrosoftBlobBackend
}

func (suite *MicrosoftTestSuite) SetupSuite() {
	backend := NewMicrosoftBlobBackend("fake-container-cant-exist-fbce123", "")
	suite.BrokenAzureBlobBackend = backend

	containerName := os.Getenv("TEST_STORAGE_AZURE_CONTAINER")
	backend = NewMicrosoftBlobBackend(containerName, "")
	suite.NoPrefixAzureBlobBackend = backend

	data := []byte("some object")
	path := "deleteme.txt"
	err := suite.NoPrefixAzureBlobBackend.PutObject(path, data)
	suite.Nil(err, "no error putting deleteme.txt using Azure backend")
}

func (suite *MicrosoftTestSuite) TearDownSuite() {
	err := suite.NoPrefixAzureBlobBackend.DeleteObject("deleteme.txt")
	suite.Nil(err, "no error deleting deleteme.txt using Azure backend")
}

func (suite *MicrosoftTestSuite) TestListObjects() {
	_, err := suite.BrokenAzureBlobBackend.ListObjects("")
	suite.NotNil(err, "cannot list objects with bad bucket")

	_, err = suite.NoPrefixAzureBlobBackend.ListObjects("")
	suite.Nil(err, "can list objects with good bucket, no prefix")
}

func (suite *MicrosoftTestSuite) TestListObjectsWithPaging() {
	// create 5001 objects to trigger the need of paging, since
	// the default page size for Azure Blobs are 5000.
	data := []byte("some object")
	for i := 0; i < 5001; i++ {
		path := fmt.Sprintf("deleteme-%d.txt", i)
		suite.NoPrefixAzureBlobBackend.PutObject(path, data)
	}

	// check if 5002 (5001 plus deleteme.txt from SetupSuite()) objects
	// are returned. Without paging, we would get only 5000.
	res, err := suite.NoPrefixAzureBlobBackend.ListObjects("")
	suite.Nil(err, "can list objects with good bucket, no prefix")
	suite.Equal(5001, len(res))

	// clean up
	for i := 0; i < 5001; i++ {
		path := fmt.Sprintf("deleteme-%d.txt", i)
		suite.NoPrefixAzureBlobBackend.DeleteObject(path)
	}
}

func (suite *MicrosoftTestSuite) TestGetObject() {
	_, err := suite.BrokenAzureBlobBackend.GetObject("this-file-cannot-possibly-exist.tgz")
	suite.NotNil(err, "cannot get objects with bad bucket")
}

func (suite *MicrosoftTestSuite) TestPutObject() {
	err := suite.BrokenAzureBlobBackend.PutObject("this-file-will-not-upload.txt", []byte{})
	suite.NotNil(err, "cannot put objects with bad bucket")
}

func TestAzureStorageTestSuite(t *testing.T) {
	if os.Getenv("TEST_CLOUD_STORAGE") == "1" &&
		os.Getenv("TEST_STORAGE_AZURE_CONTAINER") != "" {
		suite.Run(t, new(MicrosoftTestSuite))
	}
}
