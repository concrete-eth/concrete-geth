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

package solgen

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

//go:embed solgen.tpl
var solgenTpl string

type Config struct {
	Name    string
	Address common.Address
	ABI     string
	Sol     string
	Out     string
}

func isValidSolidityContractName(name string) bool {
	if len(name) == 0 {
		return false
	}
	re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return re.MatchString(name) && len(strings.TrimSpace(name)) == len(name)
}

func formatPath(path string) string {
	if len(path) == 0 {
		return path
	}
	if filepath.IsAbs(path) {
		return path
	} else if path[0] == '.' {
		return path
	} else {
		return "./" + path
	}
}

func getTypeString(internalType string, arg abi.Argument) string {
	if strings.HasPrefix(internalType, "struct ") {
		return strings.TrimPrefix(internalType, "struct ")
	} else {
		return arg.Type.String()
	}
}

func withLocation(typeStr string, arg abi.Argument) string {
	var argName string
	if len(arg.Name) > 0 {
		argName = " " + arg.Name
	}
	if typeStr == arg.Type.String() {
		if typeStr == "bytes" || typeStr == "string" || strings.Contains(typeStr, "[") {
			return typeStr + " memory" + argName
		}
		return typeStr + argName
	}
	return typeStr + " memory" + argName
}

func generateSolidityLibrary(ABI abi.ABI, cABI customABI, config Config) (string, error) {
	config.Sol = filepath.ToSlash(config.Sol)

	tmpl, err := template.New("solgen").Parse(solgenTpl)
	if err != nil {
		return "", err
	}

	if !isValidSolidityContractName(config.Name) {
		return "", fmt.Errorf("invalid contract name: '%s'", config.Name)
	}

	var importPath string
	if len(config.Sol) > 0 {
		importPath, err = filepath.Rel(filepath.Dir(config.Out), config.Sol)
		if err != nil {
			return "", err
		}
		importPath = formatPath(importPath)
	}

	data := map[string]interface{}{
		"Name":        config.Name,
		"Address":     config.Address.Hex(),
		"Methods":     []map[string]interface{}{},
		"ImportPaths": []string{importPath},
	}

	for mIdx, method := range ABI.Methods {
		inputSig := []string{}
		inputTypes := []string{}
		inputNames := []string{}
		for inIdx, input := range method.Inputs {
			internalType := cABI.MethodsByName[mIdx].Inputs[inIdx].InternalType
			typeStr := getTypeString(internalType, input)
			inputSig = append(inputSig, withLocation(typeStr, input))
			inputTypes = append(inputTypes, typeStr)
			inputNames = append(inputNames, input.Name)
		}

		outputSig := []string{}
		outputTypes := []string{}
		for outIdx, output := range method.Outputs {
			internalType := cABI.MethodsByName[mIdx].Outputs[outIdx].InternalType
			typeStr := getTypeString(internalType, output)
			outputSig = append(outputSig, withLocation(typeStr, output))
			outputTypes = append(outputTypes, typeStr)
		}

		methodData := map[string]interface{}{
			"Name":        method.Name,
			"Signature":   method.Sig,
			"IsStatic":    method.IsConstant(),
			"Inputs":      strings.Join(inputSig, ", "),
			"Outputs":     strings.Join(outputSig, ", "),
			"InputTypes":  strings.Join(inputTypes, ", "),
			"OutputTypes": strings.Join(outputTypes, ", "),
			"InputNames":  strings.Join(inputNames, ", "),
		}
		data["Methods"] = append(data["Methods"].([]map[string]interface{}), methodData)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type customMethod struct {
	Name    string                   `json:"name"`
	Inputs  []abi.ArgumentMarshaling `json:"inputs"`
	Outputs []abi.ArgumentMarshaling `json:"outputs"`
}

type customABI struct {
	Methods       []customMethod           `json:"methods"`
	MethodsByName map[string]*customMethod `json:"-"`
}

func (c *customABI) UnmarshalJSON(data []byte) error {
	type Alias []customMethod
	var a Alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	c.Methods = a
	c.MethodsByName = make(map[string]*customMethod)
	for i := range c.Methods {
		c.MethodsByName[c.Methods[i].Name] = &c.Methods[i]
	}
	return nil
}

func GetABI(path string) (abi.ABI, customABI, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return abi.ABI{}, customABI{}, err
	}

	var jsonData struct {
		ABI abi.ABI `json:"abi"`
	}
	var cJsonData struct {
		ABI customABI `json:"abi"`
	}
	err = json.Unmarshal(content, &jsonData)
	if err == nil {
		err = json.Unmarshal(content, &cJsonData)
		if err != nil {
			return abi.ABI{}, customABI{}, err
		}
		return jsonData.ABI, cJsonData.ABI, nil
	}

	var ABI abi.ABI
	var cABI customABI
	err = json.Unmarshal(content, &ABI)
	if err == nil {
		err = json.Unmarshal(content, &cABI)
		if err != nil {
			return abi.ABI{}, customABI{}, err
		}
		return ABI, cABI, nil
	}

	return abi.ABI{}, customABI{}, err
}

func GenerateSolidityLibrary(config Config) error {
	ABI, cABI, err := GetABI(config.ABI)
	if err != nil {
		return err
	}
	code, err := generateSolidityLibrary(ABI, cABI, config)
	if err != nil {
		return err
	}
	err = os.WriteFile(config.Out, []byte(code), 0644)
	if err != nil {
		return err
	}
	return nil
}
