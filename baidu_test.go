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
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
)

type BaiduTestSuite struct {
	suite.Suite
	BrokenBaiduBOSBackend   *BaiduBOSBackend
	NoPrefixBaiduBOSBackend *BaiduBOSBackend
}

const bosTestCount = 100

func (suite *BaiduTestSuite) SetupSuite() {
	backend := NewBaiDuBOSBackend("fake-container-cant-exist-fbce123", "", "")
	suite.BrokenBaiduBOSBackend = backend

	bosBucket := os.Getenv("TEST_STORAGE_BAIDU_BUCKET")
	bosEndpoint := os.Getenv("TEST_STORAGE_BAIDU_ENDPOINT")
	backend = NewBaiDuBOSBackend(bosBucket, "", bosEndpoint)
	suite.NoPrefixBaiduBOSBackend = backend

	data := []byte("some object")
	path := "deleteme.txt"

	for i := 0; i < bosTestCount; i++ {
		newPath := strconv.Itoa(i) + path
		err := suite.NoPrefixBaiduBOSBackend.PutObject(newPath, data)
		suite.Nil(err, "no error putting deleteme.txt using Baidu Cloud BOS backend")
	}
}

func (suite *BaiduTestSuite) TearDownSuite() {
	path := "deleteme.txt"
	for i := 0; i < bosTestCount; i++ {
		newPath := strconv.Itoa(i) + path

		err := suite.NoPrefixBaiduBOSBackend.DeleteObject(newPath)
		suite.Nil(err, "no error deleting deleteme.txt using BaiduBOS backend")
	}
}

func (suite *BaiduTestSuite) TestListObjects() {
	_, err := suite.BrokenBaiduBOSBackend.ListObjects("")
	suite.NotNil(err, "cannot list objects with bad bucket")

	objs, err := suite.NoPrefixBaiduBOSBackend.ListObjects("")
	suite.Nil(err, "can list objects with good bucket, no prefix")
	suite.Equal(len(objs), bosTestCount, "able to list objects")
}

func (suite *BaiduTestSuite) TestGetObject() {
	_, err := suite.BrokenBaiduBOSBackend.GetObject("this-file-cannot-possibly-exist.tgz")
	suite.NotNil(err, "cannot get objects with bad bucket")
}

func (suite *BaiduTestSuite) TestPutObject() {
	err := suite.BrokenBaiduBOSBackend.PutObject("this-file-will-not-upload.txt", []byte{})
	suite.NotNil(err, "cannot put objects with bad bucket")
}

func TestBaiduStorageTestSuite(t *testing.T) {
	if os.Getenv("TEST_CLOUD_STORAGE") == "1" &&
		os.Getenv("TEST_STORAGE_BAIDU_BUCKET") != "" &&
		os.Getenv("TEST_STORAGE_BAIDU_ENDPOINT") != "" {
		suite.Run(t, new(BaiduTestSuite))
	}
}
