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
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/crypto"
)

var (
	// Redeclare constant from go-ethereum/accounts/abi to avoid importing
	// the module and having issues with tinygo.
	MaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 256), common.Big1)
)

type KeyValueStore interface {
	Set(key common.Hash, value common.Hash)
	Get(key common.Hash) common.Hash
}

type envPersistentKV struct {
	env api.Environment
}

func newEnvPersistentKeyValueStore(env api.Environment) *envPersistentKV {
	return &envPersistentKV{env: env}
}

func (kv *envPersistentKV) Set(key common.Hash, value common.Hash) {
	kv.env.PersistentStore(key, value)
}

func (kv *envPersistentKV) Get(key common.Hash) common.Hash {
	return kv.env.PersistentLoad(key)
}

var _ KeyValueStore = (*envPersistentKV)(nil)

type envEphemeralKV struct {
	env api.Environment
}

func newEnvEphemeralKeyValueStore(env api.Environment) *envEphemeralKV {
	return &envEphemeralKV{env: env}
}

func (kv *envEphemeralKV) Set(key common.Hash, value common.Hash) {
	kv.env.EphemeralStore_Unsafe(key, value)
}

func (kv *envEphemeralKV) Get(key common.Hash) common.Hash {
	return kv.env.EphemeralLoad_Unsafe(key)
}

var _ KeyValueStore = (*envEphemeralKV)(nil)

type Datastore interface {
	Get(key []byte) DatastoreSlot
}

type datastore struct {
	kv KeyValueStore
}

func newDatastore(kv KeyValueStore) *datastore {
	return &datastore{kv: kv}
}

func (ds *datastore) value(key []byte) *dsSlot {
	if len(key) > 32 {
		key = crypto.Keccak256(key)
	}
	slot := common.BytesToHash(key)
	return newDatastoreSlot(ds, slot)
}

func (ds *datastore) Get(key []byte) DatastoreSlot {
	return ds.value(key)
}

var _ Datastore = (*datastore)(nil)

func NewPersistentDatastore(env api.Environment) Datastore {
	kv := newEnvPersistentKeyValueStore(env)
	return newDatastore(kv)
}

func NewEphemeralDatastore(env api.Environment) Datastore {
	kv := newEnvEphemeralKeyValueStore(env)
	return newDatastore(kv)
}

func NewDatastore(env api.Environment) Datastore {
	return NewPersistentDatastore(env)
}

type DatastoreSlot interface {
	Datastore() Datastore
	Slot() common.Hash

	SlotArray(length []int) SlotArray
	BytesArray(length []int, itemSize int) BytesArray
	Mapping() Mapping
	DynamicArray() DynamicArray

	Bytes32() common.Hash
	SetBytes32(value common.Hash)
	Bool() bool
	SetBool(value bool)
	Address() common.Address
	SetAddress(value common.Address)
	BigUint() *big.Int
	SetBigUint(value *big.Int)
	BigInt() *big.Int
	SetBigInt(value *big.Int)
	Uint64() uint64
	SetUint64(value uint64)
	Int64() int64
	SetInt64(value int64)
	Bytes() []byte
	SetBytes(value []byte)
}

type dsSlot struct {
	ds       *datastore
	slot     common.Hash
	slotHash *common.Hash
}

func newDatastoreSlot(ds *datastore, slot common.Hash) *dsSlot {
	return &dsSlot{ds: ds, slot: slot}
}

func (r *dsSlot) getSlotHash() common.Hash {
	if r.slotHash == nil {
		hash := crypto.Keccak256Hash(r.slot.Bytes())
		r.slotHash = &hash
	}
	return *r.slotHash
}

func (r *dsSlot) slotArray(length []int) *slotArray {
	return newSlotArray(r, length)
}

func (r *dsSlot) bytesArray(length []int, itemSize int) *bytesArray {
	return newBytesArray(r, length, itemSize)
}

func (r *dsSlot) mapping() *mapping {
	return newMapping(r)
}

func (r *dsSlot) array() *dynamicArray {
	return newDynamicArray(r)
}

func (r *dsSlot) getBytes32() common.Hash {
	return r.ds.kv.Get(r.slot)
}

func (r *dsSlot) setBytes32(value common.Hash) {
	r.ds.kv.Set(r.slot, value)
}

