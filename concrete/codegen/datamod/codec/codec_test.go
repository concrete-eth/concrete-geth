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
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestCodec(t *testing.T) {
	r := require.New(t)

	t.Run("address", func(t *testing.T) {
		addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
		encoded := EncodeAddress(20, addr)
		decoded := DecodeAddress(20, encoded)
		r.Equal(addr, decoded)
	})

	t.Run("bool", func(t *testing.T) {
		encoded := EncodeBool(1, true)
		decoded := DecodeBool(1, encoded)
		r.True(decoded)
	})

	t.Run("hash", func(t *testing.T) {
		hash := common.HexToHash("0x1234567890123456789012345678901234567890123456789012345678901234")
		encoded := EncodeHash(32, hash)
		decoded := DecodeHash(32, encoded)
		r.Equal(hash, decoded)
	})

	t.Run("bytesN", func(t *testing.T) {
		for size := 1; size < 32; size++ {
			b := make([]byte, size)
			for i := 0; i < size; i++ {
				b[i] = byte(i)
			}
			encoded := EncodeFixedBytes(size, b)
			decoded := DecodeFixedBytes(size, encoded)
			r.Equal(b, decoded)
		}
	})

	t.Run("bytes", func(t *testing.T) {
		b := []byte{0x01, 0x02, 0x03}
		encoded := EncodeBytes(-1, b)
		decoded := DecodeBytes(-1, encoded)
		r.Equal(b, decoded)
	})

	t.Run("string", func(t *testing.T) {
		str := "hello world"
		encoded := EncodeString(-1, str)
		decoded := DecodeString(-1, encoded)
		r.Equal(str, decoded)
	})

	t.Run("uint256", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			u := big.NewInt(1234567890)
			encoded := EncodeUint256(32, u)
			decoded := DecodeUint256(32, encoded)
			r.Equal(u.Int64(), decoded.Int64())
		}
	})

	t.Run("int256", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			u := big.NewInt(-1234567890)
			encoded := EncodeInt256(32, u)
			decoded := DecodeInt256(32, encoded)
			r.Equal(u.Int64(), decoded.Int64())
		}
	})

	t.Run("uint8", func(t *testing.T) {
		u := uint8(123)
		encoded := EncodeSmallUint8(1, u)
		decoded := DecodeSmallUint8(1, encoded)
		r.Equal(u, decoded)
	})

	t.Run("uint16", func(t *testing.T) {
		u := uint16(123)
		encoded := EncodeSmallUint16(2, u)
		decoded := DecodeSmallUint16(2, encoded)
		r.Equal(u, decoded)
	})

	t.Run("uint32", func(t *testing.T) {
		u := uint32(123)
		encoded := EncodeSmallUint32(4, u)
		decoded := DecodeSmallUint32(4, encoded)
		r.Equal(u, decoded)
	})

	t.Run("uint64", func(t *testing.T) {
		u := uint64(123)
		encoded := EncodeSmallUint64(8, u)
		decoded := DecodeSmallUint64(8, encoded)
		r.Equal(u, decoded)
	})

	t.Run("int8", func(t *testing.T) {
		u := int8(123)
		encoded := EncodeSmallInt8(1, u)
		decoded := DecodeSmallInt8(1, encoded)
		r.Equal(u, decoded)
	})

	t.Run("int16", func(t *testing.T) {
		u := int16(123)
		encoded := EncodeSmallInt16(2, u)
		decoded := DecodeSmallInt16(2, encoded)
		r.Equal(u, decoded)
	})

	t.Run("int32", func(t *testing.T) {
		u := int32(123)
		encoded := EncodeSmallInt32(4, u)
		decoded := DecodeSmallInt32(4, encoded)
		r.Equal(u, decoded)
	})

	t.Run("int64", func(t *testing.T) {
		u := int64(123)
		encoded := EncodeSmallInt64(8, u)
		decoded := DecodeSmallInt64(8, encoded)
		r.Equal(u, decoded)
	})
}
