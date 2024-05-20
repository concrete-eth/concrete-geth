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
	"strings"
	"text/template"

	"github.com/iancoleman/orderedmap"
)

//go:embed table.tpl
var tableTpl string

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

func newFieldSchema(name string, index int, typeStr string) (FieldSchema, error) {
	if !isValidName(name) {
		return FieldSchema{}, fmt.Errorf("invalid field name '%s'", name)
	}
	fieldType, err := nameToFieldType(typeStr)
	if err != nil {
		return FieldSchema{}, fmt.Errorf("invalid type '%s' for field '%s': %w", typeStr, name, err)
	}
	return FieldSchema{
		Name:  lowerFirstLetter(name),
		Title: upperFirstLetter(name),
		Index: index,
		Type:  fieldType,
	}, nil
}

func UnmarshalTableSchemas(jsonContent []byte, allowTableTypes bool) ([]TableSchema, error) {
	jsonSchemas := orderedmap.New()
	err := json.Unmarshal(jsonContent, &jsonSchemas)
	if err != nil {
		return []TableSchema{}, err
	}

	var tableSchemas []TableSchema
	for _, tableName := range jsonSchemas.Keys() {
		_jsonTableSchema, _ := jsonSchemas.Get(tableName)
		jsonTableSchema, ok := _jsonTableSchema.(orderedmap.OrderedMap)
		if !ok {
			return []TableSchema{}, fmt.Errorf("invalid schema for table '%s'", tableName)
		}

		if !isValidName(tableName) {
			return []TableSchema{}, fmt.Errorf("invalid table name '%s'", tableName)
		}
		if len(jsonTableSchema.Keys()) == 0 {
			return []TableSchema{}, fmt.Errorf("no schema for table '%s'", tableName)
		}

		tableSchema := TableSchema{Name: upperFirstLetter(tableName)}

		_jsonKeySchema, ok := jsonTableSchema.Get("keySchema")
		if ok {
			jsonKeySchema, ok := _jsonKeySchema.(orderedmap.OrderedMap)
			if !ok {
				return []TableSchema{}, fmt.Errorf("invalid key schema for table '%s'", tableName)
			}
			for _, keyName := range jsonKeySchema.Keys() {
				_keyType, _ := jsonKeySchema.Get(keyName)
				keyType, ok := _keyType.(string)
				if !ok {
					return []TableSchema{}, fmt.Errorf("invalid schema for key '%s' in table '%s'", keyName, tableName)
				}
				fieldSchema, err := newFieldSchema(keyName, len(tableSchema.Keys), keyType)
				if err != nil {
					return []TableSchema{}, err
				}
				if fieldSchema.Type.Type == TableType {
					return []TableSchema{}, fmt.Errorf("table '%s' cannot have table keys", tableName)
				}
				tableSchema.Keys = append(tableSchema.Keys, fieldSchema)
			}
		}

		_jsonValueSchema, ok := jsonTableSchema.Get("schema")
		if !ok {
			return []TableSchema{}, fmt.Errorf("no value schema for table '%s'", tableName)
		}
		jsonValueSchema, ok := _jsonValueSchema.(orderedmap.OrderedMap)
		if !ok {
			return []TableSchema{}, fmt.Errorf("invalid value schema for table '%s'", tableName)
		}
		for _, valueName := range jsonValueSchema.Keys() {
			_valueType, _ := jsonValueSchema.Get(valueName)
			valueType, ok := _valueType.(string)
			if !ok {
				return []TableSchema{}, fmt.Errorf("invalid schema for value '%s' in table '%s'", valueName, tableName)
			}
			fieldSchema, err := newFieldSchema(valueName, len(tableSchema.Values), valueType)
			if err != nil {
				return []TableSchema{}, err
			}
			if fieldSchema.Type.Type == TableType {
				if !allowTableTypes {
					return []TableSchema{}, fmt.Errorf("invalid type '%s' for field '%s': table values cannot be tables", fieldSchema.Type.Name, fieldSchema.Name)
				}
				_, ok := jsonSchemas.Get(fieldSchema.Type.Name)
				if !ok {
					return []TableSchema{}, fmt.Errorf("table '%s' does not exist", fieldSchema.Type.Name)
				}
			}
			tableSchema.Values = append(tableSchema.Values, fieldSchema)
		}
		tableSchemas = append(tableSchemas, tableSchema)
	}
	return tableSchemas, nil
}

type Config struct {
	SchemaFilePath string
	OutDir         string
	Package        string
}

func GenerateDataModel(config Config, allowTableTypes bool) error {
	if !isValidName(config.Package) {
		return fmt.Errorf("invalid package name: %s", config.Package)
	}

	jsonContent, err := os.ReadFile(config.SchemaFilePath)
	if err != nil {
		return err
	}
	schemas, err := UnmarshalTableSchemas(jsonContent, allowTableTypes)
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
	}

	for _, schema := range schemas {
		tableName := formatTableName(schema.Name)
		rowName := formatRowName(schema.Name)

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

		filename := lowerFirstLetter(tableName) + ".go"

		tpl, err := template.New("table").Funcs(funcMap).Parse(tableTpl)
		if err != nil {
			return err
		}
		if err = ExecuteTemplate(tpl, data, filepath.Join(config.OutDir, filename+".go")); err != nil {
			return err
		}
	}
	return nil
}

func ExecuteTemplate(tpl *template.Template, data map[string]interface{}, outPath string) error {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}
	return os.WriteFile(outPath, buf.Bytes(), 0644)
}
