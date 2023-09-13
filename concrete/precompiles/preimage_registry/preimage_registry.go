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
	_ "embed"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/concrete/precompiles"
	"github.com/ethereum/go-ethereum/concrete/utils"
)

//go:embed sol/abi/PreimageRegistry.abi
var abiFile string

var ABI abi.ABI

var PreimageRegistryMetadata = precompiles.PrecompileMetadata{
	Name:        "PreimageRegistry",
	Version:     precompiles.Version{common.Big0, common.Big1, common.Big0},
	Author:      "The concrete-geth Authors",
	Description: "A registry of stored preimages indexed by their hash.",
	Source:      "https://github.com/therealbytes/concrete-geth/tree/concrete/concrete/precompiles/preimage_registry.go",
	ABI:         abiFile,
}

var (
	// crypto.Keccak256Hash(nil)
	EmptyPreimageHash = common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
)

func init() {
	abiReader := strings.NewReader(abiFile)
	var err error
	ABI, err = abi.JSON(abiReader)
	if err != nil {
		panic(err)
	}
}

type PreimageRegistry struct {
	lib.BlankPrecompile
}

func (p *PreimageRegistry) IsStatic(input []byte) bool {
	methodID, _ := utils.SplitInput(input)
	method, err := ABI.MethodById(methodID)
	if err != nil {
		return true
	}
	return method.IsConstant()
}

func (p *PreimageRegistry) Run(env api.Environment, input []byte) ([]byte, error) {
	methodID, data := utils.SplitInput(input)
	method, err := ABI.MethodById(methodID)
	if err != nil {
		return nil, precompiles.ErrMethodNotFound
	}
	args, err := method.Inputs.Unpack(data)
	if err != nil {
		return nil, precompiles.ErrInvalidInput
	}
	var result interface{}

	preimageMap := lib.NewDatastore(env).Get([]byte("map.size.v1")).Mapping()

	switch method.Name {

	case "addPreimage":
		preimage := args[0].([]byte)
		length := len(preimage)
		if length == 0 {
			result = EmptyPreimageHash
		} else {
			hash := env.PersistentPreimageStore_Unsafe(preimage)
			preimageMap.Get(hash.Bytes()).SetUint64(uint64(length))
			result = hash
		}

	case "hasPreimage":
		hash := common.Hash(args[0].([32]byte))
		if hash == EmptyPreimageHash {
			result = true
		} else {
			result = preimageMap.Get(hash.Bytes()).Uint64() > 0
		}

	case "getPreimageSize":
		hash := common.Hash(args[0].([32]byte))
		if hash == EmptyPreimageHash {
			result = big.NewInt(0)
		} else {
			result = preimageMap.Get(hash.Bytes()).BigUint()
		}

	case "getPreimage":
		hash := common.Hash(args[0].([32]byte))
		if hash == EmptyPreimageHash {
			result = []byte{}
		} else {
			size := preimageMap.Get(hash.Bytes()).Uint64()
			if size == 0 {
				result = []byte{}
			}
			result = env.PersistentPreimageLoad_Unsafe(hash)
		}

	default:
		return nil, precompiles.ErrMethodNotFound
	}

	output, err := method.Outputs.Pack(result)
	if err != nil {
		// Panic because this is a bug in the precompile.
		panic(err)
	}
	return output, nil
}

var _ precompiles.Precompile = (*PreimageRegistry)(nil)
