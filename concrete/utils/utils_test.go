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

//go:build !tinygo

// This file will ignored when building with tinygo to prevent compatibility
// issues.

package utils

import (
	"errors"
	"math"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestUint64BytesConversion(t *testing.T) {
	var uint64AndBytesTestCases = []struct {
		num   uint64
		bytes []byte
	}{
		{0, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{1, []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{256, []byte{0, 0, 0, 0, 0, 0, 1, 0}},
		{math.MaxUint8, []byte{0, 0, 0, 0, 0, 0, 0, 255}},
		{math.MaxUint16, []byte{0, 0, 0, 0, 0, 0, 255, 255}},
		{math.MaxUint32, []byte{0, 0, 0, 0, 255, 255, 255, 255}},
		{math.MaxUint64 / 2, []byte{127, 255, 255, 255, 255, 255, 255, 255}},
		{math.MaxUint64, []byte{255, 255, 255, 255, 255, 255, 255, 255}},
	}

	t.Run("Uint64ToBytes", func(t *testing.T) {
		for _, test := range uint64AndBytesTestCases {
			require.Equal(t, test.bytes, Uint64ToBytes(test.num))
		}
	})

	t.Run("BytesToUint64", func(t *testing.T) {
		for _, test := range uint64AndBytesTestCases {
			require.Equal(t, test.num, BytesToUint64(test.bytes))
		}
	})
}

func TestErrorCodec(t *testing.T) {
	var errorEncodeDecodeTestCases = []struct {
		err    error
		errEnc []byte
	}{
		{nil, []byte{nil_error}},
		{errors.New("test error"), []byte{notNil_error, 't', 'e', 's', 't', ' ', 'e', 'r', 'r', 'o', 'r'}},
	}

	t.Run("EncodeError", func(t *testing.T) {
		for _, test := range errorEncodeDecodeTestCases {
			require.Equal(t, test.errEnc, EncodeError(test.err))
		}
	})

	t.Run("DecodeError", func(t *testing.T) {
		for _, test := range errorEncodeDecodeTestCases {
			require.Equal(t, test.err, DecodeError(test.errEnc))
		}
		// Additional cases specifically for DecodeError
		require.Equal(t, nil, DecodeError([]byte{}))
		require.Equal(t, nil, DecodeError(nil))
	})
}

func TestDataUtils(t *testing.T) {
	t.Run("GetData", func(t *testing.T) {
		data := []byte("testdata")
		start := uint64(2)
		size := uint64(4)
		expected := common.RightPadBytes([]byte("stda"), int(size))
		result := GetData(data, start, size)
		require.Equal(t, expected, result)
	})

	t.Run("SplitData", func(t *testing.T) {
		data := []byte("testdata")
		size := uint64(4)
		part1, part2 := SplitData(data, size)
		require.Equal(t, part1, data[:size])
		require.Equal(t, part2, data[size:])
	})

	t.Run("SplitInput", func(t *testing.T) {
		input := []byte("testdata")
		size := uint64(4)
		part1, part2 := SplitInput(input)
		require.Equal(t, part1, input[:size])
		require.Equal(t, part2, input[size:])
	})
}
