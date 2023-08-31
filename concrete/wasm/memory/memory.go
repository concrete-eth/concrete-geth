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

package memory

import (
	"github.com/ethereum/go-ethereum/concrete/utils"
)

type MemPointer uint64

const (
	NullPointer = MemPointer(0)
)

func (pointer MemPointer) Uint64() uint64 {
	return uint64(pointer)
}

func (pointer MemPointer) IsNull() bool {
	return pointer == NullPointer
}

func (pointer MemPointer) Offset() uint32 {
	return uint32(pointer.Uint64() >> 32)
}

func (pointer MemPointer) Size() uint32 {
	return uint32(pointer.Uint64())
}

func (pointer *MemPointer) Pack(offset, size uint32) {
	*pointer = MemPointer(uint64(offset)<<32 | uint64(size))
}

func (pointer MemPointer) Unpack() (uint32, uint32) {
	return pointer.Offset(), pointer.Size()
}

func (pointer MemPointer) Encode() []byte {
	return utils.Uint64ToBytes(uint64(pointer))
}

func (pointer *MemPointer) Decode(data []byte) {
	*pointer = MemPointer(utils.BytesToUint64(data))
}

func PackPointers(pointers []MemPointer) []byte {
	output := make([]byte, len(pointers)*8)
	for i, pointer := range pointers {
		copy(output[i*8:], pointer.Encode())
	}
	return output
}

func UnpackPointers(data []byte) []MemPointer {
	var pointers []MemPointer
	for i := 0; i < len(data); i += 8 {
		var pointer MemPointer
		pointer.Decode(data[i : i+8])
		pointers = append(pointers, pointer)
	}
	return pointers
}

type Memory interface {
	Allocator() Allocator
	Read(MemPointer) []byte
	Write(data []byte) MemPointer
}

type Allocator interface {
	Malloc(size int) MemPointer
	Free(pointer MemPointer)
	Prune()
}

func PutValue(memory Memory, value []byte) MemPointer {
	return memory.Write(value)
}

func GetValue(memory Memory, pointer MemPointer) []byte {
	return memory.Read(pointer)
}

func PutValues(memory Memory, values [][]byte) MemPointer {
	if len(values) == 0 {
		return NullPointer
	}
	var pointers []MemPointer
	for _, v := range values {
		pointers = append(pointers, PutValue(memory, v))
	}
	packedPointers := PackPointers(pointers)
	return PutValue(memory, packedPointers)
}

func GetValues(memory Memory, pointer MemPointer, free bool) [][]byte {
	if pointer.IsNull() {
		return [][]byte{}
	}
	allocator := memory.Allocator()
	var values [][]byte
	valPointers := UnpackPointers(GetValue(memory, pointer))
	if free {
		allocator.Free(pointer)
	}
	for _, p := range valPointers {
		values = append(values, GetValue(memory, p))
		if free {
			allocator.Free(p)
		}
	}
	return values
}

func PutArgs(memory Memory, args [][]byte) MemPointer {
	return PutValues(memory, args)
}

func GetArgs(memory Memory, pointer MemPointer, free bool) [][]byte {
	return GetValues(memory, pointer, free)
}

func PutReturn(memory Memory, retValues [][]byte) MemPointer {
	return PutValues(memory, retValues)
}

func GetReturn(memory Memory, retPointer MemPointer, free bool) [][]byte {
	return GetValues(memory, retPointer, free)
}

func PutError(memory Memory, err error) MemPointer {
	return PutValue(memory, utils.EncodeError(err))
}

func GetError(memory Memory, errPointer MemPointer) error {
	return utils.DecodeError(GetValue(memory, errPointer))
}

func PutReturnWithError(memory Memory, retValues [][]byte, retErr error) MemPointer {
	retValues = append(retValues, utils.EncodeError(retErr))
	return PutReturn(memory, retValues)
}

func GetReturnWithError(memory Memory, retPointer MemPointer, free bool) ([][]byte, error) {
	retValues := GetReturn(memory, retPointer, free)
	err := utils.DecodeError(retValues[len(retValues)-1])
	return retValues[:len(retValues)-1], err
}
