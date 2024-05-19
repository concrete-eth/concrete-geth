/* Autogenerated file. Do not edit manually. */

package testdata

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod/codec"
	"github.com/ethereum/go-ethereum/concrete/crypto"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/holiman/uint256"
)

// Reference imports to suppress errors if they are not used.
var (
	_ = common.Big1
	_ = codec.EncodeAddress
	_ = uint256.NewInt
)

// var (
//	KeyedWithKeylessTableValueDefaultKey = crypto.Keccak256([]byte("datamod.v1.KeyedWithKeylessTableValue"))
// )

func KeyedWithKeylessTableValueDefaultKey() []byte {
	return crypto.Keccak256([]byte("datamod.v1.KeyedWithKeylessTableValue"))
}

type KeyedWithKeylessTableValueRow struct {
	lib.DatastoreStruct
}

func NewKeyedWithKeylessTableValueRow(dsSlot lib.DatastoreSlot) *KeyedWithKeylessTableValueRow {
	sizes := []int{32}
	return &KeyedWithKeylessTableValueRow{*lib.NewDatastoreStruct(dsSlot, sizes)}
}

func (v *KeyedWithKeylessTableValueRow) Get() (
	*KeylessTable,
) {
	return NewKeylessTableFromSlot(v.GetField_slot(0))
}

func (v *KeyedWithKeylessTableValueRow) Set(
) {
}

func (v *KeyedWithKeylessTableValueRow) GetValueTable() *KeylessTable {
	dsSlot := v.GetField_slot(0)
	return NewKeylessTableFromSlot(dsSlot)
}

type KeyedWithKeylessTableValue struct {
	dsSlot lib.DatastoreSlot
}

func NewKeyedWithKeylessTableValue(ds lib.Datastore) *KeyedWithKeylessTableValue {
	dsSlot := ds.Get(KeyedWithKeylessTableValueDefaultKey())
	return &KeyedWithKeylessTableValue{dsSlot}
}

func NewKeyedWithKeylessTableValueFromSlot(dsSlot lib.DatastoreSlot) *KeyedWithKeylessTableValue {
	return &KeyedWithKeylessTableValue{dsSlot}
}

func (m *KeyedWithKeylessTableValue) Get(
	keyUint *uint256.Int,
) *KeyedWithKeylessTableValueRow {
	dsSlot := m.dsSlot.Mapping().GetNested(
		codec.EncodeUint256(32, keyUint),
	)
	return NewKeyedWithKeylessTableValueRow(dsSlot)
}
