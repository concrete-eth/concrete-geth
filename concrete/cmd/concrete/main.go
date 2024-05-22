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
	"github.com/naoina/toml"
	"github.com/spf13/cobra"
)

func isDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
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
	filename := strings.Split(filenameWithExt, ".")[0]
	return filename
}

func getStringFlags(cmd *cobra.Command, flagPairs ...interface{}) error {
	if len(flagPairs)%2 != 0 {
		return fmt.Errorf("flagPairs must have an even number of elements")
	}
	for i := 0; i < len(flagPairs); i += 2 {
		ptr := flagPairs[i].(*string)
		name := flagPairs[i+1].(string)
		val, err := cmd.Flags().GetString(name)
		if err != nil {
			return err
		}
		*ptr = val
	}
	return nil
}

func logMustBeProvided(cmd *cobra.Command, flagName string) {
	logFatalNoContext(fmt.Errorf("%s must be provided", flagName))
}

func logConfig(config interface{}) {
	configToml, err := toml.Marshal(config)
	if err != nil {
		logFatal(err)
	}
	logDebug(string(configToml))
}

func main() {
	var rootCmd = &cobra.Command{Use: "concrete"}

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")

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

	cmdSolgen.Flags().StringP("name", "n", "", "name for the generated library")
	cmdSolgen.Flags().StringP("address", "a", "", "precompile address")
	cmdSolgen.Flags().StringP("pragma", "p", "^0.8.0", "precompile address")
	cmdSolgen.Flags().String("abi", "", "path to the ABI file")
	cmdSolgen.Flags().StringP("out", "o", "./", "path to the output file")
	cmdSolgen.Flags().StringP("import", "i", "", "solidity file to import in the generated LIBRARY")
	rootCmd.AddCommand(cmdSolgen)

	var cmdDatamod = &cobra.Command{
		Use:   "datamod <path>",
		Short: "Generate type safe go wrappers for datastore structures from a json definition",
		Args:  cobra.MinimumNArgs(1),
		Run:   runDatamod,
	}

	cmdDatamod.Flags().StringP("out", "o", "./", "dir to write the generated files to")
	cmdDatamod.Flags().StringP("pkg", "p", "main", "package name for the generated files")
	cmdDatamod.Flags().Bool("table-type-experimental", false, "whether to enable experimental features for table types")
	rootCmd.AddCommand(cmdDatamod)

	if err := rootCmd.Execute(); err != nil {
		logFatalNoContext(err)
	}
}

func runSolgen(cmd *cobra.Command, args []string) {
	var name, addressHex, pragma, abiPath, outPath, importPath string
	if err := getStringFlags(cmd, &name, "name", &addressHex, "address", &pragma, "pragma", &abiPath, "abi", &outPath, "out", &importPath, "import"); err != nil {
		logFatal(err)
	}

	if abiPath == "" {
		logMustBeProvided(cmd, "abi path")
	}
	if outPath == "" {
		logMustBeProvided(cmd, "output path")
	}
	if addressHex == "" {
		logMustBeProvided(cmd, "precompile address")
	}

	var address common.Address
	if strings.HasPrefix(addressHex, "0x") && len(addressHex) < 42 {
		addressHex = "0x" + strings.Repeat("0", 42-len(addressHex)) + addressHex[2:]
	}
	if common.IsHexAddress(addressHex) {
		address = common.HexToAddress(addressHex)
	} else {
		logFatalNoContext(fmt.Errorf("invalid precompile address: %s", addressHex))
	}

	var err error
	var abiIsDir, outIsDir bool

	if abiIsDir, err = isDir(abiPath); err != nil {
		logFatal(err)
	}
	if abiIsDir {
		logFatalNoContext(fmt.Errorf("ABI path must be a file"))
	}

	if outIsDir, err = isDir(outPath); err != nil {
		logFatal(err)
	}
	if outIsDir {
		outPath = filepath.Join(outPath, name+".sol")
	}

	if name == "" {
		name = fileName(abiPath) + "Precompile"
	}

	config := solgen.Config{
		Name:       name,
		Address:    address,
		Pragma:     pragma,
		AbiPath:    abiPath,
		OutPath:    outPath,
		ImportPath: importPath,
	}

	if v, err := cmd.Flags().GetBool("verbose"); err != nil {
		logFatal(err)
	} else if v {
		logConfig(config)
	}

	if err := solgen.GenerateSolidityLibrary(config); err != nil {
		logFatal(err)
	}

	logInfo("Library generated successfully.")
	logInfo("Library written to: %s", outPath)
}

func runDatamod(cmd *cobra.Command, args []string) {
	jsonPath := args[0]

	var outPath, pkg string
	if err := getStringFlags(cmd, &outPath, "out", &pkg, "pkg"); err != nil {
		logFatal(err)
	}

	var err error
	var allowTableTypes bool
	if allowTableTypes, err = cmd.Flags().GetBool("table-type-experimental"); err != nil {
		logFatal(err)
	}

	var jsonIsDir, outIsDir bool

	if jsonIsDir, err = isDir(jsonPath); err != nil {
		logFatal(err)
	}
	if jsonIsDir {
		logFatalNoContext(fmt.Errorf("JSON path must be a file"))
	}

	if outIsDir, err = isDir(outPath); err != nil {
		logFatal(err)
	}
	if !outIsDir {
		logFatalNoContext(fmt.Errorf("output path must be a directory"))
	}

	config := datamod.Config{
		SchemaFilePath: jsonPath,
		OutDir:         outPath,
		Package:        pkg,
	}

	if v, err := cmd.Flags().GetBool("verbose"); err != nil {
		logFatal(err)
	} else if v {
		logConfig(config)
	}

	if err := datamod.GenerateDataModel(config, allowTableTypes); err != nil {
		logFatal(err)
	}

	logInfo("Data model wrappers generated successfully.")
	logInfo("Files written to: %s", outPath)
}
