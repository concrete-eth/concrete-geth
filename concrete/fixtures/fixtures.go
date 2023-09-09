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

package fixtures

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
	fixture_datamod "github.com/ethereum/go-ethereum/concrete/fixtures/datamod"
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

const KkvAbiString = "[{\"inputs\":[{\"name\":\"k1\",\"type\":\"bytes32\"},{\"name\":\"k2\",\"type\":\"bytes32\"},{\"name\":\"v\",\"type\":\"bytes32\"}],\"name\":\"set\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"k1\",\"type\":\"bytes32\"},{\"name\":\"k2\",\"type\":\"bytes32\"}],\"name\":\"get\",\"outputs\":[{\"name\":\"v\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

var (
	kkvSetMethodID = common.Hex2Bytes("bb40a4a9")
	kkvGetMethodID = common.Hex2Bytes("658cc1f6")
)

type KeyKeyValuePrecompile struct {
	lib.BlankPrecompile
}

func (a *KeyKeyValuePrecompile) IsStatic(input []byte) bool {
	methodID, _ := utils.SplitInput(input)
	if bytes.Equal(methodID, kkvGetMethodID) {
		return true
	} else if bytes.Equal(methodID, kkvSetMethodID) {
		return false
	}
	return true
}

func (a *KeyKeyValuePrecompile) Run(env api.Environment, input []byte) ([]byte, error) {
	methodID, data := utils.SplitInput(input)
	if bytes.Equal(methodID, kkvGetMethodID) {
		if len(data) != 64 {
			return nil, precompiles.ErrInvalidInput
		}
		k1 := common.BytesToHash(data[:32])
		k2 := common.BytesToHash(data[32:])
		kkv := fixture_datamod.NewKkv(lib.NewDatastore(env))
		v := kkv.Get(k1, k2).GetValue()
		return v.Bytes(), nil
	} else if bytes.Equal(methodID, kkvSetMethodID) {
		if len(data) != 96 {
			return nil, precompiles.ErrInvalidInput
		}
		k1 := common.BytesToHash(data[:32])
		k2 := common.BytesToHash(data[32:64])
		v := common.BytesToHash(data[64:])
		kkv := fixture_datamod.NewKkv(lib.NewDatastore(env))
		kkv.Get(k1, k2).SetValue(v)
		return nil, nil
	}
	return nil, precompiles.ErrMethodNotFound
}
