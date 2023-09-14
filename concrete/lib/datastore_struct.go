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
	"github.com/ethereum/go-ethereum/common"
)

type DatastoreStruct struct {
	store   DatastoreSlot
	arr     SlotArray
	offsets []int
	sizes   []int
}

func NewDatastoreStruct(store DatastoreSlot, sizes []int) *DatastoreStruct {
	var (
		offset  = 0
		offsets = make([]int, len(sizes))
		nSlots  int
	)
	for ii, size := range sizes {
		if size < 0 {
			panic("negative field size")
		}
		if size > 32 {
			panic("field size too large")
		}
		if offset/32 != (offset+size-1)/32 {
			offset = (offset/32 + 1) * 32
		}
		offsets[ii] = offset
		offset += size
	}
	nSlots = (offset + 31) / 32

	return &DatastoreStruct{
		store:   store,
		arr:     store.SlotArray([]int{nSlots}),
		offsets: offsets,
		sizes:   sizes,
	}
}

func (s *DatastoreStruct) GetField(index int) []byte {
	fieldSize := s.sizes[index]
	absOffset := s.offsets[index]
	slotIndex, slotOffset := absOffset/32, absOffset%32
	slotData := s.arr.Get(slotIndex).Bytes32()
	return slotData[slotOffset : slotOffset+fieldSize]
}

func (s *DatastoreStruct) SetField(index int, data []byte) {
	fieldSize := s.sizes[index]

	if len(data) != fieldSize {
		panic("invalid data size")
	}

	absOffset := s.offsets[index]
	slotIndex, slotOffset := absOffset/32, absOffset%32
	slotRef := s.arr.Get(slotIndex)

	var slotData common.Hash
	if fieldSize < 32 {
		slotData = slotRef.Bytes32()
		copy(slotData[slotOffset:slotOffset+fieldSize], data)
	} else {
		slotData = common.BytesToHash(data)
	}

	slotRef.SetBytes32(slotData)
}

func (s *DatastoreStruct) GetField_slot(index int) DatastoreSlot {
	absOffset := s.offsets[index]
	slotIndex := absOffset / 32
	return s.arr.Get(slotIndex)
}

func (s *DatastoreStruct) GetField_bytes(index int) []byte {
	slotRef := s.GetField_slot(index)
	return slotRef.Bytes()
}

func (s *DatastoreStruct) SetField_bytes(index int, data []byte) {
	slotRef := s.GetField_slot(index)
	slotRef.SetBytes(data)
}
