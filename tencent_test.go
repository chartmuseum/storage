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

type TencentTestSuite struct {
	suite.Suite
	BrokenTencentCloudCOSBackend   *TencentCloudCOSBackend
	NoPrefixTencentCloudCOSBackend *TencentCloudCOSBackend
}

const testCounts = 1

func (suite *TencentTestSuite) SetupSuite() {
	backend := NewTencentCloudCOSBackend("fake-bucket-cant-exist-fbce123", "", "")
	suite.BrokenTencentCloudCOSBackend = backend

	cosBucket := os.Getenv("TEST_STORAGE_TENCENT_BUCKET")
	cosEndpoint := os.Getenv("TEST_STORAGE_TENCENT_ENDPOINT")
	backend = NewTencentCloudCOSBackend(cosBucket, "", cosEndpoint)
	suite.NoPrefixTencentCloudCOSBackend = backend

	data := []byte("some object")
	path := "deleteme.txt"

	for i := 0; i < testCounts; i++ {
		newPath := strconv.Itoa(i) + path
		err := suite.NoPrefixTencentCloudCOSBackend.PutObject(newPath, data)
		suite.Nil(err, "no error putting deleteme.txt using Tencent Cloud COS backend")
	}
}

func (suite *TencentTestSuite) TearDownSuite() {
	path := "deleteme.txt"
	for i := 0; i < testCounts; i++ {
		newPath := strconv.Itoa(i) + path

		err := suite.NoPrefixTencentCloudCOSBackend.DeleteObject(newPath)
		suite.Nil(err, "no error deleting deleteme.txt using Tencent Cloud COS backend")
	}
}

func (suite *TencentTestSuite) TestListObjects() {
	_, err := suite.BrokenTencentCloudCOSBackend.ListObjects("")
	suite.NotNil(err, "cannot list objects with bad bucket")

	objs, err := suite.NoPrefixTencentCloudCOSBackend.ListObjects("")
	suite.Nil(err, "can list objects with good bucket, no prefix")
	suite.Equal(len(objs), testCounts, "able to list objects")

}

func (suite *TencentTestSuite) TestGetObject() {
	_, err := suite.BrokenTencentCloudCOSBackend.GetObject("this-file-cannot-possibly-exist.tgz")
	suite.NotNil(err, "cannot get objects with bad bucket")
}

func (suite *TencentTestSuite) TestPutObject() {
	err := suite.BrokenTencentCloudCOSBackend.PutObject("this-file-will-not-upload.txt", []byte{})
	suite.NotNil(err, "cannot put objects with bad bucket")
}

func TestTencentStorageTestSuite(t *testing.T) {
	if os.Getenv("TEST_CLOUD_STORAGE") == "1" &&
		os.Getenv("TEST_STORAGE_TENCENT_BUCKET") != "" &&
		os.Getenv("TEST_STORAGE_TENCENT_ENDPOINT") != "" {
		suite.Run(t, new(TencentTestSuite))
	}
}