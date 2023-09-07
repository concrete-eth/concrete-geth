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

//go:embed table.tpl
var tableTpl string

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

type TableSchema struct {
	Name   string
	Keys   []FieldSchema
	Values []FieldSchema
}

type TableSchemaUnmarshal struct {
	Keys   map[string]string `json:"keySchema"`
	Values map[string]string `json:"schema"`
}

type TableSchemasUnmarshal map[string]TableSchemaUnmarshal

func newFieldSchema(name string, index int, typeStr string) (FieldSchema, error) {
	if !isValidName(name) {
		return FieldSchema{}, fmt.Errorf("invalid field name: %s", name)
	}
	fieldType, err := nameToFieldType(typeStr)
	if err != nil {
		return FieldSchema{}, err
	}
	return FieldSchema{
		Name:  lowerFirstLetter(name),
		Title: upperFirstLetter(name),
		Index: index,
		Type:  fieldType,
	}, nil
}

func unmarshalTableSchemas(jsonContent []byte) ([]TableSchema, error) {
	var unmarshaledSchemas TableSchemasUnmarshal
	err := json.Unmarshal(jsonContent, &unmarshaledSchemas)
	if err != nil {
		return []TableSchema{}, err
	}

	var tableSchemas []TableSchema
	for tableName, unmarshaledTableSchema := range unmarshaledSchemas {
		if !isValidName(tableName) {
			return []TableSchema{}, fmt.Errorf("invalid table name: %s", tableName)
		}
		if len(unmarshaledTableSchema.Values) == 0 {
			return []TableSchema{}, fmt.Errorf("no values in table: '%s'", tableName)
		}

		tableSchema := TableSchema{
			Name: upperFirstLetter(tableName),
		}
		for keyName, keyType := range unmarshaledTableSchema.Keys {
			fieldSchema, err := newFieldSchema(keyName, len(tableSchema.Keys), keyType)
			if err != nil {
				return []TableSchema{}, err
			}
			tableSchema.Keys = append(tableSchema.Keys, fieldSchema)
		}
		for valueName, valueType := range unmarshaledTableSchema.Values {
			fieldSchema, err := newFieldSchema(valueName, len(tableSchema.Values), valueType)
			if err != nil {
				return []TableSchema{}, err
			}
			tableSchema.Values = append(tableSchema.Values, fieldSchema)
		}
		tableSchemas = append(tableSchemas, tableSchema)
	}
	return tableSchemas, nil
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
	schemas, err := unmarshalTableSchemas(jsonContent)
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
	}
	tpl, err := template.New("table").Funcs(funcMap).Parse(tableTpl)
	if err != nil {
		return err
	}

	for _, schema := range schemas {
		tableName := schema.Name
		rowName := schema.Name + "Row"

		_sizes := make([]string, len(schema.Values))
		for i, field := range schema.Values {
			_sizes[i] = fmt.Sprint(field.Type.Size)
		}
		sizesStr := fmt.Sprintf("[]int{%s}", strings.Join(_sizes, ", "))

		_keys := make([]string, len(schema.Keys))
		for i, field := range schema.Keys {
			_keys[i] = fmt.Sprint(field.Type.Size)
		}

		data := map[string]interface{}{
			"Package":         config.Package,
			"Schema":          schema,
			"TableStructName": tableName,
			"RowStructName":   rowName,
			"SizesStr":        sizesStr,
		}

		var buf bytes.Buffer
		if err := tpl.Execute(&buf, data); err != nil {
			return err
		}
		outPath := filepath.Join(config.Out, tableName+".go")
		err := os.WriteFile(outPath, buf.Bytes(), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
