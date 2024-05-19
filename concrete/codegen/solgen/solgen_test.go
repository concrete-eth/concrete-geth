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

package solgen

import (
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestValidContractName(t *testing.T) {
	testCases := []struct {
		name     string
		expected bool
	}{
		{"", false},
		{"0", false},
		{"_", false},
		{"a a", false},
		{"a", true},
		{"a0", true},
		{"a_", true},
		{"a0_a", true},
	}
	for _, testCase := range testCases {
		if isValidSolidityContractName(testCase.name) != testCase.expected {
			t.Errorf("unexpected result for %q", testCase.name)
		}
	}
}

func TestSolgen(t *testing.T) {
	config := Config{
		Name:       "CounterPrecompile",
		Address:    common.Address{0x80},
		Pragma:     "^0.8.0",
		AbiPath:    filepath.Join("testdata", "Counter.abi.json"),
		OutPath:    filepath.Join("testdata", "CounterPrecompile.sol"),
		ImportPath: filepath.Join("testdata", "Dependency.sol"),
	}
	if err := GenerateSolidityLibrary(config); err != nil {
		t.Fatal(err)
	}
}
