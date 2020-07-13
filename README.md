# chartmuseum/storage

[![GitHub Actions status](https://github.com/chartmuseum/storage/workflows/build/badge.svg)](https://github.com/chartmuseum/storage/actions?query=workflow%3Abuild)
[![Go Report Card](https://goreportcard.com/badge/github.com/chartmuseum/storage)](https://goreportcard.com/report/github.com/chartmuseum/storage)
[![GoDoc](https://godoc.org/github.com/chartmuseum/storage?status.svg)](https://godoc.org/github.com/chartmuseum/storage)

Go library providing a common interface for working across multiple storage backends.

Supported storage backends:

- [Alibaba Cloud OSS Storage](https://www.alibabacloud.com/product/oss) ([alibaba.go](./alibaba.go))
- [Amazon S3](https://aws.amazon.com/s3/) ([amazon.go](./amazon.go))
- [Baidu Cloud BOS Storage](https://cloud.baidu.com/product/bos.html) ([baidu.go](./baidu.go))
- [DigitalOcean Spaces](https://www.digitalocean.com/products/spaces/) ([amazon.go](./amazon.go), using custom endpoint and us-east-1)
- [etcd](https://etcd.io/) ([etcd.go](./etcd.go))
- [Google Cloud Storage](https://cloud.google.com/storage/) ([google.go](./google.go))
- Local filesystem ([local.go](./local.go))
- [Microsoft Azure Blob Storage](https://azure.microsoft.com/en-us/services/storage/blobs/) ([microsoft.go](./microsoft.go))
- [Minio](https://min.io/) ([amazon.go](./amazon.go), using custom endpoint and us-east-1)
- [Netease Cloud NOS Storage](https://www.163yun.com/product/nos) ([netease.go](./netease.go))
- [Openstack Object Storage](https://developer.openstack.org/api-ref/object-store/) ([openstack.go](./openstack.go))
- [Oracle Cloud Infrastructure Object Storage](https://cloud.oracle.com/storage) ([oracle.go](./oracle.go))
- [Tencent Cloud Object Storage](https://intl.cloud.tencent.com/product/cos) ([tencent.go](./tencent.go))

*This code was originally part of the [Helm](https://github.com/helm/helm) project: [ChartMuseum](https://github.com/helm/chartmuseum),
but has since been released as a standalone package for others to use in their own projects.*

## Primary Components

### Backend (interface)

`Backend` is a common interface that is implemented by all the supported storage backends and their associated types:

```go
type Backend interface {
    ListObjects(prefix string) ([]Object, error)
    GetObject(path string) (Object, error)
    PutObject(path string, content []byte) error
    DeleteObject(path string) error
}
```

### Object (struct)

`Object` is a struct that represents a single storage object:

```go
type Object struct {
    Path         string
    Content      []byte
    LastModified time.Time
}
```

### ObjectSliceDiff (struct)

`ObjectSliceDiff` is a struct that represents overall changes between two `Object` slices:

```go
type ObjectSliceDiff struct {
    Change  bool
    Removed []Object
    Added   []Object
    Updated []Object
}
```

### GetObjectSliceDiff (function)

`GetObjectSliceDiff` is a function that takes two `Object` slices, compares them, and returns an `ObjectSliceDiff`:

```go
func GetObjectSliceDiff(prev []Object, curr []Object, timestampTolerance time.Duration) ObjectSliceDiff
```

## Usage

### Simple example

The following is a simple program that will upload a file either to an Azure Blob Storage bucket (container) or a Google Cloud Storage bucket based on the command line options provided:

```go
// Usage: go run example.go <cloud> <bucket> <file>

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/chartmuseum/storage"
)

type (
	Uploader struct {
		Backend storage.Backend
	}
)

func NewUploader(cloud string, bucket string) *Uploader {
	var backend storage.Backend
	switch cloud {
	case "azure":
		backend = storage.NewMicrosoftBlobBackend(bucket, "")
	case "google":
		backend = storage.NewGoogleCSBackend(bucket, "")
	default:
		panic("cloud provider " + cloud + " not supported")
	}
	uploader := Uploader{Backend: backend}
	fmt.Printf("uploader created (cloud: %s, bucket: %s)\n", cloud, bucket)
	return &uploader
}

func (uploader *Uploader) Upload(filename string) {
	basename := filepath.Base(filename)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = uploader.Backend.PutObject(basename, content)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s successfully uploaded\n", basename)
}

func main() {
	args := os.Args[1:]
	uploader := NewUploader(args[0], args[1])
	uploader.Upload(args[2])
}
```

Example of using to upload the file `index.html` to an Azure bucket:

```bash
go run example.go azure mycontainer index.html
```

Example of using to upload the file `index.html` to a Google Cloud bucket:

```bash
go run example.go google mybucket index.html
```

### Per backend

Each supported storage backend has its own type that implements the `Backend` interface.
All available types are described in detail on [GoDoc](https://godoc.org/github.com/chartmuseum/storage).

In addition, authentication methods are based on the runtime environment and vary from cloud to cloud.
