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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type LocalTestSuite struct {
	suite.Suite
	LocalFilesystemBackend *LocalFilesystemBackend
	TempRootDirectory      string
}

func (suite *LocalTestSuite) SetupSuite() {
	timestamp := time.Now().Format("20060102150405")
	suite.TempRootDirectory = filepath.Join(".test", timestamp)
	backend := NewLocalFilesystemBackend(suite.TempRootDirectory)
	suite.LocalFilesystemBackend = backend
	_, err := os.Stat(filepath.Join(suite.LocalFilesystemBackend.TempDirectory))
	suite.Nil(err, "should have created a .tmp/ directory for partial uploads")
}

func (suite *LocalTestSuite) TestListObjects() {
	_, err := suite.LocalFilesystemBackend.ListObjects("")
	suite.Nil(err, "list objects does not return error if dir does not exist")
}

func (suite *LocalTestSuite) TestGetObject() {
	_, err := suite.LocalFilesystemBackend.GetObject("this-file-cannot-possibly-exist.tgz")
	suite.NotNil(err, "cannot get objects with bad path")
}

func (suite *LocalTestSuite) TestPutObjectWithNonExistentPath() {
	err := suite.LocalFilesystemBackend.PutObject("testdir/test/test.tgz", []byte("test content"))
	suite.Nil(err)
}

func TestLocalStorageTestSuite(t *testing.T) {
	ts := new(LocalTestSuite)
	defer os.RemoveAll(ts.TempRootDirectory)
	suite.Run(t, ts)
}
