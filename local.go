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
}

type LocalItem struct {
	obj *Object
	dir bool
	err error
}

// NewLocalFilesystemBackend creates a new instance of LocalFilesystemBackend
func NewLocalFilesystemBackend(rootDirectory string) *LocalFilesystemBackend {
	absPath, err := filepath.Abs(rootDirectory)
	if err != nil {
		panic(err)
	}
	b := &LocalFilesystemBackend{RootDirectory: absPath}
	return b
}

// ListObjects lists all objects in root directory (depth 1)
func (b LocalFilesystemBackend) ListObjects(prefix string) ([]Object, error) {
	objects := []Object{}
	for item := range(b.FileIter(prefix)) {
		if item.obj == nil || item.err != nil {
			return objects, item.err
		}
		if !item.dir {
			objects = append(objects, *item.obj)
		}
	}
	return objects, nil
}

// ListFolders lists all folders in root directory (depth 1)
func (b LocalFilesystemBackend) ListFolders(prefix string) ([]Folder, error) {
	folders := []Folder{}
	for item := range(b.FileIter(prefix)) {
		if item.obj == nil || item.err != nil {
			return folders, item.err
		}
		if item.dir {
			folders = append(folders, Folder{Path: item.obj.Path, LastModified: item.obj.LastModified})
		}
	}
	return folders, nil
}

func (b LocalFilesystemBackend) FileIter(prefix string) <-chan LocalItem {
	ch := make(chan(LocalItem))
	go func() {
		files, err := ioutil.ReadDir(pathutil.Join(b.RootDirectory, prefix))
		if err != nil {
			if os.IsNotExist(err) {  // OK if the directory doesnt exist yet
				err = nil
			}
			ch <- LocalItem{obj: nil, dir: false, err: err}
		}
		for _, f := range files {
			ch <- LocalItem{
				obj: &Object{
					Path: f.Name(),
					Content: []byte{},
					LastModified: f.ModTime(),
				},
				dir: f.IsDir(),
				err: nil,
			}
		}
		close(ch)
	} ()
	return ch
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
			err := os.MkdirAll(folderPath, 0777)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	err = ioutil.WriteFile(fullpath, content, 0644)
	return err
}

// DeleteObject removes an object from root directory
func (b LocalFilesystemBackend) DeleteObject(path string) error {
	fullpath := pathutil.Join(b.RootDirectory, path)
	err := os.Remove(fullpath)
	return err
}
