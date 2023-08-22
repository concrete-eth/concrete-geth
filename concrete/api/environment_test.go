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

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestGas(t *testing.T) {
	var (
		r       = require.New(t)
		address = common.HexToAddress("0xc0ffee0001")
		config  = EnvConfig{
			Static:    false,
			Ephemeral: true,
			Preimages: true,
			Trusted:   true,
		}
		meterGas = true
		gas      = uint64(1e6)
	)

	env := newMockEnv(address, config, meterGas, gas)

	r.Equal(gas, env.Gas())

	gasOpCost := env.table[GetGasLeft_OpCode].constantGas

	r.Equal(gas-gasOpCost, env.GetGasLeft())
}
