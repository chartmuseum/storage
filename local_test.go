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
	"time"

	"github.com/stretchr/testify/suite"
)

type LocalTestSuite struct {
	suite.Suite
	BrokenLocalFilesytemBackend *LocalFilesystemBackend
	BrokenTempDirectory string
	LocalFilesystemBackend *LocalFilesystemBackend
	TempDirectory string
}

func (suite *LocalTestSuite) SetupSuite() {
	timestamp := time.Now().Format("20060102150405")
	suite.BrokenTempDirectory = fmt.Sprintf(".test/storage-local/%s-broken", timestamp)
	defer os.RemoveAll(suite.BrokenTempDirectory)
	suite.TempDirectory = fmt.Sprintf(".test/storage-local/%s", timestamp)
	suite.BrokenLocalFilesytemBackend = NewLocalFilesystemBackend(suite.BrokenTempDirectory)
	suite.LocalFilesystemBackend = NewLocalFilesystemBackend(suite.TempDirectory)
	data := []byte("some object")
	err := suite.LocalFilesystemBackend.PutObject("deleteme.txt", data)
	suite.Nil(err, "no error putting deleteme.txt using Local backend")
	err = suite.LocalFilesystemBackend.PutObject("testdir/deleteme.txt", data)
	suite.Nil(err, "no error putting testdir/deleteme.txt using Local backend")
}

//func (suite *LocalTestSuite) TeardownSuite() {
//	err := suite.LocalFilesystemBackend.DeleteObject("deleteme.txt")
//	suite.Nil(err, "no error deleting deleteme.txt using Local backend")
//	err = suite.LocalFilesystemBackend.DeleteObject("testdir/deleteme.txt")
//	suite.Nil(err, "no error deleting testdir/deleteme.txt using Local backend")
//	os.RemoveAll(suite.BrokenTempDirectory)
//	os.RemoveAll(suite.TempDirectory)
//}

func (suite *LocalTestSuite) TestListObjects() {
	_, err := suite.BrokenLocalFilesytemBackend.ListObjects("")
	suite.Nil(err, "list objects does not return error if dir does not exist")

	objs, err := suite.LocalFilesystemBackend.ListObjects("")
	suite.Nil(err, "can list objects with good dir")
	suite.Equal(len(objs), 1, "able to list objects")
}

func (suite *LocalTestSuite) TestListFolders() {
	_, err := suite.BrokenLocalFilesytemBackend.ListFolders("")
	suite.Nil(err, "list folders does not return error if dir does not exist")

	folders, err := suite.LocalFilesystemBackend.ListFolders("")
	suite.Nil(err, "can list folders with good dir")
	suite.Equal(len(folders), 1, "able to list folders")
}

func (suite *LocalTestSuite) TestGetObject() {
	_, err := suite.BrokenLocalFilesytemBackend.GetObject("this-file-cannot-possibly-exist.tgz")
	suite.NotNil(err, "cannot get objects with bad path")

	obj, err := suite.LocalFilesystemBackend.GetObject("deleteme.txt")
	suite.Nil(err, "can get objects with good dir")
	suite.Equal([]byte("some object"), obj.Content, "able to get object with local")
}

func (suite *LocalTestSuite) TestPutObjectWithNonExistentPath() {
	err := suite.BrokenLocalFilesytemBackend.PutObject("testdir/test/test.tgz", []byte("test content"))
	suite.Nil(err)
}

func TestLocalStorageTestSuite(t *testing.T) {
	suite.Run(t, new(LocalTestSuite))
}
