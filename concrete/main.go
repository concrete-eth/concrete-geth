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

package concrete

import (
	"errors"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/cmd/geth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/precompiles"
	"github.com/urfave/cli/v2"
)

type ConcreteApp interface {
	RunWithArgs(args []string) error
	RunWithOsArgs() error
	Run() error
	AddPrecompileWithMetadata(addr common.Address, pc precompiles.Precompile, metadata precompiles.PrecompileMetadata)
	AddPrecompile(addr common.Address, pc precompiles.Precompile)
}

type concreteGeth struct {
	app *cli.App
}

var ConcreteGeth ConcreteApp = &concreteGeth{
	app: geth.App,
}

func (a *concreteGeth) RunWithArgs(arguments []string) error {
	return a.app.Run(arguments)
}

func (a *concreteGeth) RunWithOsArgs() error {
	return a.RunWithArgs(os.Args)
}

func (a *concreteGeth) Run() error {
	return a.RunWithOsArgs()
}

func (a *concreteGeth) validateNewPCAddress(addr common.Address) error {
	if addr.Big().Cmp(big.NewInt(128)) < 0 {
		return errors.New("precompile address cannot be below 0x80")
	}
	return nil
}

func (a *concreteGeth) AddPrecompileWithMetadata(addr common.Address, pc precompiles.Precompile, metadata precompiles.PrecompileMetadata) {
	var err error
	if err = a.validateNewPCAddress(addr); err == nil {
		err = precompiles.AddPrecompileWithMetadata(addr, pc, metadata)
	}
	if err != nil {
		panic(err)
	}
}

func (a *concreteGeth) AddPrecompile(addr common.Address, pc precompiles.Precompile) {
	a.AddPrecompileWithMetadata(addr, pc, precompiles.PrecompileMetadata{})
}
