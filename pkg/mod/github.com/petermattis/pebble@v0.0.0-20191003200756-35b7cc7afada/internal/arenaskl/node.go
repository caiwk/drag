/*
 * Copyright 2017 Dgraph Labs, Inc. and Contributors
 * Modifications copyright (C) 2017 Andy Kimball and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package arenaskl

import (
	"math"
	"sync/atomic"

	"github.com/cockroachdb/pebble/internal/base"
)

// MaxNodeSize returns the maximum space needed for a node with the specified
// key and value sizes.
func MaxNodeSize(keySize, valueSize uint32) uint32 {
	return uint32(maxNodeSize) + keySize + valueSize + align4
}

type links struct {
	nextOffset uint32
	prevOffset uint32
}

func (l *links) init(prevOffset, nextOffset uint32) {
	l.nextOffset = nextOffset
	l.prevOffset = prevOffset
}

type node struct {
	// Immutable fields, so no need to lock to access key.
	keyOffset uint32
	keySize   uint32
	// If valueSize is negative, the value is stored separately from the node in
	// arena.extValues.
	valueSize int32
	allocSize uint32

	// Most nodes do not need to use the full height of the tower, since the
	// probability of each successive level decreases exponentially. Because
	// these elements are never accessed, they do not need to be allocated.
	// Therefore, when a node is allocated in the arena, its memory footprint
	// is deliberately truncated to not include unneeded tower elements.
	//
	// All accesses to elements should use CAS operations, with no need to lock.
	tower [maxHeight]links
}

func newNode(
	arena *Arena, height uint32, key base.InternalKey, value []byte,
) (nd *node, err error) {
	if height < 1 || height > maxHeight {
		panic("height cannot be less than one or greater than the max height")
	}
	keySize := key.Size()
	if int64(keySize) > math.MaxUint32 {
		panic("key is too large")
	}
	valueSize := len(value)
	if int64(len(value)) > math.MaxUint32 {
		panic("value is too large")
	}

	nd, err = newRawNode(arena, height, uint32(keySize), uint32(valueSize))
	if err != nil {
		return
	}

	key.Encode(nd.getKeyBytes(arena))
	copy(nd.getValue(arena), value)
	return
}

func newRawNode(arena *Arena, height uint32, keySize, valueSize uint32) (nd *node, err error) {
	// Compute the amount of the tower that will never be used, since the height
	// is less than maxHeight.
	unusedSize := (maxHeight - int(height)) * linksSize
	nodeSize := uint32(maxNodeSize - unusedSize)
	valueIndex := int32(valueSize)

	nodeOffset, allocSize, err := arena.alloc(nodeSize+keySize+valueSize, align4)
	if err != nil {
		return
	}

	nd = (*node)(arena.getPointer(nodeOffset))
	nd.keyOffset = nodeOffset + nodeSize
	nd.keySize = uint32(keySize)
	nd.valueSize = valueIndex
	nd.allocSize = allocSize
	return
}

func (n *node) getKeyBytes(arena *Arena) []byte {
	return arena.getBytes(n.keyOffset, n.keySize)
}

func (n *node) getValue(arena *Arena) []byte {
	return arena.getBytes(n.keyOffset+n.keySize, uint32(n.valueSize))
}

func (n *node) nextOffset(h int) uint32 {
	return atomic.LoadUint32(&n.tower[h].nextOffset)
}

func (n *node) prevOffset(h int) uint32 {
	return atomic.LoadUint32(&n.tower[h].prevOffset)
}

func (n *node) casNextOffset(h int, old, val uint32) bool {
	return atomic.CompareAndSwapUint32(&n.tower[h].nextOffset, old, val)
}

func (n *node) casPrevOffset(h int, old, val uint32) bool {
	return atomic.CompareAndSwapUint32(&n.tower[h].prevOffset, old, val)
}