func (r *dsSlot) getBytes() []byte {
	slotData := r.ds.kv.Get(r.slot)
	lsb := slotData[len(slotData)-1]
	isShort := lsb&1 == 0
	if isShort {
		length := int(lsb) / 2
		return slotData[:length]
	}

	length := slotData.Big().Int64()
	ptr := r.getSlotHash().Big()

	data := make([]byte, length)
	for ii := 0; ii < len(data); ii += 32 {
		copy(data[ii:], r.ds.kv.Get(common.BigToHash(ptr)).Bytes())
		ptr = ptr.Add(ptr, common.Big1)
	}

	return data
}

func (r *dsSlot) setBytes(value []byte) {
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

	ptr := r.getSlotHash().Big()
	for ii := 0; ii < len(value); ii += 32 {
		var data common.Hash
		copy(data[:], value[ii:])
		r.ds.kv.Set(common.BigToHash(ptr), data)
		ptr = ptr.Add(ptr, common.Big1)
	}
}

func (r *dsSlot) Datastore() Datastore {
	return r.ds
}

func (r *dsSlot) Slot() common.Hash {
	return r.slot
}

func (r *dsSlot) SlotArray(length []int) SlotArray {
	return r.slotArray(length)
}

func (r *dsSlot) BytesArray(length []int, itemSize int) BytesArray {
	return r.bytesArray(length, itemSize)
}

func (r *dsSlot) Mapping() Mapping {
	return r.mapping()
}

func (r *dsSlot) DynamicArray() DynamicArray {
	return r.array()
}

func (r *dsSlot) Bytes32() common.Hash {
	return r.getBytes32()
}

func (r *dsSlot) SetBytes32(value common.Hash) {
	r.setBytes32(value)
}

func (r *dsSlot) Bool() bool {
	return r.getBytes32().Big().Cmp(common.Big0) != 0
}

func (r *dsSlot) SetBool(value bool) {
	if value {
		r.setBytes32(common.BigToHash(common.Big1))
	} else {
		r.setBytes32(common.BigToHash(common.Big0))
	}
}

func (r *dsSlot) Address() common.Address {
	return common.BytesToAddress(r.getBytes32().Bytes())
}

func (r *dsSlot) SetAddress(value common.Address) {
	r.setBytes32(common.BytesToHash(value.Bytes()))
}

func (r *dsSlot) BigUint() *big.Int {
	return r.getBytes32().Big()
}

func (r *dsSlot) SetBigUint(value *big.Int) {
	r.setBytes32(common.BigToHash(value))
}

func (r *dsSlot) BigInt() *big.Int {
	ret := r.getBytes32().Big()
	if ret.Bit(255) == 1 {
		ret.Add(MaxUint256, new(big.Int).Neg(ret))
		ret.Add(ret, common.Big1)
		ret.Neg(ret)
	}
	return ret
}

func (r *dsSlot) SetBigInt(value *big.Int) {
	value = new(big.Int).Set(value)
	math.U256Bytes(value)
	r.setBytes32(common.BigToHash(value))
}

func (r *dsSlot) SetUint64(value uint64) {
	r.SetBigUint(new(big.Int).SetUint64(value))
}

func (r *dsSlot) Uint64() uint64 {
	return r.BigUint().Uint64()
}

func (r *dsSlot) SetInt64(value int64) {
	r.SetBigInt(big.NewInt(value))
}

func (r *dsSlot) Int64() int64 {
	return r.BigInt().Int64()
}

func (r *dsSlot) Bytes() []byte {
	return r.getBytes()
}

func (r *dsSlot) SetBytes(value []byte) {
	r.setBytes(value)
}

var _ DatastoreSlot = (*dsSlot)(nil)

type SlotArray interface {
	Length() int
	Get(index ...int) DatastoreSlot
	SlotArray(index ...int) SlotArray
}

type slotArray struct {
	dsSlot     *dsSlot
	length     []int
	flatLength []int
}

func newSlotArray(dsSlot *dsSlot, length []int) *slotArray {
	if len(length) == 0 {
		return nil
	}
	flatLength := make([]int, len(length))
	for ii := len(length) - 1; ii >= 0; ii-- {
		if length[ii] <= 0 {
			return nil
		}
		if ii == len(length)-1 {
			flatLength[ii] = 1
		} else {
			flatLength[ii] = flatLength[ii+1] * length[ii+1]
		}
	}
	return &slotArray{dsSlot: dsSlot, length: length, flatLength: flatLength}
}

