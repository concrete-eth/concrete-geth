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

package testtool

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"golang.org/x/exp/slog"
)

// TODO: WIP -- missing parameter and file validation, and more

var (
	PrintLogs = true
)

func runTestMethod(t *testing.T, concreteRegistry concrete.PrecompileRegistry, bytecode []byte, method abi.Method, shouldFail bool) {
	var (
		key, _          = crypto.HexToECDSA("d17bd946feb884d463d58fb702b94dd0457ca349338da1d732a57856cf777ccd") // 0xCcca11AbAC28D9b6FceD3a9CA73C434f6b33B215
		senderAddress   = crypto.PubkeyToAddress(key.PublicKey)
		contractAddress = common.HexToAddress("cc73570000000000000000000000000000000000")
		gspec           = &core.Genesis{
			GasLimit: 2e9,
			Config:   params.TestChainConfig,
			Alloc: types.GenesisAlloc{
				senderAddress:   {Balance: big.NewInt(1e18)},
				contractAddress: {Balance: common.Big0, Code: bytecode},
			},
		}
		signer   = types.LatestSigner(gspec.Config)
		gasLimit = uint64(1e7)
		setupId  = crypto.Keccak256([]byte("setUp()"))[:4]
	)

	_, _, receipts := core.GenerateChainWithGenesisWithConcrete(gspec, ethash.NewFaker(), 1, concreteRegistry, func(ii int, block *core.BlockGen) {
		for _, id := range [][]byte{setupId, method.ID} {
			tx := types.NewTransaction(block.TxNonce(senderAddress), contractAddress, common.Big0, gasLimit, block.BaseFee(), id)
			signed, err := types.SignTx(tx, signer, key)
			if err != nil {
				t.Fatal(err)
			}
			block.AddTx(signed)
		}
	})

	if len(receipts[0]) != 2 {
		t.Fatalf("expected 2 receipts, got %d", len(receipts[0]))
	}
	setupReceipt := receipts[0][0]
	testReceipt := receipts[0][1]

	if setupReceipt.Status != types.ReceiptStatusSuccessful {
		t.Fatal("setup failed")
	}
	if (testReceipt.Status == types.ReceiptStatusSuccessful) == shouldFail {
		t.Fail()
	}

	t.Logf("Gas used: %d", testReceipt.GasUsed)

	if PrintLogs && len(testReceipt.Logs) > 0 {
		for ii, log := range testReceipt.Logs {
			logStr := fmt.Sprintf("\nLogs[%d]\nAddress : %s\n", ii, log.Address)
			if len(log.Topics) > 0 {
				logStr += fmt.Sprintf("Topics  : %s\n", log.Topics[0].String())
				for _, topic := range log.Topics[1:] {
					logStr += fmt.Sprintf("         : %s\n", topic.String())
				}
			}
			logStr += fmt.Sprintf("Data    : 0x%x\n", log.Data)
			t.Log(logStr)
		}
	}
}

func runTestContract(t *testing.T, concreteRegistry concrete.PrecompileRegistry, bytecode []byte, ABI abi.ABI) {
	for _, method := range ABI.Methods {
		if !strings.HasPrefix(method.Name, "test") {
			continue
		}
		t.Run(method.Name, func(t *testing.T) {
			shouldFail := strings.HasPrefix(method.Name, "testFail")
			runTestMethod(t, concreteRegistry, bytecode, method, shouldFail)
		})
	}
}

func extractTestData(contractJsonBytes []byte) ([]byte, abi.ABI, string, string, error) {
	var jsonData struct {
		ABI              abi.ABI `json:"abi"`
		DeployedBytecode struct {
			Object string `json:"object"`
		} `json:"deployedBytecode"`
		Metadata struct {
			Settings struct {
				CompilationTarget map[string]string `json:"compilationTarget"`
			} `json:"settings"`
		} `json:"metadata"`
	}
	err := json.Unmarshal(contractJsonBytes, &jsonData)
	if err != nil {
		return nil, abi.ABI{}, "", "", err
	}
	bytecode := common.FromHex(jsonData.DeployedBytecode.Object)
	var testPath, contractName string
	for path, name := range jsonData.Metadata.Settings.CompilationTarget {
		testPath = path
		contractName = name
		break
	}
	return bytecode, jsonData.ABI, testPath, contractName, nil
}

func extractTestDataFromPath(contractJsonPath string) ([]byte, abi.ABI, string, string, error) {
	contractJsonBytes, err := os.ReadFile(contractJsonPath)
	if err != nil {
		return nil, abi.ABI{}, "", "", err
	}
	return extractTestData(contractJsonBytes)
}

func getFileNames(dir string, ext string) ([]string, error) {
	var fileNames []string

	files, err := os.ReadDir(dir)
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

func getTestPaths(testDir, outDir string) ([]string, error) {
	paths := make([]string, 0)
	seenFiles := make(map[string]struct{})
	testFileNames, err := getFileNames(testDir, ".sol")
	if err != nil {
		return nil, err
	}
	for _, fileName := range testFileNames {
		if _, ok := seenFiles[fileName]; ok {
			continue
		}
		seenFiles[fileName] = struct{}{}
		contractNames, err := getFileNames(filepath.Join(outDir, fileName), ".json")
		if err != nil {
			return nil, err
		}
		for _, contractName := range contractNames {
			path := filepath.Join(outDir, fileName, contractName)
			paths = append(paths, path)
		}
	}
	return paths, nil
}

func runTestPaths(t *testing.T, concreteRegistry concrete.PrecompileRegistry, contractJsonPaths []string) {
	for _, path := range contractJsonPaths {
		bytecode, ABI, testPath, contractName, err := extractTestDataFromPath(path)
		if err != nil {
			t.Fatalf("Error extracting test data from %s: %s\n", path, err)
			continue
		}
		t.Logf("\nRunning tests for %s:%s\n", testPath, contractName)
		runTestContract(t, concreteRegistry, bytecode, ABI)
	}
}

func setGethVerbosity(lvl slog.Level) func() {
	handler := log.Root()
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, lvl, true)))
	return func() {
		log.SetDefault(handler)
	}
}

type TestConfig struct {
	Contract string
	TestDir  string
	OutDir   string
}

func RunTestContract(t *testing.T, concreteRegistry concrete.PrecompileRegistry, bytecode []byte, ABI abi.ABI) {
	resetGethLogger := setGethVerbosity(log.LevelWarn)
	defer resetGethLogger()
	runTestContract(t, concreteRegistry, bytecode, ABI)
}

func Test(t *testing.T, concreteRegistry concrete.PrecompileRegistry, config TestConfig) {
	resetGethLogger := setGethVerbosity(log.LevelWarn)
	defer resetGethLogger()

	// Get test paths
	var testPaths []string

	if config.Contract != "" {
		parts := strings.SplitN(config.Contract, ":", 2)
		if len(parts) != 2 {
			t.Fatalf("Invalid contract: %s. Must follow format Path:Contract\n", config.Contract)
		}
		_, fileName := filepath.Split(parts[0])
		contractName := parts[1]
		path := filepath.Join(config.OutDir, fileName, contractName+".json")
		testPaths = []string{path}
	} else {
		var err error
		testPaths, err = getTestPaths(config.TestDir, config.OutDir)
		if err != nil {
			t.Fatalf("Error getting test paths: %s\n", err)
		}
	}

	// Run tests
	runTestPaths(t, concreteRegistry, testPaths)
}
