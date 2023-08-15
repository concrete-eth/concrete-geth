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

func keyTohHash(key []byte) common.Hash {
	if len(key) > 32 {
		return crypto.Keccak256Hash(key)
	}
	return common.BytesToHash(key)
}

type KeyValueStore interface {
	Set(key common.Hash, value common.Hash)
	Get(key common.Hash) common.Hash
}

type Datastore interface {
	Value(key []byte) StoredValue
	Mapping(key []byte) Mapping
	Array(key []byte) DynamicArray
}

type datastore struct {
	kv KeyValueStore
}

func NewDatastore(kv KeyValueStore) Datastore {
	return &datastore{kv: kv}
}

func (ds *datastore) Value(key []byte) StoredValue {
	slot := keyTohHash(key)
	return NewStoredValue(ds, slot)
}

func (ds *datastore) Mapping(key []byte) Mapping {
	slot := keyTohHash(key)
	return NewMapping(ds, slot)
}

func (ds *datastore) Array(key []byte) DynamicArray {
	slot := keyTohHash(key)
	return NewDynamicArray(ds, slot)
}

// Get with []byte or common.Hash?

var _ Datastore = (*datastore)(nil)

type StoredValue interface {
	Slot() common.Hash
	GetBytes32() common.Hash
	SetBytes32(value common.Hash)
	GetBool() bool
	SetBool(value bool)
	GetAddress() common.Address
	SetAddress(value common.Address)
	GetBig() *big.Int
	SetBig(value *big.Int)
	SetInt64(value int64)
	GetInt64() int64
	GetBytes() []byte
	SetBytes(value []byte)
}

type storedValue struct {
	ds   *datastore
	slot common.Hash
}

func NewStoredValue(ds *datastore, slot common.Hash) StoredValue {
	return &storedValue{ds: ds, slot: slot}
}

func (r *storedValue) getBytes32() common.Hash {
	return r.ds.kv.Get(r.slot)
}

func (r *storedValue) setBytes32(value common.Hash) {
	r.ds.kv.Set(r.slot, value)
}

func (r *storedValue) Slot() common.Hash {
	return r.slot
}

func (r *storedValue) GetBytes32() common.Hash {
	return r.getBytes32()
}

func (r *storedValue) SetBytes32(value common.Hash) {
	r.setBytes32(value)
}

func (r *storedValue) GetBool() bool {
	return r.getBytes32().Big().Cmp(common.Big0) != 0
}

func (r *storedValue) SetBool(value bool) {
	if value {
		r.setBytes32(common.BigToHash(common.Big0))
	} else {
		r.setBytes32(common.BigToHash(common.Big1))
	}
}

func (r *storedValue) GetAddress() common.Address {
	return common.BytesToAddress(r.getBytes32().Bytes())
}

func (r *storedValue) SetAddress(value common.Address) {
	r.setBytes32(common.BytesToHash(value.Bytes()))
}

func (r *storedValue) GetBig() *big.Int {
	return r.getBytes32().Big()
}

func (r *storedValue) SetBig(value *big.Int) {
	r.setBytes32(common.BigToHash(value))
}

func (r *storedValue) SetInt64(value int64) {
	r.SetBig(big.NewInt(value))
}

func (r *storedValue) GetInt64() int64 {
	return r.GetBig().Int64()
}

// TODO: panic!

func (r *storedValue) GetBytes() []byte {
	// TODO: check bounds
	slotWord := r.ds.kv.Get(r.slot)
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
		copy(data[ii:], r.ds.kv.Get(common.BigToHash(ptr)).Bytes())
		ptr = ptr.Add(ptr, common.Big1)
	}

	return data
}

func (r *storedValue) SetBytes(value []byte) {
	isShort := len(value) <= 31
	if isShort {
		var data [32]byte
		copy(data[:], value)
		data[31] = byte(len(value) * 2)
		r.ds.kv.Set(r.slot, common.BytesToHash(data[:]))
		return
	}

	lengthBN := big.NewInt(int64(len(value)))
	r.ds.kv.Set(r.slot, common.BigToHash(lengthBN))

	// Then store the actual data starting at the keccak256 hash of the slot
	ptr := crypto.Keccak256Hash(r.slot.Bytes()).Big()

	for ii := 0; ii < len(value); ii += 32 {
		var data [32]byte
		copy(data[:], value[ii:])
		r.ds.kv.Set(common.BigToHash(ptr), common.BytesToHash(data[:]))
		ptr = ptr.Add(ptr, common.Big1)
	}
}