func (a *slotArray) indexSlot(index []int) *common.Hash {
	if len(index) > len(a.length) {
		return nil
	}
	flatIndex := 0
	for ii := 0; ii < len(index); ii++ {
		if index[ii] >= a.length[ii] || index[ii] < 0 {
			return nil
		}
		flatIndex += index[ii] * a.flatLength[ii]
	}
	slotIndex := new(big.Int).Add(big.NewInt(int64(flatIndex)), a.dsSlot.slot.Big())
	slot := common.BigToHash(slotIndex)
	return &slot
}

func (a *slotArray) getLength() int {
	return a.length[0]
}

func (a *slotArray) value(index []int) *dsSlot {
	if len(index) != len(a.length) {
		return nil
	}
	slot := a.indexSlot(index)
	if slot == nil {
		return nil
	}
	return newDatastoreSlot(a.dsSlot.ds, *slot)
}

func (a *slotArray) slotArray(index []int) *slotArray {
	if len(index) == 0 {
		return a
	}
	if len(index) >= len(a.length) {
		return nil
	}
	slot := a.indexSlot(index)
	if slot == nil {
		return nil
	}
	length := a.length[len(index):]
	dsSlot := newDatastoreSlot(a.dsSlot.ds, *slot)
	return newSlotArray(dsSlot, length)
}

func (a *slotArray) Length() int {
	return a.getLength()
}

func (a *slotArray) Get(index ...int) DatastoreSlot {
	return a.value(index)
}

func (a *slotArray) SlotArray(index ...int) SlotArray {
	return a.slotArray(index)
}

var _ SlotArray = (*slotArray)(nil)

type BytesArray interface {
	Length() int
	Get(index ...int) []byte
	BytesArray(index ...int) BytesArray
}

type bytesArray struct {
	arr      *slotArray
	itemSize int
}

func newBytesArray(dsSlot *dsSlot, _length []int, itemSize int) *bytesArray {
	// Validate inputs
	if len(_length) == 0 || itemSize == 0 {
		return nil
	}

	// Copy length because it might be modified
	length := make([]int, len(_length))
	copy(length, _length)

	// Convert length to the length of the underlying slot array
	itemsPerSlot := 32 / itemSize
	if itemsPerSlot > 1 {
		length[len(length)-1] /= itemsPerSlot
	} else if itemsPerSlot < 1 {
		slotsPerItem := (itemSize + 31) / 32
		length[len(length)-1] *= slotsPerItem
	}
	arr := newSlotArray(dsSlot, length)
	return &bytesArray{arr: arr, itemSize: itemSize}
}

func (a *bytesArray) getLength() int {
	return a.arr.getLength()
}

func (a *bytesArray) value(_index []int) []byte {
	// Validate inputs
	if len(_index) != len(a.arr.length) {
		return nil
	}

	// Copy index because it might be modified
	index := make([]int, len(_index))
	copy(index, _index)

	// Map index to underlying slot array
	itemsPerSlot := 32 / a.itemSize
	slotsPerItem := (a.itemSize + 31) / 32

	if itemsPerSlot > 1 {
		lastIndex := index[len(index)-1]
		slotIndex, slotItemOffset := lastIndex/itemsPerSlot, lastIndex%itemsPerSlot
		index[len(index)-1] = slotIndex
		slotRef := a.arr.value(index)
		if slotRef == nil {
			return nil
		}
		data := slotRef.getBytes32().Bytes()
		return data[slotItemOffset*a.itemSize : (slotItemOffset+1)*a.itemSize]
	} else if itemsPerSlot < 1 {
		index[len(index)-1] *= slotsPerItem
	}

	// Read data from underlying slot array
	data := make([]byte, a.itemSize)
	for ii := 0; ii < a.itemSize; ii++ {
		slotRef := a.arr.value(index)
		if slotRef == nil {
			return nil
		}
		value := slotRef.getBytes32().Bytes()
		copy(data[ii*32:], value)
		index[len(index)-1]++
	}
	return data
}

func (a *bytesArray) bytesArray(index []int) *bytesArray {
	if len(index) == 0 {
		return a
	}
	if len(index) >= len(a.arr.length) {
		return nil
	}
	slot := a.arr.indexSlot(index)
	if slot == nil {
		return nil
	}
	length := a.arr.length[len(index):]
	dsSlot := newDatastoreSlot(a.arr.dsSlot.ds, *slot)
	return newBytesArray(dsSlot, length, a.itemSize)
}

