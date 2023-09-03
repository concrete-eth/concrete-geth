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

package precompile_registry

import (
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/concrete/precompiles"
	"github.com/ethereum/go-ethereum/concrete/utils"
)

//go:embed sol/abi/PrecompileRegistry.abi
var abiJson string

var ABI abi.ABI

var PrecompileRegistryMetadata = precompiles.PrecompileMetadata{
	Name:        "PrecompileRegistry",
	Version:     precompiles.Version{common.Big0, common.Big1, common.Big0},
	Author:      "The concrete-geth Authors",
	Description: "A registry of precompiles indexed by address and name.",
	Source:      "https://github.com/therealbytes/concrete-geth/tree/concrete/concrete/precompiles/precompile_registry.go",
	ABI:         abiJson,
}

type FrameworkMetadata = struct {
	Name    string              `json:"name"`
	Version precompiles.Version `json:"version"`
	Source  string              `json:"source"`
}

var frameworkMetadata = FrameworkMetadata{
	Name:    "Concrete",
	Version: precompiles.Version{common.Big0, common.Big1, common.Big0},
	Source:  "https://github.com/therealbytes/concrete-geth",
}

func init() {
	abiReader := strings.NewReader(abiJson)
	var err error
	ABI, err = abi.JSON(abiReader)
	if err != nil {
		panic(err)
	}
}

type PrecompileRegistry struct {
	lib.BlankPrecompile
}

func (p *PrecompileRegistry) IsStatic(input []byte) bool {
	methodID, _ := utils.SplitInput(input)
	method, err := ABI.MethodById(methodID)
	if err != nil {
		return true
	}
	return method.IsConstant()
}

func (p *PrecompileRegistry) Run(env api.Environment, input []byte) ([]byte, error) {
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

	switch method.Name {

	case "getFramework":
		return method.Outputs.Pack(frameworkMetadata)

	case "getPrecompile":
		address := common.Address(args[0].(common.Address))
		metadata := precompiles.GetPrecompileMetadataByAddress(address)
		if metadata == nil {
			metadata = &precompiles.PrecompileMetadata{}
		}
		result = *metadata

	case "getPrecompileByName":
		name := args[0].(string)
		metadata := precompiles.GetPrecompileMetadataByName(name)
		if metadata == nil {
			metadata = &precompiles.PrecompileMetadata{}
		}
		result = metadata.Address

	case "getPrecompiledAddresses":
		result = precompiles.ActivePrecompiles()

	case "getPrecompiles":
		addresses := precompiles.ActivePrecompiles()
		metadata := make([]precompiles.PrecompileMetadata, len(addresses))
		for ii, address := range addresses {
			metadata[ii] = *precompiles.GetPrecompileMetadataByAddress(address)
		}
		result = metadata

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

var _ precompiles.Precompile = (*PrecompileRegistry)(nil)
