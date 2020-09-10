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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	pathutil "path"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	osObjects "github.com/gophercloud/gophercloud/openstack/objectstorage/v1/objects"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/swauth"
	"github.com/gophercloud/gophercloud/pagination"
)

// ReauthRoundTripper satisfies the http.RoundTripper interface and is used to
// limit the number of consecutive re-auth attempts (infinite by default)
type ReauthRoundTripper struct {
	rt                http.RoundTripper
	numReauthAttempts int
}

// RoundTrip performs a round-trip HTTP request and logs relevant information about it.
func (rrt *ReauthRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := rrt.rt.RoundTrip(request)
	if response == nil {
		return nil, err
	}

	if response.StatusCode == http.StatusUnauthorized {
		if rrt.numReauthAttempts == 3 {
			return response, errors.New("tried to re-authenticate 3 times with no success")
		}
		rrt.numReauthAttempts++
	} else {
		rrt.numReauthAttempts = 0
	}

	return response, nil
}

// OpenstackOSBackend is a storage backend for Openstack Object Storage
type OpenstackOSBackend struct {
	Container string
	Prefix    string
	Region    string
	CACert    string
	Client    *gophercloud.ServiceClient
}

// NewOpenstackOSBackend creates a new instance of OpenstackOSBackend
func NewOpenstackOSBackend(container string, prefix string, region string, caCert string) *OpenstackOSBackend {
	authOptions, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		panic(fmt.Sprintf("Openstack (environment): %s", err))
	}
	authOptions.AllowReauth = true

	if authScope := getAuthScope(); authScope != nil {
		authOptions.Scope = authScope
	}

	if userDomainName := os.Getenv("OS_USER_DOMAIN_NAME"); userDomainName != "" {
		authOptions.DomainName = userDomainName
	}
	if userDomainID := os.Getenv("OS_USER_DOMAIN_ID"); userDomainID != "" {
		authOptions.DomainID = userDomainID
	}

	// Create a custom HTTP client to handle reauth retry and custom CACERT if needed
	roundTripper := ReauthRoundTripper{}
	if caCert != "" {
		caCert, err := ioutil.ReadFile(caCert)
		if err != nil {
			panic(fmt.Sprintf("Openstack (ca certificates): %s", err))
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			panic(fmt.Sprintf("Openstack (ca certificates): unable to read certificate bundle"))
		}

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
		roundTripper.rt = transport
	} else {
		roundTripper.rt = http.DefaultTransport
	}

	provider, err := openstack.NewClient(authOptions.IdentityEndpoint)
	if err != nil {
		panic(fmt.Sprintf("Openstack (client): %s", err))
	}

	provider.HTTPClient = http.Client{
		Transport: &roundTripper,
	}

	err = openstack.Authenticate(provider, authOptions)
	if err != nil {
		panic(fmt.Sprintf("Openstack (authenticate): %s", err))
	}

	client, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		panic(fmt.Sprintf("Openstack (object storage): %s", err))
	}

	b := &OpenstackOSBackend{
		Container: container,
		Prefix:    prefix,
		Region:    region,
		Client:    client,
	}

	return b
}

// NewOpenstackOSBackendV1Auth creates a new instance of OpenstackOSBackend using Swift V1 Auth
func NewOpenstackOSBackendV1Auth(container string, prefix string, caCert string) *OpenstackOSBackend {
	for _, e := range []string{"ST_USER", "ST_KEY", "ST_AUTH"} {
		if os.Getenv(e) == "" {
			panic(fmt.Sprintf("Openstack (object storage): missing environment variable %s", e))
		}
	}

	authOpts := swauth.AuthOpts{
		User: os.Getenv("ST_USER"),
		Key:  os.Getenv("ST_KEY"),
	}
	identityEndpoint := os.Getenv("ST_AUTH")

	// Create a custom HTTP client to handle custom CACERT if needed
	httpTransport := http.DefaultTransport
	if caCert != "" {
		caCert, err := ioutil.ReadFile(caCert)
		if err != nil {
			panic(fmt.Sprintf("Openstack (ca certificates): %s", err))
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			panic(fmt.Sprintf("Openstack (ca certificates): unable to read certificate bundle"))
		}

		httpTransport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
	}

	provider, err := openstack.NewClient(identityEndpoint)
	if err != nil {
		panic(fmt.Sprintf("Openstack (client): %s", err))
	}

	provider.HTTPClient = http.Client{
		Transport: httpTransport,
	}

	// gophercloud does not support reauth for Swift V1 clients, so we handle this here.
	// This is more or less a carbon copy of what gophercloud/openstack/client.go does vor v2.
	//
	// here we're creating a throw-away client (tac). it's a copy of the user's provider client, but
	// with the token and reauth func zeroed out. This should retry authentication only once.
	tac := *provider
	tac.SetThrowaway(true)
	tac.ReauthFunc = nil
	tac.SetTokenAndAuthResult(nil)
	tao := authOpts
	provider.ReauthFunc = func() error {
		auth, err := swauth.Auth(&tac, tao).Extract()
		if err != nil {
			return err
		}
		// safely copy the token from tac to this ProviderClient
		provider.SetToken(auth.Token)
		return nil
	}

	client, err := swauth.NewObjectStorageV1(provider, authOpts)
	if err != nil {
		panic(fmt.Sprintf("Openstack (object storage): %s", err))
	}

	b := &OpenstackOSBackend{
		Container: container,
		Prefix:    prefix,
		Client:    client,
	}

	return b
}

