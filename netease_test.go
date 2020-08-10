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

type NeteaseTestSuite struct {
	suite.Suite
	BrokenNeteaseNOSBackend   *NeteaseNOSBackend
	NoPrefixNeteaseNOSBackend *NeteaseNOSBackend
}

const nosTestCount = 100

func (suite *NeteaseTestSuite) SetupSuite() {
	backend := NewNeteaseNOSBackend("fake-container-cant-exist-fbce123", "", "")
	suite.BrokenNeteaseNOSBackend = backend

	nosBucket := os.Getenv("TEST_STORAGE_NETEASE_BUCKET")
	nosEndpoint := os.Getenv("TEST_STORAGE_NETEASE_ENDPOINT")
	backend = NewNeteaseNOSBackend(nosBucket, "", nosEndpoint)
	suite.NoPrefixNeteaseNOSBackend = backend

	data := []byte("some object")

	for i := 0; i < nosTestCount; i++ {
		err := suite.NoPrefixNeteaseNOSBackend.PutObject(fmt.Sprintf("%ddeleteme.txt", i), data)
		suite.Nil(err, "no error putting deleteme.txt using Netease Cloud NOS backend")
		err = suite.NoPrefixNeteaseNOSBackend.PutObject(fmt.Sprintf("testdir%d/deleteme.txt", i), data)
		suite.Nil(err, "no error putting testdir/deleteme.txt using Netease Cloud NOS backend")
	}
}

func (suite *NeteaseTestSuite) TearDownSuite() {
	for i := 0; i < bosTestCount; i++ {
		err := suite.NoPrefixNeteaseNOSBackend.DeleteObject(fmt.Sprintf("%ddeleteme.txt", i))
		suite.Nil(err, "no error deleting deleteme.txt using Netease NOS backend")
		err = suite.NoPrefixNeteaseNOSBackend.DeleteObject(fmt.Sprintf("testdir%d/deleteme.txt", i))
		suite.Nil(err, "no error deleting deleteme.txt using Netease NOS backend")
	}
}

func (suite *NeteaseTestSuite) TestListObjects() {
	_, err := suite.BrokenNeteaseNOSBackend.ListObjects("")
	suite.NotNil(err, "cannot list objects with bad bucket")

	objs, err := suite.NoPrefixNeteaseNOSBackend.ListObjects("")
	suite.Nil(err, "can list objects with good bucket, no prefix")
	suite.Equal(len(objs), nosTestCount, "able to list objects")
}

func (suite *NeteaseTestSuite) TestListFolders() {
	_, err := suite.BrokenNeteaseNOSBackend.ListFolders("")
	suite.NotNil(err, "cannot list folders with bad bucket")

	folders, err := suite.NoPrefixNeteaseNOSBackend.ListFolders("")
	suite.Nil(err, "can list folders with good bucket, no prefix")
	suite.Equal(len(folders), nosTestCount, "able to list folders")
}

func (suite *NeteaseTestSuite) TestGetObject() {
	_, err := suite.BrokenNeteaseNOSBackend.GetObject("this-file-cannot-possibly-exist.tgz")
	suite.NotNil(err, "cannot get objects with bad bucket")
}

func (suite *NeteaseTestSuite) TestPutObject() {
	err := suite.BrokenNeteaseNOSBackend.PutObject("this-file-will-not-upload.txt", []byte{})
	suite.NotNil(err, "cannot put objects with bad bucket")
}

func TestNeteaseStorageTestSuite(t *testing.T) {
	if os.Getenv("TEST_CLOUD_STORAGE") == "1" &&
		os.Getenv("TEST_STORAGE_NETEASE_BUCKET") != "" &&
		os.Getenv("TEST_STORAGE_NETEASE_ENDPOINT") != "" {
		suite.Run(t, new(NeteaseTestSuite))
	}
}
