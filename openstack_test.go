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
	"testing"

	osContainers "github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"

	"github.com/stretchr/testify/suite"
)

type OpenstackTestSuite struct {
	suite.Suite
	BrokenOpenstackOSBackend   []*OpenstackOSBackend
	NoPrefixOpenstackOSBackend []*OpenstackOSBackend
}

func (suite *OpenstackTestSuite) SetupSuite() {
	osRegion := os.Getenv("TEST_STORAGE_OPENSTACK_REGION")
	osContainer := os.Getenv("TEST_STORAGE_OPENSTACK_CONTAINER")

	if os.Getenv("OS_AUTH_URL") != "" && osRegion != "" {
		backend := NewOpenstackOSBackend("fake-container-cant-exist-fbce123", "", osRegion, "")
		suite.BrokenOpenstackOSBackend = append(suite.BrokenOpenstackOSBackend, backend)

		backend = NewOpenstackOSBackend(osContainer, "", osRegion, "")
		suite.NoPrefixOpenstackOSBackend = append(suite.NoPrefixOpenstackOSBackend, backend)
	} else {
		suite.T().Log("Skipping OpenStack Swift tests due to missing ENV vars.")
	}

	if os.Getenv("ST_AUTH") != "" {
		backend := NewOpenstackOSBackendV1Auth("fake-container-cant-exist-fbce123", "", "")
		suite.BrokenOpenstackOSBackend = append(suite.BrokenOpenstackOSBackend, backend)

		backend = NewOpenstackOSBackendV1Auth(osContainer, "", "")
		suite.NoPrefixOpenstackOSBackend = append(suite.NoPrefixOpenstackOSBackend, backend)
	} else {
		suite.T().Log("Skipping Swift TempAuth (V1 Auth) tests due to missing ENV vars.")
	}

	data := []byte("some object")
	path := "deleteme.txt"
	for _, backend := range suite.NoPrefixOpenstackOSBackend {
		_, err := osContainers.Create(backend.Client, osContainer, nil).Extract()
		suite.Nil(err, "error creating container %s: %v", osContainer, err)
		err = backend.PutObject(path, data)
		suite.Nil(err, "error putting deleteme.txt using openstack backend")
	}
}

func (suite *OpenstackTestSuite) TearDownSuite() {
	for _, backend := range suite.NoPrefixOpenstackOSBackend {
		err := backend.DeleteObject("deleteme.txt")
		suite.Nil(err, "error deleting deleteme.txt using Openstack backend")
	}
}

func (suite *OpenstackTestSuite) TestListObjects() {
	for _, backend := range suite.BrokenOpenstackOSBackend {
		_, err := backend.ListObjects("")
		suite.NotNil(err, "cannot list objects with bad container")
	}

	for _, backend := range suite.NoPrefixOpenstackOSBackend {
		_, err := backend.ListObjects("")
		suite.Nil(err, "can list objects with good container, no prefix")
	}
}

func (suite *OpenstackTestSuite) TestGetObject() {
	for _, backend := range suite.BrokenOpenstackOSBackend {
		_, err := backend.GetObject("this-file-cannot-possibly-exist.tgz")
		suite.NotNil(err, "cannot get objects with bad container")
	}
}

func (suite *OpenstackTestSuite) TestPutObject() {
	for _, backend := range suite.BrokenOpenstackOSBackend {
		err := backend.PutObject("this-file-will-not-upload.txt", []byte{})
		suite.NotNil(err, "cannot put objects with bad container")
	}
}

func TestOpenstackOSStorageTestSuite(t *testing.T) {
	if os.Getenv("TEST_CLOUD_STORAGE") == "1" &&
		os.Getenv("TEST_STORAGE_OPENSTACK_CONTAINER") != "" {
		suite.Run(t, new(OpenstackTestSuite))
	}
}
