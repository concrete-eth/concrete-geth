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

package testing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

func runTest(bytecode []byte, method abi.Method, shouldFail bool) (uint64, error) {
	var (
		key, _          = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		senderAddress   = crypto.PubkeyToAddress(key.PublicKey)
		contractAddress = common.HexToAddress("0x0000000000000000000000000000000000c0ffee")
		gspec           = &core.Genesis{
			Config: params.TestChainConfig,
			Alloc: core.GenesisAlloc{
				senderAddress:   {Balance: big.NewInt(1e18)},
				contractAddress: {Balance: common.Big0, Code: bytecode},
			},
		}
		signer   = types.LatestSigner(gspec.Config)
		gasLimit = uint64(1e6)
		setupId  = crypto.Keccak256([]byte("setUp()"))[:4]
	)

	_, _, receipts := core.GenerateChainWithGenesis(gspec, ethash.NewFaker(), 1, func(ii int, block *core.BlockGen) {
		for _, id := range [][]byte{setupId, method.ID} {
			tx := types.NewTransaction(block.TxNonce(senderAddress), contractAddress, common.Big0, gasLimit, block.BaseFee(), id)
			signed, err := types.SignTx(tx, signer, key)
			if err != nil {
				panic(err)
			}
			block.AddTx(signed)
		}
	})

	if len(receipts[0]) != 2 {
		return 0, fmt.Errorf("expected 2 receipts, got %d", len(receipts))
	}
	setupReceipt := receipts[0][0]
	testReceipt := receipts[0][1]
	if setupReceipt.Status != types.ReceiptStatusSuccessful {
		return 0, fmt.Errorf("setup failed")
	}
	if (testReceipt.Status == types.ReceiptStatusSuccessful) == shouldFail {
		return 0, fmt.Errorf("test failed")
	}

	return testReceipt.GasUsed, nil
}

func runTests(bytecode []byte, ABI abi.ABI) (int, int) {
	passed := 0
	failed := 0
	for _, method := range ABI.Methods {
		if !strings.HasPrefix(method.Name, "test") {
			continue
		}
		shouldFail := strings.HasPrefix(method.Name, "testFail")
		gas, err := runTest(bytecode, method, shouldFail)
		if err == nil {
			passed++
			fmt.Printf("[PASS] %s() (gas: %d)\n", method.Name, gas)
		} else {
			failed++
			fmt.Printf("[FAIL] %s() (gas: %d): %s\n", method.Name, gas, err)
		}
	}
	return passed, failed
}

func extractTestData(contractJsonBytes []byte) ([]byte, abi.ABI, string, error) {
	var jsonData struct {
		ABI              abi.ABI `json:"abi"`
		DeployedBytecode struct {
			Object string `json:"object"`
		} `json:"deployedBytecode"`
		Ast struct {
			AbsolutePath string `json:"absolutePath"`
		} `json:"ast"`
	}
	err := json.Unmarshal(contractJsonBytes, &jsonData)
	if err != nil {
		return nil, abi.ABI{}, "", err
	}
	bytecode := common.FromHex(jsonData.DeployedBytecode.Object)
	return bytecode, jsonData.ABI, jsonData.Ast.AbsolutePath, nil
}

func getFileNames(dir string, ext string) ([]string, error) {
	var fileNames []string

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fileNames, err
	}

	for _, file := range files {
		filePath := filepath.Join(dir, file.Name())
		if file.IsDir() {
			subDirFileNames, err := getFileNames(filePath, ext)
			if err != nil {
				return fileNames, err
			}
			fileNames = append(fileNames, subDirFileNames...)
		} else if ext == "" || filepath.Ext(file.Name()) == ext {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}

func RunTests(testDir, outDir string) {
	var totalPassed, totalFailed int
	seenFiles := make(map[string]struct{})
	startTime := time.Now()

	testFileNames, err := getFileNames(testDir, ".sol")
	if err != nil {
		fmt.Println("Error finding tests:", err)
		return
	}
	for _, fileName := range testFileNames {
		if _, ok := seenFiles[fileName]; ok {
			continue
		}
		seenFiles[fileName] = struct{}{}
		contractNames, err := getFileNames(filepath.Join(outDir, fileName), ".json")
		if err != nil {
			fmt.Println("Error finding contract data:", err)
			return
		}
		for _, contractName := range contractNames {
			data, err := ioutil.ReadFile(filepath.Join(outDir, fileName, contractName))
			if err != nil {
				fmt.Println("Error reading contract data:", err)
				return
			}
			bytecode, ABI, path, err := extractTestData(data)
			if err != nil {
				fmt.Println("Error extracting contract data:", err)
				return
			}
			fmt.Printf("\nRunning tests for %s\n", path)
			passed, failed := runTests(bytecode, ABI)
			totalPassed += passed
			totalFailed += failed
		}
	}

	timeMs := float64(time.Since(startTime).Microseconds()) / 1000

	var result string
	if totalFailed == 0 {
		result = "ok"
	} else {
		result = "FAILED"
	}

	fmt.Printf("\nTest result: %s. %d passed; %d failed; finished in %.2fms\n", result, totalPassed, totalFailed, timeMs)
}
