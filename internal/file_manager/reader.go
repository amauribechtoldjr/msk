package file_manager

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
)

func Read() {
	filepath := "./examples/test.csv"

	openFile, _ := os.Open(filepath)

	r := csv.NewReader(openFile)

	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(record[1])
	}
}

func ReadCSV[T any](filePath string) ([]T, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("csv must have header and at least one row")
	}

	headers := records[0]

	var result []T
	var sample T
	tType := reflect.TypeOf(sample)

	if tType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("generic type must be a struct")
	}

	// Map header -> struct field index
	fieldMap := make(map[string]int)

	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)
		name := field.Tag.Get("csv")
		if name == "" {
			name = field.Name
		}
		fieldMap[name] = i
	}

	fmt.Println("cheguei aqui")

	// Validate headers
	for _, h := range headers {
		if _, ok := fieldMap[h]; !ok {
			return nil, fmt.Errorf("csv column '%s' does not match any struct field", h)
		}
	}
fmt.Println("cheguei aqui3")
	// Parse rows
	for _, row := range records[1:] {
		v := reflect.New(tType).Elem()

		for colIndex, value := range row {
			fieldIndex := fieldMap[headers[colIndex]]
			field := v.Field(fieldIndex)

			if !field.CanSet() {
				continue
			}

			if err := setValue(field, value); err != nil {
				return nil, fmt.Errorf("error parsing field '%s': %w", headers[colIndex], err)
			}
		}

		result = append(result, v.Interface().(T))
	}
fmt.Println("cheguei aqui 2")
	return result, nil
}

func setValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(v)

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(v)

	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(v)

	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}