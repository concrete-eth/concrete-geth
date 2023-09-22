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

package wasm

import (
	_ "embed"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/mock"
)

//go:embed testdata/gas.wasm
var errCode []byte

func TestWasmEnvErr(t *testing.T) {
	input := []byte{1}
	runtimes := []struct {
		name string
		pc   concrete.Precompile
	}{
		{"wazero", NewWazeroPrecompile(errCode)},
		{"wasmer", NewWasmerPrecompile(errCode)},
	}
	for _, runtime := range runtimes {
		t.Run(runtime.name, func(t *testing.T) {
			env := mock.NewMockEnvironment(common.Address{}, api.EnvConfig{Trusted: true}, true, 0)
			output, _, err := concrete.RunPrecompile(runtime.pc, env, input, true)
			if err == nil {
				t.Fatal("expected error")
			}
			if err != api.ErrOutOfGas {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(output) != 0 {
				t.Fatalf("unexpected output: %x", output)
			}
		})
	}
}
