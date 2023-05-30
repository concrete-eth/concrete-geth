// Copyright 2023 The concrete-geth Authors
//
// The concrete-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The concrete library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the concrete library. If not, see <http://www.gnu.org/licenses/>.

package lib

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	cc_api "github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/crypto"
)

func NewKey(name string) common.Hash {
	// Use /concrete/crypto instead of /crypto because the latter won't compile
	// in tinygo as it has unsupported dependencies.
	return crypto.Keccak256Hash([]byte(name))
}

type Counter struct {
	cc_api.Reference
}

func NewCounter(ref cc_api.Reference) *Counter {
	return &Counter{ref}
}

func (c *Counter) Get() *big.Int {
	return c.Reference.Get().Big()
}

func (c *Counter) Set(value *big.Int) {
	c.Reference.Set(common.BigToHash(value))
}

func (c *Counter) Add(delta *big.Int) {
	value := c.Get()
	value.Add(value, delta)
	c.Set(value)
}

func (c *Counter) Sub(delta *big.Int) {
	value := c.Get()
	value.Sub(value, delta)
	c.Set(value)
}

func (c *Counter) Inc() {
	c.Add(common.Big1)
}

func (c *Counter) Dec() {
	c.Sub(common.Big1)
}

type nestedMap struct {
	cc_api.Mapping
	depth int
}

func NewNestedMap(mapping cc_api.Mapping, depth int) cc_api.Mapping {
	if depth < 2 {
		panic("depth must be at least 2")
	}
	return &nestedMap{mapping, depth}
}

func (m *nestedMap) GetNested(keys ...common.Hash) common.Hash {
	if len(keys) != m.depth {
		panic("wrong number of keys")
	}
	Array := m.GetNestedMap(keys[0 : m.depth-1]...)
	return Array.Get(keys[m.depth-1])
}

func (m *nestedMap) GetNestedMap(keys ...common.Hash) cc_api.Mapping {
	if len(keys) != m.depth-1 {
		panic("wrong number of keys")
	}
	next := m.Mapping
	for ii := 0; ii < m.depth-1; ii++ {
		next = next.GetMap(keys[ii])
	}
	return next
}

type nestedArray struct {
	cc_api.Array
	depth int
}

func NewNestedArray(array cc_api.Array, depth int) cc_api.Array {
	if depth < 2 {
		panic("depth must be at least 2")
	}
	return &nestedArray{array, depth}
}

func (m *nestedArray) GetNested(indexes ...int) common.Hash {
	if len(indexes) != m.depth {
		panic("wrong number of indexes")
	}
	Array := m.GetNestedArray(indexes[0 : m.depth-1]...)
	return Array.Get(indexes[m.depth-1])
}

func (m *nestedArray) GetNestedArray(indexes ...int) cc_api.Array {
	if len(indexes) != m.depth-1 {
		panic("wrong number of indexes")
	}
	next := m.Array
	for ii := 0; ii < m.depth-1; ii++ {
		next = next.GetArray(indexes[ii])
	}
	return next
}

type BigPreimageStore struct {
	ds       cc_api.Datastore
	radix    int
	leafSize int
}

const (
	LEAF_FLAG = 0x00
	NODE_FLAG = 0x01
)

func NewBigPreimageStore(ds cc_api.Datastore, radix, leafSize int) *BigPreimageStore {
	return &BigPreimageStore{
		ds:       ds,
		radix:    radix,
		leafSize: leafSize,
	}
}

func (s *BigPreimageStore) Add(preimage []byte) common.Hash {
	if len(preimage) == 0 {
		return cc_api.EmptyPreimageHash
	}

	nHashes := (len(preimage) + s.leafSize - 1) / s.leafSize
	hashes := make([][]byte, nHashes)

	// Add leaves
	for ii := 0; ii < nHashes; ii++ {
		l := ii * s.leafSize
		r := (ii + 1) * s.leafSize
		if r > len(preimage) {
			r = len(preimage)
		}
		leaf := s.newLeaf(preimage[l:r])
		hash := s.add(leaf).Bytes()
		hashes[ii] = hash
	}

	// Add internal nodes
	for nHashes != 1 {
		nHashes = (nHashes + s.radix - 1) / s.radix
		for ii := 0; ii < nHashes; ii++ {
			l := ii * s.radix
			r := (ii + 1) * s.radix
			if r > len(hashes) {
				r = len(hashes)
			}
			node := s.newNode(hashes[l:r])
			hash := s.add(node).Bytes()
			hashes[ii] = hash
		}
		hashes = hashes[:nHashes]
	}

	return common.BytesToHash(hashes[0])
}

func (s *BigPreimageStore) newLeaf(body []byte) []byte {
	leaf := make([]byte, 1+len(body))
	leaf[0] = LEAF_FLAG
	copy(leaf[1:], body)
	return leaf
}

func (s *BigPreimageStore) newNode(hashes [][]byte) []byte {
	node := make([]byte, 1+32*len(hashes))
	node[0] = NODE_FLAG
	for ii, hash := range hashes {
		copy(node[1+32*ii:], hash)
	}
	return node
}

func (s *BigPreimageStore) add(preimage []byte) common.Hash {
	hash := crypto.Keccak256Hash(preimage)
	if s.ds.HasPreimage(hash) {
		return hash
	}
	s.ds.AddPreimage(preimage)
	return hash
}

func (s *BigPreimageStore) Get(hash common.Hash) []byte {
	return s.get(hash.Bytes())
}

func (s *BigPreimageStore) get(hash []byte) []byte {
	preimage := s.ds.GetPreimage(common.BytesToHash(hash))
	flag := preimage[0]
	body := preimage[1:]

	if flag == LEAF_FLAG {
		return body
	} else if flag == NODE_FLAG {
		nHashes := len(body) / 32
		preimages := make([][]byte, 0, nHashes)
		for i := 0; i < nHashes; i++ {
			hash := body[i*32 : (i+1)*32]
			preimage := s.get(hash)
			preimages = append(preimages, preimage)
		}
		return bytes.Join(preimages, nil)
	} else {
		panic("invalid flag")
	}
}

func (s *BigPreimageStore) Has(hash common.Hash) bool {
	return s.ds.HasPreimage(hash)
}
