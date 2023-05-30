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

package test

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	cc_api "github.com/ethereum/go-ethereum/concrete/api"
	api_test "github.com/ethereum/go-ethereum/concrete/api/test"
	"github.com/ethereum/go-ethereum/concrete/crypto"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/stretchr/testify/require"
)

func TestCounter(t *testing.T) {
	var (
		r       = require.New(t)
		sdb     = api_test.NewMockStateDB()
		evm     = api_test.NewMockEVM(sdb)
		api     = cc_api.New(evm, common.Address{})
		key     = lib.NewKey("test.counter.v0")
		ref     = api.Persistent().NewReference(key)
		counter = lib.NewCounter(ref)
	)

	// Check that the counter is initially zero
	r.Equal(int64(0), counter.Get().Int64())

	// Increment the counter
	for ii := 1; ii <= 10; ii++ {
		counter.Inc()
		r.Equal(int64(ii), counter.Get().Int64())
	}

	// Decrement the counter
	for ii := 9; ii >= 0; ii-- {
		counter.Dec()
		r.Equal(int64(ii), counter.Get().Int64())
	}

	// Add to the counter
	counter.Add(common.Big2)
	r.Equal(int64(2), counter.Get().Int64())

	// Subtract from the counter
	counter.Sub(common.Big2)
	r.Equal(int64(0), counter.Get().Int64())

	// Check the reference matches the counter
	value := big.NewInt(1234)
	r.Equal(ref.Get(), common.BigToHash(counter.Get()))
	ref.Set(common.BigToHash(value))
	r.Equal(ref.Get(), common.BigToHash(counter.Get()))
	r.Equal(value, counter.Get())
}

func TestBigPreimageStore(t *testing.T) {
	var (
		r   = require.New(t)
		sdb = api_test.NewMockStateDB()
		evm = api_test.NewMockEVM(sdb)
		api = cc_api.New(evm, common.Address{})
	)
	var (
		radixCases    = []int{2, 4, 8, 16}
		leafSizeCases = []int{32, 128, 512}
		pi0           = []byte("hello world")
		pi1           = crypto.Keccak256(pi0)
		pi2           = bytes.Repeat(pi1, 100)
		preimageCases = [][]byte{pi0, pi1, pi2}
	)
	for _, radix := range radixCases {
		for _, leafSize := range leafSizeCases {
			for i, preimage := range preimageCases {
				t.Run(fmt.Sprint("r", radix, "/l", leafSize, "/pi", i), func(t *testing.T) {
					store := lib.NewBigPreimageStore(api.Persistent(), radix, leafSize)
					root := store.Add(preimage)
					retrivedPreimage := store.Get(root)
					r.True(store.Has(root))
					r.Equal(len(preimage), len(retrivedPreimage))
					r.Equal(preimage, retrivedPreimage)
				})
			}
		}
	}
}
