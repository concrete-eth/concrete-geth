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

package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var opCodeEncodeDecodeTestCases = []struct {
	code    OpCode
	codeEnc []byte
}{
	{code: OpCode(0), codeEnc: []byte{0x00}},
	{code: OpCode(1), codeEnc: []byte{0x01}},
	{code: OpCode(2), codeEnc: []byte{0x02}},
}

func TestEncodeOpCode(t *testing.T) {
	for _, test := range opCodeEncodeDecodeTestCases {
		require.Equal(t, test.codeEnc, test.code.Encode())
	}
}

func TestDecodeOpCode(t *testing.T) {
	for _, test := range opCodeEncodeDecodeTestCases {
		var codeDec OpCode
		codeDec.Decode(test.codeEnc)
		require.Equal(t, test.code, codeDec)
	}
}
