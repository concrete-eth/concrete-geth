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
	"bytes"
	"errors"
	"math"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

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

func TestUint64ToBytes(t *testing.T) {
	for _, test := range uint64AndBytesTestCases {
		require.Equal(t, test.bytes, Uint64ToBytes(test.num))
	}
}

func TestBytesToUint64(t *testing.T) {
	for _, test := range uint64AndBytesTestCases {
		require.Equal(t, test.num, BytesToUint64(test.bytes))
	}
	// require.Equal(t, uint64(0), BytesToUint64([]byte{}))
	// require.Equal(t, uint64(0), nil)
}

var errorEncodeDecodeTestCases = []struct {
	err    error
	errEnc []byte
}{
	{nil, []byte{nil_error}},
	{errors.New("test error"), []byte{notNil_error, 't', 'e', 's', 't', ' ', 'e', 'r', 'r', 'o', 'r'}},
}

func TestEncodeError(t *testing.T) {
	for _, test := range errorEncodeDecodeTestCases {
		require.Equal(t, test.errEnc, EncodeError(test.err))
	}
}

func TestDecodeError(t *testing.T) {
	for _, test := range errorEncodeDecodeTestCases {
		require.Equal(t, test.err, DecodeError(test.errEnc))
	}
	require.Equal(t, nil, DecodeError([]byte{}))
	require.Equal(t, nil, DecodeError(nil))
}

func TestGetData(t *testing.T) {
	data := []byte("testdata")
	start := uint64(2)
	size := uint64(4)
	expected := common.RightPadBytes([]byte("stda"), int(size))
	result := GetData(data, start, size)
	if !bytes.Equal(result, expected) {
		t.Errorf("GetData failed, expected %v, got %v", string(expected), string(result))
	}
}

func TestSplitData(t *testing.T) {
	data := []byte("testdata")
	size := uint64(4)
	part1, part2 := SplitData(data, size)
	if !bytes.Equal(part1, data[:size]) || !bytes.Equal(part2, data[size:]) {
		t.Errorf("SplitData failed, expected %v and %v, got %v and %v", string(data[:size]), string(data[size:]), string(part1), string(part2))
	}
}

func TestSplitInput(t *testing.T) {
	input := []byte("testdata")
	size := uint64(4)
	part1, part2 := SplitInput(input)
	if !bytes.Equal(part1, input[:size]) || !bytes.Equal(part2, input[size:]) {
		t.Errorf("SplitInput failed, expected %v and %v, got %v and %v", string(input[:size]), string(input[size:]), string(part1), string(part2))
	}
}
