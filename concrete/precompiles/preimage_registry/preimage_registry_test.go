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

package preimage_registry

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/crypto"
	"github.com/ethereum/go-ethereum/concrete/mock"
	"github.com/stretchr/testify/require"
)

func TestPreimageRegistry(t *testing.T) {
	var (
		r        = require.New(t)
		pc       = PreimageRegistry{}
		address  = common.HexToAddress("0xc0ffee0001")
		config   = api.EnvConfig{Trusted: true, Preimages: true}
		meterGas = false
		gas      = uint64(0)
	)

	r.Equal(false, pc.IsStatic(ABI.Methods["addPreimage"].ID))
	r.Equal(true, pc.IsStatic(ABI.Methods["hasPreimage"].ID))
	r.Equal(true, pc.IsStatic(ABI.Methods["getPreimageSize"].ID))
	r.Equal(true, pc.IsStatic(ABI.Methods["getPreimage"].ID))

	env := mock.NewMockEnvironment(address, config, meterGas, gas)

	addPreimage := func(preimage []byte) common.Hash {
		input, err := ABI.Pack("addPreimage", preimage)
		r.NoError(err)
		output, err := pc.Run(env, input)
		r.NoError(err)
		outputs, err := ABI.Methods["addPreimage"].Outputs.Unpack(output)
		r.NoError(err)
		return common.Hash(outputs[0].([32]byte))
	}

	hasPreimage := func(hash common.Hash) bool {
		input, err := ABI.Pack("hasPreimage", hash)
		r.NoError(err)
		output, err := pc.Run(env, input)
		r.NoError(err)
		outputs, err := ABI.Methods["hasPreimage"].Outputs.Unpack(output)
		r.NoError(err)
		return outputs[0].(bool)
	}

	getPreimageSize := func(hash common.Hash) uint64 {
		input, err := ABI.Pack("getPreimageSize", hash)
		r.NoError(err)
		output, err := pc.Run(env, input)
		r.NoError(err)
		outputs, err := ABI.Methods["getPreimageSize"].Outputs.Unpack(output)
		r.NoError(err)
		return (outputs[0].(*big.Int)).Uint64()
	}

	getPreimage := func(hash common.Hash) []byte {
		input, err := ABI.Pack("getPreimage", hash)
		r.NoError(err)
		output, err := pc.Run(env, input)
		r.NoError(err)
		outputs, err := ABI.Methods["getPreimage"].Outputs.Unpack(output)
		r.NoError(err)
		return outputs[0].([]byte)
	}

	r.Equal(true, hasPreimage(EmptyPreimageHash))
	r.Equal(uint64(0), getPreimageSize(EmptyPreimageHash))
	r.Equal([]byte{}, getPreimage(EmptyPreimageHash))
	r.Equal(EmptyPreimageHash, addPreimage([]byte{}))

	preimage := []byte{0x01}
	hash := crypto.Keccak256Hash(preimage)

	r.Equal(false, hasPreimage(hash))
	r.Equal(uint64(0), getPreimageSize(hash))
	r.Equal([]byte{}, getPreimage(hash))
	r.Equal(hash, addPreimage(preimage))

	r.Equal(true, hasPreimage(hash))
	r.Equal(uint64(len(preimage)), getPreimageSize(hash))
	r.Equal(preimage, getPreimage(hash))
}
