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

package utils

import (
	"encoding/binary"
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

func Uint64ToBytes(value uint64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, value)
	return data
}

func BytesToUint64(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

const (
	nil_error    = byte(0)
	notNil_error = byte(1)
)

func EncodeError(err error) []byte {
	if err == nil {
		return []byte{nil_error}
	}
	return append([]byte{notNil_error}, []byte(err.Error())...)
}

func DecodeError(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == nil_error {
		return nil
	}
	return errors.New(string(data[1:]))
}

func GetData(data []byte, start uint64, size uint64) []byte {
	length := uint64(len(data))
	if start > length {
		start = length
	}
	end := start + size
	if end > length {
		end = length
	}
	return common.RightPadBytes(data[start:end], int(size))
}

func SplitData(data []byte, size uint64) ([]byte, []byte) {
	if size > uint64(len(data)) {
		size = uint64(len(data))
	}
	return data[:size], data[size:]
}

func SplitInput(input []byte) ([]byte, []byte) {
	return SplitData(input, 4)
}
