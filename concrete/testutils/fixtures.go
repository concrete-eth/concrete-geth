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

package testutils

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/concrete/precompiles"
	"github.com/ethereum/go-ethereum/concrete/utils"
)

const AddAbiString = "[{\"inputs\":[{\"name\":\"x\",\"type\":\"uint256\"},{\"name\":\"y\",\"type\":\"uint256\"}],\"name\":\"add\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]"

var addMethodID = common.Hex2Bytes("771602f7")

type AdditionPrecompile struct {
	lib.BlankPrecompile
}

func (a *AdditionPrecompile) Run(env api.Environment, input []byte) ([]byte, error) {
	methodID, data := utils.SplitInput(input)
	if !bytes.Equal(methodID, addMethodID) {
		return nil, precompiles.ErrMethodNotFound
	}
	if len(data) != 64 {
		return nil, precompiles.ErrInvalidInput
	}
	x := new(big.Int).SetBytes(data[:32])
	y := new(big.Int).SetBytes(data[32:])
	z := new(big.Int).Add(x, y)
	return common.BigToHash(z).Bytes(), nil
}

var _ precompiles.Precompile = &AdditionPrecompile{}
