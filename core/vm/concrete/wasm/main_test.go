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
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm/concrete/api"
	"github.com/ethereum/go-ethereum/core/vm/concrete/lib"
	"github.com/ethereum/go-ethereum/core/vm/concrete/wasm/bridge"
)

func newTestConcreteAPI() (api.API, api.EVM, api.StateDB) {
	stateDB := lib.NewMockStateDB()
	evm := lib.NewMockEVM(stateDB)
	api := api.New(evm, common.Address{})
	return api, evm, stateDB
}

func TestWasmPrecompileBlank(t *testing.T) {
	api, _, _ := newTestConcreteAPI()
	p := Precompiles[common.BytesToAddress([]byte{129})]
	if p == nil {
		t.Fatal("Precompile not found")
	}
	gas := p.RequiredGas([]byte{})
	fmt.Println("Gas:", gas)
	starTime := time.Now()
	result, err := p.Run(api, bridge.Uint64ToBytes(1))
	fmt.Println("Duration:", time.Since(starTime))
	fmt.Println("Result:", result)
	// fmt.Println("Result (uint64):", bridge.BytesToUint64(result))
	fmt.Println("Error:", err)
}
