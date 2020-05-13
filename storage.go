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
	"path/filepath"
	"strings"
	"time"
)

type (
	// Object is a generic representation of a storage object
	Object struct {
		Path         string
		Content      []byte
		LastModified time.Time
	}

	// ObjectSliceDiff provides information on what has changed since last calling ListObjects
	ObjectSliceDiff struct {
		Change  bool
		Removed []Object
		Added   []Object
		Updated []Object
	}

	// Backend is a generic interface for storage backends
	Backend interface {
		ListObjects(prefix string) ([]Object, error)
		GetObject(path string) (Object, error)
		PutObject(path string, content []byte) error
		DeleteObject(path string) error
	}
)

// HasExtension determines whether or not an object contains a file extension
func (object Object) HasExtension(extension string) bool {
	return filepath.Ext(object.Path) == fmt.Sprintf(".%s", extension)
}

// GetObjectSliceDiff takes two objects slices and returns an ObjectSliceDiff
func GetObjectSliceDiff(prev []Object, curr []Object, timestampTolerance time.Duration) ObjectSliceDiff {
	var diff ObjectSliceDiff
	pos := make(map[string]Object)
	cos := make(map[string]Object)
	for _, o := range prev {
		pos[o.Path] = o
	}
	for _, o := range curr {
		cos[o.Path] = o
	}
	// for every object in the previous slice, if it exists in the current slice, check if it is *considered as* updated;
	// otherwise, mark it as removed
	for _, p := range prev {
		if c, found := cos[p.Path]; found {
			if c.LastModified.Sub(p.LastModified) > timestampTolerance {
				diff.Updated = append(diff.Updated, c)
			}
		} else {
			diff.Removed = append(diff.Removed, p)
		}
	}
	// for every object in the current slice, if it does not exist in the previous slice, mark it as added
	for _, c := range curr {
		if _, found := pos[c.Path]; !found {
			diff.Added = append(diff.Added, c)
		}
	}
	// if any object is marked as removed or added or updated, set change to true
	diff.Change = len(diff.Removed)+len(diff.Added)+len(diff.Updated) > 0
	return diff
}

func cleanPrefix(prefix string) string {
	return strings.Trim(prefix, "/")
}

func removePrefixFromObjectPath(prefix string, path string) string {
	if prefix == "" {
		return path
	}
	path = strings.Replace(path, fmt.Sprintf("%s/", prefix), "", 1)
	return path
}

func objectPathIsInvalid(path string) bool {
	return strings.Contains(path, "/") || path == ""
}
