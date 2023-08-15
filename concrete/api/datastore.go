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

package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/crypto"
)

// var (
// 	PrecompileRegistryAddress  = common.HexToAddress("0xcc00000000000000000000000000000000000000")
// 	PreimageRegistryAddress    = common.HexToAddress("0xcc00000000000000000000000000000000000001")
// 	BigPreimageRegistryAddress = common.HexToAddress("0xcc00000000000000000000000000000000000002")
// )

// var (
// crypto.Keccak256Hash(nil)
// EmptyPreimageHash = common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
// )

/*
- A way to revert without panic [?] Just return an error [?]
- A way to store byte arrays -- easy
- A way to store structs
*/

type ValueStore interface {
	Set(value common.Hash)
	Get() common.Hash
}

type KeyValueStore interface {
	Set(key common.Hash, value common.Hash)
	Get(key common.Hash) common.Hash
}

type Slot interface {
	Slot() common.Hash
}

type Array interface {
	Length() int
	Get(index int) common.Hash
	Set(index int, value common.Hash)
}

type DatastoreModelCreator interface {
	Reference(slot common.Hash) Reference
	Mapping(slot common.Hash) Mapping
	StaticArray(slot common.Hash, itemSize int, length ...int) StaticArray
	DynamicArray(slot common.Hash, itemSize int) DynamicArray
}

type Datastore interface {
	KeyValueStore
	DatastoreModelCreator
}

type datastore struct {
	KeyValueStore
}

func NewDatastore(kv KeyValueStore, slot []byte) Datastore {
	return &datastore{KeyValueStore: kv}
}

func (ds *datastore) Reference(slot common.Hash) Reference {
	return NewReference(ds, slot)
}

func (ds *datastore) Mapping(slot common.Hash) Mapping {
	return NewMapping(ds, slot)
}

func (ds *datastore) StaticArray(slot common.Hash, itemSize int, length ...int) StaticArray {
	return NewStaticArray(ds, slot, itemSize, length)
}

func (ds *datastore) DynamicArray(slot common.Hash, itemSize int) DynamicArray {
	return NewDynamicArray(ds, slot, itemSize)
}

var _ Datastore = (*datastore)(nil)

type Reference interface {
	Slot
	ValueStore
	// GetBool() bool
	// SetBool(value bool)
	// GetAddress() common.Address
	// SetAddress(value common.Address)
	// GetBig() *big.Int
	// SetBig(value *big.Int)
	// SetInt64(value int64)
	// GetInt64() int64
	GetBytes() []byte
	SetBytes(value []byte)
}

type reference struct {
	ds   Datastore
	slot common.Hash
}

func NewReference(ds Datastore, slot common.Hash) Reference {
	return &reference{ds: ds, slot: slot}
}

func (r *reference) Slot() common.Hash {
	return r.slot
}

func (r *reference) Get() common.Hash {
	return r.ds.Get(r.slot)
}

func (r *reference) Set(value common.Hash) {
	r.ds.Set(r.slot, value)
}

// TODO: panic!

func (r *reference) GetBytes() []byte {
	// TODO: check bounds
	slotWord := r.ds.Get(r.slot)
	lsb := slotWord[len(slotWord)-1]
	isShort := lsb&1 == 0
	if isShort {
		length := int(lsb) / 2
		// TODO: check bounds
		return slotWord[:length]
	}
	// TODO: check bounds
	// length will always be > 31
	length := slotWord.Big().Int64()
	ptr := crypto.Keccak256Hash(r.slot.Bytes()).Big()

	data := make([]byte, length)
	for ii := 0; ii < len(data); ii += 32 {
		copy(data[ii:], r.ds.Get(common.BigToHash(ptr)).Bytes())
		ptr = ptr.Add(ptr, common.Big1)
	}

	return data
}

func (r *reference) SetBytes(value []byte) {
	isShort := len(value) <= 31
	if isShort {
		var data [32]byte
		copy(data[:], value)
		data[31] = byte(len(value) * 2)
		r.ds.Set(r.slot, common.BytesToHash(data[:]))
		return
	}

	lengthBN := big.NewInt(int64(len(value)))
	r.ds.Set(r.slot, common.BigToHash(lengthBN))

	// Then store the actual data starting at the keccak256 hash of the slot
	ptr := crypto.Keccak256Hash(r.slot.Bytes()).Big()

	for ii := 0; ii < len(value); ii += 32 {
		var data [32]byte
		copy(data[:], value[ii:])
		r.ds.Set(common.BigToHash(ptr), common.BytesToHash(data[:]))
		ptr = ptr.Add(ptr, common.Big1)
	}
}

var _ Reference = (*reference)(nil)

type Mapping interface {
	Slot
	Datastore
}

type mapping struct {
	Datastore
	slot common.Hash
}

func NewMapping(ds Datastore, slot common.Hash) Mapping {
	return &mapping{Datastore: ds, slot: slot}
}