// ListObjects lists all objects in an Openstack container, at prefix
func (b OpenstackOSBackend) ListObjects(prefix string) ([]Object, error) {
	var objects []Object

	prefix = pathutil.Join(b.Prefix, prefix)
	opts := &osObjects.ListOpts{
		Full:   true,
		Prefix: prefix,
	}

	pager := osObjects.List(b.Client, b.Container, opts)
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		objectList, err := osObjects.ExtractInfo(page)
		if err != nil {
			return false, err
		}

		for _, openStackObject := range objectList {
			path := removePrefixFromObjectPath(prefix, openStackObject.Name)
			if objectPathIsInvalid(path) {
				continue
			}

			// This is a patch so that LastModified match between the List and GetObject function
			// Openstack seems to send a rounded up time when getting the LastModified date from an object show versus an object list
			var lastModified time.Time
			if openStackObject.LastModified.Nanosecond()/int(time.Microsecond) == 0 {
				lastModified = openStackObject.LastModified
			} else {
				lastModified = openStackObject.LastModified.Truncate(time.Second).Add(time.Second)
			}

			object := Object{
				Path:         path,
				Content:      []byte{},
				LastModified: lastModified,
			}
			objects = append(objects, object)
		}
		return true, nil
	})

	return objects, err
}

// GetObject retrieves an object from an Openstack container, at prefix
func (b OpenstackOSBackend) GetObject(path string) (Object, error) {
	var object Object
	object.Path = path

	result := osObjects.Download(b.Client, b.Container, pathutil.Join(b.Prefix, path), nil)
	headers, err := result.Extract()
	if err != nil {
		return object, err
	}
	object.LastModified = headers.LastModified

	content, err := result.ExtractContent()
	if err != nil {
		return object, err
	}
	object.Content = content
	return object, nil
}

// PutObject uploads an object to Openstack container, at prefix
func (b OpenstackOSBackend) PutObject(path string, content []byte) error {
	reader := bytes.NewReader(content)
	createOpts := osObjects.CreateOpts{
		Content: reader,
	}
	_, err := osObjects.Create(b.Client, b.Container, pathutil.Join(b.Prefix, path), createOpts).Extract()
	return err
}

// DeleteObject removes an object from an Openstack container, at prefix
func (b OpenstackOSBackend) DeleteObject(path string) error {
	_, err := osObjects.Delete(b.Client, b.Container, pathutil.Join(b.Prefix, path), nil).Extract()
	return err
}

func getAuthScope() *gophercloud.AuthScope {
	scope := &gophercloud.AuthScope{}

	// Scope to project by ID.
	if projectID := os.Getenv("OS_PROJECT_ID"); projectID != "" {
		scope.ProjectID = projectID
		return scope
	}

	// Scope to project by name. Requires the project domain name or ID as well.
	projectName := os.Getenv("OS_PROJECT_NAME")
	if projectName == "" {
		return nil
	}
	scope.ProjectName = projectName

	projectDomainName := os.Getenv("OS_PROJECT_DOMAIN_NAME")
	projectDomainID := os.Getenv("OS_PROJECT_DOMAIN_ID")

	if projectDomainName == "" && projectDomainID == "" {
		return nil
	}

	if projectDomainName != "" {
		scope.DomainName = projectDomainName
	}
	scope.DomainID = projectDomainID
	return scope
}
