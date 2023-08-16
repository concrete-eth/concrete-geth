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

func newDatastore(kv KeyValueStore) *datastore {
	return &datastore{kv: kv}
}

func (ds *datastore) value(key []byte) *storedValue {
	slot := keyTohHash(key)
	return newStoredValue(ds, slot)
}

func (ds *datastore) mapping(key []byte) *mapping {
	slot := keyTohHash(key)
	return newMapping(ds, slot)
}

func (ds *datastore) array(key []byte) *dynamicArray {
	slot := keyTohHash(key)
	return newDynamicArray(ds, slot)
}

func NewDatastore(kv KeyValueStore) Datastore {
	return newDatastore(kv)
}

func (ds *datastore) Value(key []byte) StoredValue {
	return ds.value(key)
}

func (ds *datastore) Mapping(key []byte) Mapping {
	return ds.mapping(key)
}

func (ds *datastore) Array(key []byte) DynamicArray {
	return ds.array(key)
}

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

func newStoredValue(ds *datastore, slot common.Hash) *storedValue {
	return &storedValue{ds: ds, slot: slot}
}

func (r *storedValue) getBytes32() common.Hash {
	return r.ds.kv.Get(r.slot)
}

func (r *storedValue) setBytes32(value common.Hash) {
	r.ds.kv.Set(r.slot, value)
}

func (r *storedValue) getBytes() []byte {
	slotWord := r.ds.kv.Get(r.slot)
	lsb := slotWord[len(slotWord)-1]
	isShort := lsb&1 == 0
	if isShort {
		length := int(lsb) / 2
		return slotWord[:length]
	}

	length := slotWord.Big().Int64()
	// TODO: cache hash [?]
	ptr := crypto.Keccak256Hash(r.slot.Bytes()).Big()

	data := make([]byte, length)
	for ii := 0; ii < len(data); ii += 32 {
		copy(data[ii:], r.ds.kv.Get(common.BigToHash(ptr)).Bytes())
		ptr = ptr.Add(ptr, common.Big1)
	}

	return data
}

func (r *storedValue) setBytes(value []byte) {
	isShort := len(value) <= 31
	if isShort {
		var data common.Hash
		copy(data[:], value)
		data[31] = byte(len(value) * 2)
		r.ds.kv.Set(r.slot, data)
		return
	}

	lengthBN := big.NewInt(int64(len(value)))
	r.ds.kv.Set(r.slot, common.BigToHash(lengthBN))

	ptr := crypto.Keccak256Hash(r.slot.Bytes()).Big()
	for ii := 0; ii < len(value); ii += 32 {
		var data common.Hash
		copy(data[:], value[ii:])
		r.ds.kv.Set(common.BigToHash(ptr), data)
		ptr = ptr.Add(ptr, common.Big1)
	}
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

func (r *storedValue) GetBytes() []byte {
	return r.getBytes()
}

func (r *storedValue) SetBytes(value []byte) {
	r.setBytes(value)
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

func newMapping(ds *datastore, slot common.Hash) *mapping {
	return &mapping{ds: ds, slot: slot}
}

func (m *mapping) keySlot(key []byte) common.Hash {
	return crypto.Keccak256Hash(keyTohHash(key).Bytes(), m.slot.Bytes())
}

func (m *mapping) value(key []byte) *storedValue {
	slot := m.keySlot(key)
	return newStoredValue(m.ds, slot)
}

func (m *mapping) nestedValue(keys [][]byte) *storedValue {
	currentMapping := m
	for _, key := range keys {
		currentMapping = currentMapping.mapping(key)
	}
	return m.ds.value(currentMapping.slot.Bytes())
}

func (m *mapping) mapping(key []byte) *mapping {
	slot := m.keySlot(key)
	return newMapping(m.ds, slot)
}

func (m *mapping) array(key []byte) *dynamicArray {
	slot := m.keySlot(key)
	return newDynamicArray(m.ds, slot)
}

func (m *mapping) Value(key []byte) StoredValue {
	return m.value(key)
}

func (m *mapping) NestedValue(keys ...[]byte) StoredValue {
	return m.nestedValue(keys)
}

func (m *mapping) Mapping(key []byte) Mapping {
	return m.mapping(key)
}

func (m *mapping) Array(key []byte) DynamicArray {
	return m.array(key)
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

func newDynamicArray(ds *datastore, slot common.Hash) *dynamicArray {
	return &dynamicArray{ds: ds, slot: slot}
}

// Dynamic arrays are laid out on memory like solidity mappings (same as the mappings above),
// but storing the length of the array in the slot.
// Requiring an item size would degrade the developer experience.
// Note this is different from the layout of solidity dynamic arrays, which are laid out
// contiguously.
func (m *dynamicArray) indexSlot(index int) common.Hash {
	return crypto.Keccak256Hash(
		common.BigToHash(big.NewInt(int64(index))).Bytes(),
		m.slot.Bytes(),
	)
}

func (a *dynamicArray) setLength(length int) {
	a.ds.kv.Set(a.slot, common.BigToHash(big.NewInt(int64(length))))
}

func (a *dynamicArray) getLength() int {
	return int(a.ds.kv.Get(a.slot).Big().Int64())
}

func (a *dynamicArray) value(index int) *storedValue {
	if index >= a.getLength() {
		return nil
	}
	slot := a.indexSlot(index)
	return newStoredValue(a.ds, slot)
}

func (a *dynamicArray) nestedValue(indexes []int) *storedValue {
	currentArray := a
	for _, index := range indexes {
		currentArray = currentArray.array(index)
	}
	return a.ds.value(currentArray.slot.Bytes())
}

func (a *dynamicArray) mapping(index int) *mapping {
	slot := a.indexSlot(index)
	return newMapping(a.ds, slot)
}

func (a *dynamicArray) array(index int) *dynamicArray {
	slot := a.indexSlot(index)
	return newDynamicArray(a.ds, slot)
}

func (a *dynamicArray) Length() int {
	return a.getLength()
}

func (a *dynamicArray) Value(index int) StoredValue {
	return a.value(index)
}

func (a *dynamicArray) NestedValue(indexes ...int) StoredValue {
	return a.nestedValue(indexes)
}

func (m *dynamicArray) Mapping(index int) Mapping {
	return m.mapping(index)
}

func (m *dynamicArray) Array(index int) DynamicArray {
	return m.array(index)
}

func (a *dynamicArray) Push() StoredValue {
	length := a.getLength()
	a.setLength(length + 1)
	return a.value(length)
}

func (a *dynamicArray) Pop() StoredValue {
	length := a.getLength()
	if length == 0 {
		return nil
	}
	value := a.value(length - 1)
	a.setLength(length - 1)
	return value
}

var _ DynamicArray = (*dynamicArray)(nil)
