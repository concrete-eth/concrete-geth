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

package contracts

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

func TestAddPrecompile(t *testing.T) {
	pcs := ActivePrecompiles()
	if len(pcs) != 0 {
		t.Errorf("expected no precompiles")
	}

	for i := byte(0); i < 10; i++ {
		addr := common.BytesToAddress([]byte{i})
		err := AddPrecompile(addr, &lib.Blank{})
		if err != nil {
			t.Error(err)
		}
		_, ok := GetPrecompile(addr)
		if !ok {
			t.Errorf("expected precompile at address %x", addr)
		}
		pcAddr := ActivePrecompiles()[i]
		if pcAddr != addr {
			t.Errorf("expected precompile at address %x, got %x", addr, pcAddr)
		}
	}

	pcs = ActivePrecompiles()
	if len(pcs) != 10 {
		t.Errorf("expected 10 precompiles")
	}
}

func TestRunPrecompile(t *testing.T) {}
