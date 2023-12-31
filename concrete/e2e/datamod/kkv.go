/* Autogenerated file. Do not edit manually. */

package datamod

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod/codec"
	"github.com/ethereum/go-ethereum/concrete/crypto"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

// Reference imports to suppress errors if they are not used.
var (
	_ = big.NewInt
	_ = common.Big1
	_ = codec.EncodeAddress
)

// var (
//	KkvDefaultKey = crypto.Keccak256([]byte("datamod.v1.Kkv"))
// )

func defaultKey() []byte {
	return crypto.Keccak256([]byte("datamod.v1.Kkv"))
}

type KkvRow struct {
	lib.DatastoreStruct
}

func NewKkvRow(dsSlot lib.DatastoreSlot) *KkvRow {
	sizes := []int{32}
	return &KkvRow{*lib.NewDatastoreStruct(dsSlot, sizes)}
}

func (v *KkvRow) Get() (
	common.Hash,
) {
	return codec.DecodeHash(32, v.GetField(0))
}

func (v *KkvRow) Set(
	value common.Hash,
) {
	v.SetField(0, codec.EncodeHash(32, value))
}

func (v *KkvRow) GetValue() common.Hash {
	data := v.GetField(0)
	return codec.DecodeHash(32, data)
}

func (v *KkvRow) SetValue(value common.Hash) {
	data := codec.EncodeHash(32, value)
	v.SetField(0, data)
}

type Kkv struct {
	dsSlot lib.DatastoreSlot
}

func NewKkv(ds lib.Datastore) *Kkv {
	dsSlot := ds.Get(defaultKey())
	return &Kkv{dsSlot}
}

func NewKkvFromSlot(dsSlot lib.DatastoreSlot) *Kkv {
	return &Kkv{dsSlot}
}

func (m *Kkv) Get(
	key1 common.Hash,
	key2 common.Hash,
) *KkvRow {
	dsSlot := m.dsSlot.Mapping().GetNested(
		codec.EncodeHash(32, key1),
		codec.EncodeHash(32, key2),
	)
	return NewKkvRow(dsSlot)
}