func (m *mapping) valueSlot(key common.Hash) common.Hash {
	return crypto.Keccak256Hash(key.Bytes(), m.slot.Bytes())
}

func (m *mapping) Slot() common.Hash {
	return m.slot
}

func (m *mapping) Reference(key common.Hash) Reference {
	slot := m.valueSlot(key)
	return NewReference(m, slot)
}

func (m *mapping) Mapping(key common.Hash) Mapping {
	slot := m.valueSlot(key)
	return NewMapping(m, slot)
}

func (m *mapping) StaticArray(key common.Hash, itemSize int, length ...int) StaticArray {
	slot := m.valueSlot(key)
	return NewStaticArray(m, slot, itemSize, length)
}

func (m *mapping) DynamicArray(key common.Hash, itemSize int) DynamicArray {
	slot := m.valueSlot(key)
	return NewDynamicArray(m, slot, itemSize)
}

var _ Mapping = (*mapping)(nil)

type StaticArray interface {
	Slot
	Array
}

type staticArray struct {
	ds       Datastore
	slot     common.Hash
	length   []int
	itemSize int
}

func NewStaticArray(ds Datastore, slot common.Hash, itemSize int, length []int) StaticArray {
	if len(length) == 0 {
		length = []int{0}
	}
	return &staticArray{ds: ds, slot: slot, length: length, itemSize: itemSize}
}

func (a *staticArray) Length() int {
	return a.length[0]
}

func (a *staticArray) Get(index int) common.Hash {
	if index >= a.Length() {
		return common.Hash{}
	}
	itemsPerSlot := 32 / a.itemSize
	slot := new(big.Int).Add(a.slot.Big(), big.NewInt(int64(index/itemsPerSlot)))
	offset := (index % itemsPerSlot) * a.itemSize
	word := a.ds.Get(common.BigToHash(slot))
	return a.ds.Get(a.key(index))
}

func (a *staticArray) Set(index int, value common.Hash) {
	if index >= a.Length() {
		panic("index out of bounds")
	}
	a.ds.Set(a.key(index), value)
}

type DynamicArray interface {
	DatastoreModel
	DatastoreModelCreator
	Array
	Push(value common.Hash)
	Pop() common.Hash
	Swap(i, j int)
}

type dynamicArray struct {
	ds     Datastore
	id     common.Hash
	idHash common.Hash
}

func NewDynamicArray(ds Datastore, slot common.Hash, itemSize int) DynamicArray {
	return &dynamicArray{ds: ds, id: id}
}

func (a *dynamicArray) getIdHash() common.Hash {
	if a.idHash == (common.Hash{}) {
		a.idHash = crypto.Keccak256Hash(a.id.Bytes())
	}
	return a.idHash
}

func (a *dynamicArray) key(index int) common.Hash {
	a.getIdHash()
	slot := new(big.Int).Add(a.idHash.Big(), big.NewInt(int64(index)))
	return common.BigToHash(slot)
}

func (a *dynamicArray) setLength(length int) {
	a.ds.Set(a.id, common.BigToHash(big.NewInt(int64(length))))
}

func (a *dynamicArray) getLength() int {
	return int(a.ds.Get(a.id).Big().Int64())
}

func (a *dynamicArray) Id() common.Hash {
	return a.id
}

func (a *dynamicArray) Length() int {
	return a.getLength()
}

func (a *dynamicArray) Get(index int) common.Hash {
	if index >= a.Length() {
		return common.Hash{}
	}
	return a.ds.Get(a.key(index))
}

func (a *dynamicArray) Set(index int, value common.Hash) {
	if index >= a.Length() {
		panic("index out of bounds")
	}
	a.ds.Set(a.key(index), value)
}

func (a *dynamicArray) Push(value common.Hash) {
	length := a.Length()
	a.setLength(length + 1)
	a.Set(length, value)
}

func (a *dynamicArray) Pop() common.Hash {
	length := a.Length()
	if length == 0 {
		return common.Hash{}
	}
	value := a.Get(length - 1)
	a.setLength(length - 1)
	return value
}

func (a *dynamicArray) Swap(i, j int) {
	if i >= a.Length() || j >= a.Length() {
		panic("index out of bounds")
	}
	iVal := a.Get(i)
	a.Set(i, a.Get(j))
	a.Set(j, iVal)
}

func (a *dynamicArray) GetReference(index int) Reference {
	return &reference{
		ds:  a.ds,
		key: a.key(index),
	}
}

func (a *dynamicArray) GetMap(index int) Mapping {
	return &mapping{
		ds: a.ds,
		id: a.key(index),
	}
}

func (a *dynamicArray) GetArray(index int) Array {
	return &array{
		ds: a.ds,
		id: a.key(index),
	}
}

var _ DynamicArray = (*dynamicArray)(nil)
