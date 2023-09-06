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
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"
)

//go:embed schema.tpl
var schemaTpl string

func lowerFirstLetter(str string) string {
	if len(str) == 0 {
		return ""
	}
	runes := []rune(str)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func upperFirstLetter(str string) string {
	if len(str) == 0 {
		return ""
	}
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func isValidName(name string) bool {
	if len(name) == 0 {
		return false
	}
	re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return re.MatchString(name) && len(strings.TrimSpace(name)) == len(name)
}

type FieldSchema struct {
	Name  string
	Title string
	Index int
	Type  FieldType
}

type MappingSchema struct {
	Name   string
	Keys   []FieldSchema
	Values []FieldSchema
}

type ModelSchema []MappingSchema

type MappingUnmarshal struct {
	KeySchema map[string]string `json:"keySchema"`
	Schema    map[string]string `json:"schema"`
}

type ModelUnmarshal map[string]MappingUnmarshal

func newFieldSchema(name string, index int, typeStr string) (FieldSchema, error) {
	if !isValidName(name) {
		return FieldSchema{}, fmt.Errorf("invalid name: %s", name)
	}
	fieldType, ok := NameToFieldType[typeStr]
	if !ok {
		return FieldSchema{}, fmt.Errorf("invalid type: %s", typeStr)
	}
	return FieldSchema{
		Name:  lowerFirstLetter(name),
		Title: upperFirstLetter(name),
		Index: index,
		Type:  fieldType,
	}, nil
}

func unmarshalModel(jsonContent []byte) (ModelSchema, error) {
	var unmarshaledModel ModelUnmarshal
	err := json.Unmarshal(jsonContent, &unmarshaledModel)
	if err != nil {
		return ModelSchema{}, err
	}

	var model ModelSchema
	for name, mapping := range unmarshaledModel {
		if !isValidName(name) {
			return ModelSchema{}, fmt.Errorf("invalid name: %s", name)
		}
		newMapping := MappingSchema{Name: upperFirstLetter(name)}
		for keyName, keyType := range mapping.KeySchema {
			fieldSchema, err := newFieldSchema(keyName, len(newMapping.Keys), keyType)
			if err != nil {
				return ModelSchema{}, err
			}
			newMapping.Keys = append(newMapping.Keys, fieldSchema)
		}
		for valueName, valueType := range mapping.Schema {
			fieldSchema, err := newFieldSchema(valueName, len(newMapping.Values), valueType)
			if err != nil {
				return ModelSchema{}, err
			}
			newMapping.Values = append(newMapping.Values, fieldSchema)
		}
		model = append(model, newMapping)
	}
	return model, nil
}

type Config struct {
	JSON    string
	Out     string
	Package string
}

func GenerateDataModel(config Config) error {
	if !isValidName(config.Package) {
		return fmt.Errorf("invalid package name: %s", config.Package)
	}

	jsonContent, err := os.ReadFile(config.JSON)
	if err != nil {
		return err
	}
	model, err := unmarshalModel(jsonContent)
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
	}
	tmpl, err := template.New("struct").Funcs(funcMap).Parse(schemaTpl)
	if err != nil {
		return err
	}

	for _, mapping := range model {
		mappingName := mapping.Name
		structName := mapping.Name + "Item"

		_sizes := make([]string, len(mapping.Values))
		for i, field := range mapping.Values {
			_sizes[i] = fmt.Sprint(field.Type.Size)
		}
		sizesStr := fmt.Sprintf("[]int{%s}", strings.Join(_sizes, ", "))

		_keys := make([]string, len(mapping.Keys))
		for i, field := range mapping.Keys {
			_keys[i] = fmt.Sprint(field.Type.Size)
		}

		data := map[string]interface{}{
			"Package":     config.Package,
			"Schema":      mapping,
			"MappingName": mappingName,
			"StructName":  structName,
			"SizesStr":    sizesStr,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return err
		}
		outPath := filepath.Join(config.Out, lowerFirstLetter(mapping.Name)+".go")
		err := os.WriteFile(outPath, buf.Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
