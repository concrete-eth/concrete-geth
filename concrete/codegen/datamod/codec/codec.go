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

package codec

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

var (
	// Redeclare constant from go-ethereum/accounts/abi to avoid importing
	// the module and having issues with tinygo.
	Uint256_0  = uint256.NewInt(0)
	Uint256_1  = uint256.NewInt(1)
	MaxUint256 = new(uint256.Int).Not(Uint256_0)
)

func EncodeAddress(_ int, address common.Address) []byte {
	return address.Bytes()
}

func DecodeAddress(_ int, data []byte) common.Address {
	return common.BytesToAddress(data)
}

func EncodeBool(_ int, b bool) []byte {
	if b {
		return []byte{1}
	}
	return []byte{0}
}

func DecodeBool(_ int, data []byte) bool {
	return data[0]&1 == byte(0x01)
}

func EncodeHash(_ int, hash common.Hash) []byte {
	return hash.Bytes()
}

func DecodeHash(_ int, data []byte) common.Hash {
	return common.BytesToHash(data)
}

func EncodeFixedBytes(size int, b []byte) []byte {
	return common.RightPadBytes(b, size)
}

func DecodeFixedBytes(_ int, data []byte) []byte {
	return data
}

func EncodeBytes(_ int, b []byte) []byte {
	return b
}

func DecodeBytes(_ int, data []byte) []byte {
	return data
}

func EncodeString(_ int, s string) []byte {
	return []byte(s)
}

func DecodeString(_ int, data []byte) string {
	return string(data)
}

func EncodeUint256(_ int, i *uint256.Int) []byte {
	b := i.Bytes32()
	return b[:]
}

func DecodeUint256(_ int, data []byte) *uint256.Int {
	return new(uint256.Int).SetBytes(data)
}

func EncodeUint8(_ int, value uint8) []byte {
	return []byte{value}
}

func EncodeUint16(_ int, value uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, value)
	return buf
}

func EncodeUint32(_ int, value uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, value)
	return buf
}

func EncodeUint64(_ int, value uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, value)
	return buf
}

func DecodeUint8(_ int, data []byte) uint8   { return data[0] }
func DecodeUint16(_ int, data []byte) uint16 { return binary.BigEndian.Uint16(data) }
func DecodeUint32(_ int, data []byte) uint32 { return binary.BigEndian.Uint32(data) }
func DecodeUint64(_ int, data []byte) uint64 { return binary.BigEndian.Uint64(data) }

func EncodeInt8(_ int, value int8) []byte {
	return []byte{byte(value)}
}

func EncodeInt16(_ int, value int16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(value))
	return buf
}

func EncodeInt32(_ int, value int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(value))
	return buf
}

func EncodeInt64(_ int, value int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(value))
	return buf
}

func DecodeInt8(_ int, data []byte) int8   { return int8(data[0]) }
func DecodeInt16(_ int, data []byte) int16 { return int16(binary.BigEndian.Uint16(data)) }
func DecodeInt32(_ int, data []byte) int32 { return int32(binary.BigEndian.Uint32(data)) }
func DecodeInt64(_ int, data []byte) int64 { return int64(binary.BigEndian.Uint64(data)) }
