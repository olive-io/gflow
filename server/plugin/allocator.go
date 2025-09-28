/*
Copyright 2025 The gflow Authors

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

package plugin

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
)

// This is an application-wide global ID allocator.  Unfortunately we need
// to have unique IDs globally to permit certain things to work correctly.
type idAllocator struct {
	used map[uint64]struct{}
	next uint64
	lock sync.Mutex
}

func newIDAllocator() *idAllocator {
	b := make([]byte, 8)
	// The following could in theory fail, but in that case
	// we will wind up with IDs starting at zero.  It should
	// not happen unless the platform can't get good entropy.
	_, _ = rand.Read(b)
	used := make(map[uint64]struct{})
	next := binary.BigEndian.Uint64(b)
	alloc := &idAllocator{
		used: used,
		next: next,
	}
	return alloc
}

func (alloc *idAllocator) Get() uint64 {
	alloc.lock.Lock()
	defer alloc.lock.Unlock()
	for {
		id := alloc.next & 0x7fffffff
		alloc.next++
		if id == 0 {
			continue
		}
		if _, ok := alloc.used[id]; ok {
			continue
		}
		alloc.used[id] = struct{}{}
		return id
	}
}

func (alloc *idAllocator) Free(id uint64) {
	alloc.lock.Lock()
	if _, ok := alloc.used[id]; ok {
		delete(alloc.used, id)
	}
	alloc.lock.Unlock()
}
