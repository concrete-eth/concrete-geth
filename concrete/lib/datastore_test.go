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

package lib

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/testutils"
	"github.com/stretchr/testify/require"
)

func testKeyValueStore(t *testing.T, kv KeyValueStore) {
	r := require.New(t)
	key := common.Hash{0x01}
	value := kv.Get(key)
	r.Equal(common.Hash{}, value)
}

func TestEnvPersistentKeyValueStore(t *testing.T) {
	var (
		address = common.HexToAddress("0xc0ffee0001")
		config  = api.EnvConfig{
			Static:    false,
			Ephemeral: false,
			Preimages: false,
			Trusted:   false,
		}
		meterGas = true
		gas      = uint64(1e6)
	)
	env := testutils.NewMockEnv(address, config, meterGas, gas)
	kv := newEnvPersistentKeyValueStore(env)
	testKeyValueStore(t, kv)
}

func TestEnvEphemeralKeyValueStore(t *testing.T) {
	var (
		address = common.HexToAddress("0xc0ffee0001")
		config  = api.EnvConfig{
			Static:    false,
			Ephemeral: false,
			Preimages: false,
			Trusted:   false,
		}
		meterGas = true
		gas      = uint64(1e6)
	)
	env := testutils.NewMockEnv(address, config, meterGas, gas)
	kv := newEnvEphemeralKeyValueStore(env)
	testKeyValueStore(t, kv)
}
