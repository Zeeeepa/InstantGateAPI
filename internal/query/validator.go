package query

import (
	"fmt"
	"strconv"
	"time"

	"github.com/proyaai/instantgate/internal/database/mysql"
)

func ValidateColumn(col mysql.ColumnInfo, value interface{}) error {
	if value == nil {
		if !col.Nullable {
			return fmt.Errorf("column '%s' does not allow NULL values", col.Name)
		}
		return nil
	}

	switch col.GoType {
	case "int64":
		if _, ok := value.(float64); ok {
			return nil
		}
		if _, ok := value.(int64); ok {
			return nil
		}
		return fmt.Errorf("column '%s' expects an integer, got %T", col.Name, value)

	case "float64":
		if _, ok := value.(float64); ok {
			return nil
		}
		return fmt.Errorf("column '%s' expects a number, got %T", col.Name, value)

	case "string":
		if _, ok := value.(string); ok {
			if col.MaxLength.Valid && len(value.(string)) > int(col.MaxLength.Int64) {
				return fmt.Errorf("column '%s' exceeds max length of %d", col.Name, col.MaxLength.Int64)
			}
			return nil
		}
		return fmt.Errorf("column '%s' expects a string, got %T", col.Name, value)

	case "bool":
		if _, ok := value.(bool); ok {
			return nil
		}
		return fmt.Errorf("column '%s' expects a boolean, got %T", col.Name, value)

	case "time.Time":
		if str, ok := value.(string); ok {
			_, err := time.Parse(time.RFC3339, str)
			if err != nil {
				_, err = time.Parse("2006-01-02", str)
				if err != nil {
					return fmt.Errorf("column '%s' expects a valid date/time format", col.Name)
				}
			}
			return nil
		}
		return fmt.Errorf("column '%s' expects a date/time string, got %T", col.Name, value)
	}

	return nil
}

func ValidateRow(table *mysql.TableSchema, data map[string]interface{}) error {
	for key, value := range data {
		col, ok := table.Columns[key]
		if !ok {
			return fmt.Errorf("unknown column '%s'", key)
		}

		if err := ValidateColumn(col, value); err != nil {
			return err
		}
	}
	return nil
}

func ParseAndValidateInt(value interface{}) (int64, error) {
	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("invalid integer value: %v", value)
	}
}

func ParseAndValidateFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("invalid float value: %v", value)
	}
}

func IsNumericType(goType string) bool {
	switch goType {
	case "int", "int8", "int16", "int32", "int64":
		return true
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	case "float32", "float64":
		return true
	default:
		return false
	}
}
