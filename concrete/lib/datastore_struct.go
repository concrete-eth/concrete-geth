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

type StorageStruct struct {
	arr     SlotArray
	offsets []int
	sizes   []int
}

func NewStorageStruct(slot StorageSlot, sizes []int) *StorageStruct {
	var (
		offset = 0
		nSlots = 0
	)
	offsets := make([]int, len(sizes))
	for i := 1; i < len(sizes); i++ {
		size := sizes[i]
		if offset/32 < (offset+size)/32 {
			offset = (offset/32 + 1) * 32
			nSlots++
		}
		offset += size
		offsets[i] = offset
	}
	if offset%32 != 0 {
		nSlots++
	}
	return &StorageStruct{
		arr:     slot.SlotArray([]int{nSlots}),
		offsets: offsets,
		sizes:   sizes,
	}
}

func (s *StorageStruct) GetField(index int) []byte {
	absOffset := s.offsets[index]
	slotIndex, slotOffset := absOffset/32, absOffset%32

	slotData := s.arr.Value(slotIndex).Bytes32()
	fieldSize := s.sizes[index]
	return slotData[slotOffset : slotOffset+fieldSize]
}

func (s *StorageStruct) SetField(index int, data []byte) {
	absOffset := s.offsets[index]
	slotIndex, slotOffset := absOffset/32, absOffset%32

	slotRef := s.arr.Value(slotIndex)
	slotData := slotRef.Bytes32()
	fieldSize := s.sizes[index]
	copy(slotData[slotOffset:slotOffset+fieldSize], data)

	slotRef.SetBytes32(slotData)
}