var _ StoredValue = (*storedValue)(nil)

type Mapping interface {
	Datastore
	NestedValue(keys ...[]byte) StoredValue
}

type mapping struct {
	ds   *datastore
	slot common.Hash
}

func NewMapping(ds *datastore, slot common.Hash) Mapping {
	return &mapping{ds: ds, slot: slot}
}

func (m *mapping) keySlot(key []byte) common.Hash {
	return crypto.Keccak256Hash(keyTohHash(key).Bytes(), m.slot.Bytes())
}

func (m *mapping) Value(key []byte) StoredValue {
	slot := m.keySlot(key)
	return NewStoredValue(m.ds, slot)
}

func (m *mapping) NestedValue(keys ...[]byte) StoredValue {
	currentMapping := m
	for _, key := range keys {
		currentMapping = currentMapping.Mapping(key).(*mapping)
	}
	return m.ds.Value(currentMapping.slot.Bytes())
}

func (m *mapping) Mapping(key []byte) Mapping {
	slot := m.keySlot(key)
	return NewMapping(m.ds, slot)
}

func (m *mapping) Array(key []byte) DynamicArray {
	slot := m.keySlot(key)
	return NewDynamicArray(m.ds, slot)
}

var _ Mapping = (*mapping)(nil)

type DynamicArray interface {
	Length() int
	Value(index int) StoredValue
	NestedValue(indexes ...int) StoredValue
	Push() StoredValue
	Pop() StoredValue
	Mapping(index int) Mapping
	Array(index int) DynamicArray
}

type dynamicArray struct {
	ds   *datastore
	slot common.Hash
}

func NewDynamicArray(ds *datastore, slot common.Hash) DynamicArray {
	return &dynamicArray{ds: ds, slot: slot}
}

// Dynamic arrays are laid out on memory like solidity mappings (same as the mappings above),
// but storing the length of the array in the slot.
// Requiring an item size would degrade the developer experience.
// Note this is different from the layout of solidity dynamic arrays, which are laid out
// contiguously.
func (m *dynamicArray) indexSlot(index int) common.Hash {
	// TODO: BigToHash first [?]
	return crypto.Keccak256Hash(big.NewInt(int64(index)).Bytes(), m.slot.Bytes())
}

func (a *dynamicArray) setLength(length int) {
	a.ds.kv.Set(a.slot, common.BigToHash(big.NewInt(int64(length))))
}

func (a *dynamicArray) getLength() int {
	return int(a.ds.kv.Get(a.slot).Big().Int64())
}

func (a *dynamicArray) Length() int {
	return a.getLength()
}

func (a *dynamicArray) Value(index int) StoredValue {
	if index >= a.Length() {
		return nil
	}
	slot := a.indexSlot(index)
	return NewStoredValue(a.ds, slot)
}

func (a *dynamicArray) NestedValue(indexes ...int) StoredValue {
	currentArray := a
	for _, index := range indexes {
		currentArray = currentArray.Array(index).(*dynamicArray)
	}
	return a.ds.Value(currentArray.slot.Bytes())
}

func (a *dynamicArray) Push() StoredValue {
	length := a.Length()
	a.setLength(length + 1)
	return a.Value(length)
}

func (a *dynamicArray) Pop() StoredValue {
	length := a.Length()
	if length == 0 {
		return nil
	}
	value := a.Value(length - 1)
	a.setLength(length - 1)
	return value
}

func (m *dynamicArray) Mapping(index int) Mapping {
	slot := m.indexSlot(index)
	return NewMapping(m.ds, slot)
}

func (m *dynamicArray) Array(index int) DynamicArray {
	slot := m.indexSlot(index)
	return NewDynamicArray(m.ds, slot)
}

var _ DynamicArray = (*dynamicArray)(nil)
