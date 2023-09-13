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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod"
	"github.com/ethereum/go-ethereum/concrete/codegen/solgen"
	"github.com/ethereum/go-ethereum/internal/version"
	"github.com/spf13/cobra"
)

func exit(msg string) {
	fmt.Println("--------------------------------")
	fmt.Println("[ERROR]")
	fmt.Println(msg)
	os.Exit(1)
}

func checkErr(err error) {
	if err != nil {
		exit(err.Error())
	}
}

func isDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return true, nil
	} else {
		return false, nil
	}
}

func fileName(path string) string {
	filenameWithExt := filepath.Base(path)
	filename := strings.TrimSuffix(filenameWithExt, filepath.Ext(filenameWithExt))
	return filename
}

func main() {
	var rootCmd = &cobra.Command{Use: "concrete"}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Info())
		},
	}

	rootCmd.AddCommand(versionCmd)

	var cmdSolgen = &cobra.Command{
		Use:   "solgen",
		Short: "Generate a solidity precompile caller library from an ABI file",
		Run:   runSolgen,
	}

	cmdSolgen.Flags().String("abi", "", "path to the ABI file")
	cmdSolgen.Flags().String("out", "./", "path to the output file")
	cmdSolgen.Flags().String("solidity", "", "path to the solidity file")
	cmdSolgen.Flags().StringP("name", "n", "", "name for the generated library")
	cmdSolgen.Flags().StringP("address", "a", "", "precompile address")
	rootCmd.AddCommand(cmdSolgen)

	var cmdDatamod = &cobra.Command{
		Use:   "datamod <path>",
		Short: "Generate type safe go wrappers for datastore structures from a json definition",
		Args:  cobra.MinimumNArgs(1),
		Run:   runDatamod,
	}

	cmdDatamod.Flags().String("out", "./", "dir to write the generated files to")
	cmdDatamod.Flags().String("pkg", "main", "package name for the generated files")
	cmdDatamod.Flags().Bool("table-types-experimental", false, "whether to enable experimental table value type")
	rootCmd.AddCommand(cmdDatamod)

	if err := rootCmd.Execute(); err != nil {
		exit(err.Error())
	}
}

func runSolgen(cmd *cobra.Command, args []string) {
	abiPath, err := cmd.Flags().GetString("abi")
	checkErr(err)
	outPath, err := cmd.Flags().GetString("out")
	checkErr(err)
	solPath, err := cmd.Flags().GetString("solidity")
	checkErr(err)
	name, err := cmd.Flags().GetString("name")
	checkErr(err)
	address, err := cmd.Flags().GetString("address")
	checkErr(err)

	if abiPath == "" {
		exit("ABI file path (--abi) must be provided")
	}
	if outPath == "" {
		exit("Output file path (--out) must be provided")
	}

	if address == "" {
		exit("Precompile address (--address) must be provided")
	} else if !common.IsHexAddress(common.HexToAddress(address).Hex()) {
		exit("Precompile address (--address) must be a valid hex address")
	} else {
		address = common.HexToAddress(address).Hex()
	}

	abiIsDir, err := isDir(abiPath)
	checkErr(err)
	outIsDir, err := isDir(outPath)
	checkErr(err)

	if abiIsDir {
		exit("ABI path must be a file")
	}

	if name == "" {
		name = fileName(abiPath) + "Precompile"
	}

	if outIsDir {
		outPath = filepath.Join(outPath, name+".sol")
	}

	config := solgen.Config{
		Name:    name,
		Address: common.HexToAddress(address),
		ABI:     abiPath,
		Out:     outPath,
		Sol:     solPath,
	}

	fmt.Printf(`Generating solidity library
Name     : %s
Address  : %s
ABI      : %s
Output   : %s
Solidity : %s
`, name, address, abiPath, outPath, solPath)

	err = solgen.GenerateSolidityLibrary(config)
	checkErr(err)

	fmt.Printf("Library generated successfully.\nLibrary written to: %s\n", outPath)
}

func runDatamod(cmd *cobra.Command, args []string) {
	jsonPath := args[0]
	outPath, err := cmd.Flags().GetString("out")
	checkErr(err)
	pkg, err := cmd.Flags().GetString("pkg")
	checkErr(err)
	allowTableTypes, err := cmd.Flags().GetBool("table-types-experimental")
	checkErr(err)

	jsonIsDir, err := isDir(jsonPath)
	checkErr(err)
	if jsonIsDir {
		exit("JSON path must be a file")
	}

	outIsDir, err := isDir(outPath)
	checkErr(err)
	if !outIsDir {
		exit("Output path must be a directory")
	}

	config := datamod.Config{
		JSON:    jsonPath,
		Out:     outPath,
		Package: pkg,
	}

	fmt.Println("Generating data model wrappers for:", jsonPath)

	err = datamod.GenerateDataModel(config, allowTableTypes)
	checkErr(err)

	fmt.Println("Data model wrappers generated successfully.\nFiles written to:", outPath)
}
