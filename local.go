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
	"io/ioutil"
	"os"
	"sync"

	pathutil "path"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// LocalFilesystemBackend is a storage backend for local filesystem storage
type LocalFilesystemBackend struct {
	RootDirectory string
}

type NewLocalFilesystemBackendOption struct {
	notifier map[EventType]func()
}

var watcherOnce sync.Once
var globalWatcher *fsnotify.Watcher

func WithEventNotifier(e EventType, fn func()) func(*NewLocalFilesystemBackendOption) {
	return func(option *NewLocalFilesystemBackendOption) {
		if option.notifier == nil {
			option.notifier = make(map[EventType]func())
		}
		option.notifier[e] = fn
	}
}

// NewLocalFilesystemBackend creates a new instance of LocalFilesystemBackend
func NewLocalFilesystemBackend(rootDirectory string, opts ...func(*NewLocalFilesystemBackendOption)) *LocalFilesystemBackend {
	var option NewLocalFilesystemBackendOption
	for _, opt := range opts {
		opt(&option)
	}
	absPath, err := filepath.Abs(rootDirectory)
	if err != nil {
		panic(err)
	}
	b := &LocalFilesystemBackend{
		RootDirectory: absPath,
	}

	watcherOnce.Do(func() {
		globalWatcher, err = fsnotify.NewWatcher()
		if err != nil {
			panic(err)
		}
		// since it is a longTerm watcher , we do not need to close it
		go func() {
			for {
				select {
				case event, ok := <-globalWatcher.Events:
					if !ok {
						continue
					}

					switch event.Op {
					case fsnotify.Write, fsnotify.Create:
						if fn, ok := option.notifier[EventPutObject]; ok {
							fn()
						}
					case fsnotify.Remove:
						if fn, ok := option.notifier[EventDeleteObject]; ok {
							fn()
						}
					}
				case _, ok := <-globalWatcher.Errors:
					if !ok {
						continue
					}
				}
			}
		}()

		if err := globalWatcher.Add(b.RootDirectory); err != nil {
			panic(err)
		}
	})

	return b
}

func (b LocalFilesystemBackend) AddWatcherPath(path string) error {
	// TODO: to ensure that we should lock here ?
	return globalWatcher.Add(path)
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
			// NOTE: works for dynamic depth tenant
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
			// also adds the fsnotify watcher path
			if err := b.AddWatcherPath(folderPath); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if err = ioutil.WriteFile(fullpath, content, 0644); err != nil {
		return err
	}
	return nil
}

// DeleteObject removes an object from root directory
func (b LocalFilesystemBackend) DeleteObject(path string) error {
	fullpath := pathutil.Join(b.RootDirectory, path)
	if err := os.Remove(fullpath); err != nil {
		return fmt.Errorf("failed to delete object %s: %w", path, err)
	}
	return nil
}