func (a *bytesArray) Length() int {
	return a.getLength()
}

func (a *bytesArray) Get(index ...int) []byte {
	return a.value(index)
}

func (a *bytesArray) BytesArray(index ...int) BytesArray {
	return a.bytesArray(index)
}

var _ BytesArray = (*bytesArray)(nil)

type Mapping interface {
	Datastore
	GetNested(keys ...[]byte) DatastoreSlot
}

type mapping struct {
	dsSlot *dsSlot
}

func newMapping(dsSlot *dsSlot) *mapping {
	return &mapping{dsSlot: dsSlot}
}

func (m *mapping) keySlot(key []byte) common.Hash {
	return crypto.Keccak256Hash(key, m.dsSlot.slot.Bytes())
}

func (m *mapping) value(key []byte) *dsSlot {
	slot := m.keySlot(key)
	return newDatastoreSlot(m.dsSlot.ds, slot)
}

func (m *mapping) mapping(key []byte) *mapping {
	slot := m.keySlot(key)
	dsSlot := newDatastoreSlot(m.dsSlot.ds, slot)
	return newMapping(dsSlot)
}

func (m *mapping) nestedValue(keys [][]byte) *dsSlot {
	if len(keys) == 0 {
		return nil
	}
	currentMapping := m
	nestedKeys, mapKey := keys[:len(keys)-1], keys[len(keys)-1]
	for _, key := range nestedKeys {
		currentMapping = currentMapping.mapping(key)
	}
	return currentMapping.value(mapKey)
}

func (m *mapping) Get(key []byte) DatastoreSlot {
	return m.value(key)
}

func (m *mapping) GetNested(keys ...[]byte) DatastoreSlot {
	return m.nestedValue(keys)
}

var _ Mapping = (*mapping)(nil)

type DynamicArray interface {
	Length() uint64
	Get(index uint64) DatastoreSlot
	GetNested(indexes ...uint64) DatastoreSlot
	Push() DatastoreSlot
	Pop() DatastoreSlot
}

type dynamicArray struct {
	dsSlot *dsSlot
}

func newDynamicArray(dsSlot *dsSlot) *dynamicArray {
	return &dynamicArray{dsSlot: dsSlot}
}

// Dynamic arrays are laid out on memory like solidity mappings (same as the mappings above),
// but storing the length of the array in the slot.
// Note this is different from the layout of solidity dynamic arrays, which are laid out
// contiguously.
func (m *dynamicArray) indexKey(index uint64) []byte {
	if index >= m.getLength() {
		return nil
	}
	bigIndex := new(big.Int).SetUint64(index)
	return common.BigToHash(bigIndex).Bytes()
}

func (a *dynamicArray) setLength(length uint64) {
	a.dsSlot.SetUint64(length)
}

func (a *dynamicArray) getLength() uint64 {
	return a.dsSlot.Uint64()
}

func (a *dynamicArray) value(index uint64) *dsSlot {
	key := a.indexKey(index)
	if key == nil {
		return nil
	}
	return a.dsSlot.mapping().value(key)
}

func (a *dynamicArray) nestedValue(indexes []uint64) *dsSlot {
	if len(indexes) == 0 {
		return nil
	}
	keys := make([][]byte, len(indexes))
	for ii := 0; ii < len(indexes); ii++ {
		keys[ii] = a.indexKey(indexes[ii])
		if keys[ii] == nil {
			return nil
		}
	}
	return a.dsSlot.mapping().nestedValue(keys)
}

func (a *dynamicArray) Length() uint64 {
	return a.getLength()
}

func (a *dynamicArray) Get(index uint64) DatastoreSlot {
	return a.value(index)
}

func (a *dynamicArray) GetNested(indexes ...uint64) DatastoreSlot {
	return a.nestedValue(indexes)
}

func (a *dynamicArray) Push() DatastoreSlot {
	length := a.getLength()
	a.setLength(length + 1)
	return a.value(length)
}

func (a *dynamicArray) Pop() DatastoreSlot {
	length := a.getLength()
	if length == 0 {
		return nil
	}
	value := a.value(length - 1)
	a.setLength(length - 1)
	return value
}

var _ DynamicArray = (*dynamicArray)(nil)
