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
	"io/ioutil"
	"os"

	pathutil "path"
	"path/filepath"
)

// LocalFilesystemBackend is a storage backend for local filesystem storage
type LocalFilesystemBackend struct {
	RootDirectory string
	TempDirectory string
}

// NewLocalFilesystemBackend creates a new instance of LocalFilesystemBackend
func NewLocalFilesystemBackend(rootDirectory string) *LocalFilesystemBackend {
	absPath, err := filepath.Abs(rootDirectory)
	if err != nil {
		panic(err)
	}
	// Create a temporary folder for partially-completed writes (if not present)
	tempPath := filepath.Join(absPath, ".tmp")
	_, err = os.Stat(tempPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(tempPath, 0774)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	b := &LocalFilesystemBackend{RootDirectory: absPath, TempDirectory: tempPath}
	return b
}

// ListObjects lists all objects in root directory (depth 1)
func (b LocalFilesystemBackend) ListObjects(prefix string) ([]Object, error) {
	var objects []Object
	files, err := ioutil.ReadDir(pathutil.Join(b.RootDirectory, prefix))
	if err != nil {
		if os.IsNotExist(err) { // OK if the directory doesnt exist yet
			err = nil
		}
		return objects, err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		object := Object{Path: f.Name(), Content: []byte{}, LastModified: f.ModTime()}
		objects = append(objects, object)
	}
	return objects, nil
}

// GetObject retrieves an object from root directory
func (b LocalFilesystemBackend) GetObject(path string) (Object, error) {
	var object Object
	object.Path = path
	fullpath := pathutil.Join(b.RootDirectory, path)
	content, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return object, err
	}
	object.Content = content
	info, err := os.Stat(fullpath)
	if err != nil {
		return object, err
	}
	object.LastModified = info.ModTime()
	return object, err
}

// PutObject puts an object in root directory
func (b LocalFilesystemBackend) PutObject(path string, content []byte) error {
	fullpath := pathutil.Join(b.RootDirectory, path)
	folderPath := pathutil.Dir(fullpath)
	_, err := os.Stat(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(folderPath, 0774)
			if err != nil {
				return err
			}
			// os.MkdirAll set the dir permissions before the umask
			// we need to use os.Chmod to ensure the permissions of the created directory are 774
			// because the default umask will prevent that and cause the permissions to be 755
			err = os.Chmod(folderPath, 0774)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Write to a temporary file first
	tempFile, err := os.CreateTemp(b.TempDirectory, "partial-upload-")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	_, err = tempFile.Write(content)
	if err != nil {
		return err
	}
	tempFile.Close()

	// Default permissions on a temp file are 600, so let's change that
	err = os.Chmod(tempFile.Name(), 0774)
	if err != nil {
		return err
	}

	// Now that the file is written safely, atomically move it into place
	err = os.Rename(tempFile.Name(), fullpath)
	return err
}

// DeleteObject removes an object from root directory
func (b LocalFilesystemBackend) DeleteObject(path string) error {
	fullpath := pathutil.Join(b.RootDirectory, path)
	err := os.Remove(fullpath)
	return err
}
