package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod"
	"github.com/ethereum/go-ethereum/concrete/codegen/solgen"
)

func TestSolgen(t *testing.T) {
	tmpDir := "./tmp-solgen"
	os.Mkdir(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	config := solgen.Config{
		Name:       "CounterPrecompile",
		Address:    common.Address{0x80},
		Pragma:     "^0.8.0",
		AbiPath:    filepath.Join("..", "..", "codegen", "solgen", "testdata", "Counter.abi.json"),
		OutPath:    filepath.Join(tmpDir, "CounterPrecompile.sol"),
		ImportPath: filepath.Join("..", "..", "codegen", "solgen", "testdata", "Dependency.sol"),
	}
	cmd := exec.Command(
		"go", "run", ".", "solgen",
		"--name", config.Name,
		"--address", config.Address.Hex(),
		"--pragma", config.Pragma,
		"--abi", config.AbiPath,
		"--out", config.OutPath,
		"--import", config.ImportPath,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Log(stderr.String())
		t.Fatal(err)
	}
	t.Log(stdout.String())
}

func TestDatamod(t *testing.T) {
	tmpDir := "./tmp-datamod"
	os.Mkdir(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)
	config := datamod.Config{
		SchemaFilePath: filepath.Join("..", "..", "codegen", "datamod", "testdata", "good-datamod.json"),
		OutDir:         filepath.Join(tmpDir),
		Package:        "test",
	}
	cmd := exec.Command(
		"go", "run", ".", "datamod",
		config.SchemaFilePath,
		"--out", config.OutDir,
		"--pkg", config.Package,
		"--table-type-experimental",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Log(stderr.String())
		t.Fatal(err)
	}
	t.Log(stdout.String())
}
