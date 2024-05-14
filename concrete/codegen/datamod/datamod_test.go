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

package datamod

import (
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod/testdata"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/concrete/mock"
	"github.com/stretchr/testify/require"
)

func TestBadDatamod(t *testing.T) {
	dirPath := "./testdata/bad-datamods/"
	files, err := os.ReadDir(dirPath)
	if err != nil {
		t.Fatalf("Failed to read directory: %s", err)
	}
	for _, file := range files {
		t.Run(file.Name(), func(t *testing.T) {
			err := GenerateDataModel(Config{
				JSON:    dirPath + file.Name(),
				Out:     "./",
				Package: "test",
			}, true)
			if err == nil {
				t.Fatalf("Expected error but got nil")
			}
			if !strings.Contains(err.Error(), "schema for table") {
				t.Fatalf("Unexpected error: %s", err)
			}
		})
	}
}

type testRowInterface interface {
	Get() (*big.Int, *big.Int, string, []byte, bool, common.Address, []byte)
	Set(*big.Int, *big.Int, string, []byte, bool, common.Address, []byte)

	SetValueUint(*big.Int)
	SetValueInt(*big.Int)
	SetValueString(string)
	SetValueBytes([]byte)
	SetValueBool(bool)
	SetValueAddress(common.Address)
	SetValueBytes16([]byte)
}

var (
	uintVal    = big.NewInt(1)
	intVal     = new(big.Int).Neg(big.NewInt(1))
	stringVal  = "string"
	bytesVal   = []byte("bytes")
	boolVal    = true
	addrVal    = common.HexToAddress("1234567890123456789012345678901234567890")
	bytes16Val = common.Hex2Bytes("12345678901234567890123456789012")
)

func testRow(t *testing.T, getRow func() testRowInterface) {
	r := require.New(t)
	row := getRow()

	r.NotNil(row)

	uintValCur, intValCur, stringValCur, bytesValCur, boolValCur, addrValCur, bytes16ValCur := row.Get()
	r.Equal(int64(0), uintValCur.Int64())
	r.Equal(int64(0), intValCur.Int64())
	r.Equal("", stringValCur)
	r.Equal([]byte{}, bytesValCur)
	r.Equal(false, boolValCur)
	r.Equal(common.Address{}, addrValCur)
	r.Equal(make([]byte, 16), bytes16ValCur)

	row.SetValueUint(uintVal)
	row.SetValueInt(intVal)
	row.SetValueString(stringVal)
	row.SetValueBytes(bytesVal)
	row.SetValueBool(boolVal)
	row.SetValueAddress(addrVal)
	row.SetValueBytes16(bytes16Val)

	newRowInstance := getRow()

	for _, rr := range []testRowInterface{row, newRowInstance} {
		uintValCur, intValCur, stringValCur, bytesValCur, boolValCur, addrValCur, bytes16ValCur = rr.Get()
		r.Equal(uintVal.Int64(), uintValCur.Int64())
		r.Equal(intVal.Int64(), intValCur.Int64())
		r.Equal(stringVal, stringValCur)
		r.Equal(bytesVal, bytesValCur)
		r.Equal(boolVal, boolValCur)
		r.Equal(addrVal, addrValCur)
		r.Equal(bytes16Val, bytes16ValCur)
	}
}

func TestTables(t *testing.T) {
	var (
		addr     = common.HexToAddress("0x1234567890123456789012345678901234567890")
		config   = api.EnvConfig{}
		meterGas = false
		gas      = uint64(0)
		env      = mock.NewMockEnvironment(addr, config, meterGas, gas)
		ds       = lib.NewDatastore(env)
	)

	t.Run("KeyedTable", func(t *testing.T) {
		table := testdata.NewKeyedTable(ds)
		testRow(t, func() testRowInterface {
			return table.Get(uintVal, intVal, stringVal, bytesVal, boolVal, addrVal, bytes16Val)
		})
	})

	t.Run("KeylessTable", func(t *testing.T) {
		testRow(t, func() testRowInterface {
			return testdata.NewKeylessTable(ds)
		})
	})

	t.Run("KeyedWithKeyedTable", func(t *testing.T) {
		table := testdata.NewKeyedWithKeyedTableValue(ds)
		row := table.Get(uintVal)
		subTable := row.GetValueTable()
		testRow(t, func() testRowInterface {
			return subTable.Get(uintVal, intVal, stringVal, bytesVal, boolVal, addrVal, bytes16Val)
		})
	})

	t.Run("keyedWithKeylessTable", func(t *testing.T) {
		table := testdata.NewKeyedWithKeylessTableValue(ds)
		row := table.Get(uintVal)
		testRow(t, func() testRowInterface {
			return row.GetValueTable()
		})
	})

	t.Run("keylessWithKeyedTable", func(t *testing.T) {
		row := testdata.NewKeylessWithKeyedTableValue(ds)
		subTable := row.GetValueTable()
		testRow(t, func() testRowInterface {
			return subTable.Get(uintVal, intVal, stringVal, bytesVal, boolVal, addrVal, bytes16Val)
		})
	})

	t.Run("keylessWithKeylessTable", func(t *testing.T) {
		row := testdata.NewKeylessWithKeylessTableValue(ds)
		testRow(t, func() testRowInterface {
			return row.GetValueTable()
		})
	})
}
