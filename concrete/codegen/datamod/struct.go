package datamod

import (
	"github.com/ethereum/go-ethereum/concrete/lib"
)

type DatamodStruct struct {
	storage lib.SlotArray
	offsets []int
	sizes   []int
}

func NewDatamodStruct(slot lib.StorageSlot, sizes []int) *DatamodStruct {
	offsets := make([]int, len(sizes))
	offset := 0
	nSlots := 0
	for i := 1; i < len(sizes); i++ {
		size := sizes[i]
		if offset/32 < (offset+size)/32 {
			offset = (offset/32 + 1) * 32
			nSlots++
		}
		offset += size
		offsets[i] = offset
	}
	return &DatamodStruct{
		storage: slot.SlotArray([]int{nSlots}),
		offsets: offsets,
		sizes:   sizes,
	}
}

func (s *DatamodStruct) GetField(index int) []byte {
	absOffset := s.offsets[index]
	slotIdx := absOffset / 32
	slotOffset := absOffset % 32

	slotValue := s.storage.Value(slotIdx).Bytes32()
	size := s.sizes[index]
	return slotValue[slotOffset : slotOffset+size]
}

func (s *DatamodStruct) SetField(index int, data []byte) {
	absOffset := s.offsets[index]
	slotIdx := absOffset / 32
	slotOffset := absOffset % 32

	slot := s.storage.Value(slotIdx)
	slotValue := slot.Bytes32()
	size := s.sizes[index]
	copy(slotValue[slotOffset:slotOffset+size], data)
	slot.SetBytes32(slotValue)
}
