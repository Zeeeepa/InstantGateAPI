package mysql

import (
	"fmt"
)

type TypeMapping struct {
	GoType     string
	SQLType    string
	IsNumeric  bool
	IsText     bool
	IsTime     bool
	IsBinary   bool
}

var mysqlTypeMap = map[string]TypeMapping{
	"tinyint":    {GoType: "int64", SQLType: "integer", IsNumeric: true},
	"smallint":   {GoType: "int64", SQLType: "integer", IsNumeric: true},
	"mediumint":  {GoType: "int64", SQLType: "integer", IsNumeric: true},
	"int":        {GoType: "int64", SQLType: "integer", IsNumeric: true},
	"integer":    {GoType: "int64", SQLType: "integer", IsNumeric: true},
	"bigint":     {GoType: "int64", SQLType: "bigint", IsNumeric: true},
	"float":      {GoType: "float64", SQLType: "float", IsNumeric: true},
	"double":     {GoType: "float64", SQLType: "double", IsNumeric: true},
	"decimal":    {GoType: "float64", SQLType: "decimal", IsNumeric: true},
	"numeric":    {GoType: "float64", SQLType: "decimal", IsNumeric: true},
	"bit":        {GoType: "int64", SQLType: "integer", IsNumeric: true},

	"char":       {GoType: "string", SQLType: "text", IsText: true},
	"varchar":    {GoType: "string", SQLType: "text", IsText: true},
	"tinytext":   {GoType: "string", SQLType: "text", IsText: true},
	"text":       {GoType: "string", SQLType: "text", IsText: true},
	"mediumtext": {GoType: "string", SQLType: "text", IsText: true},
	"longtext":   {GoType: "string", SQLType: "text", IsText: true},

	"binary":     {GoType: "[]byte", SQLType: "blob", IsBinary: true},
	"varbinary":  {GoType: "[]byte", SQLType: "blob", IsBinary: true},
	"tinyblob":   {GoType: "[]byte", SQLType: "blob", IsBinary: true},
	"blob":       {GoType: "[]byte", SQLType: "blob", IsBinary: true},
	"mediumblob": {GoType: "[]byte", SQLType: "blob", IsBinary: true},
	"longblob":   {GoType: "[]byte", SQLType: "blob", IsBinary: true},

	"date":       {GoType: "time.Time", SQLType: "date", IsTime: true},
	"datetime":   {GoType: "time.Time", SQLType: "timestamp", IsTime: true},
	"timestamp":  {GoType: "time.Time", SQLType: "timestamp", IsTime: true},
	"time":       {GoType: "time.Time", SQLType: "time", IsTime: true},
	"year":       {GoType: "int64", SQLType: "integer", IsNumeric: true},

	"json":       {GoType: "string", SQLType: "json", IsText: true},

	"enum":       {GoType: "string", SQLType: "text", IsText: true},
	"set":        {GoType: "string", SQLType: "text", IsText: true},

	"bool":       {GoType: "bool", SQLType: "boolean", IsNumeric: true},
}

func GetTypeMapping(mysqlType string) (TypeMapping, bool) {
	mapping, ok := mysqlTypeMap[mysqlType]
	return mapping, ok
}

func ParseType(typeDef string) string {
	typeDef = normalizeType(typeDef)

	for i, c := range typeDef {
		if c == '(' || c == ' ' {
			return typeDef[:i]
		}
	}
	return typeDef
}

func normalizeType(t string) string {
	suffixes := []string{" unsigned", " signed", " zerofill", " unsigned zerofill"}

	for _, suffix := range suffixes {
		if idx := indexIgnoreCase(t, suffix); idx != -1 {
			t = t[:idx]
		}
	}

	return t
}

func indexIgnoreCase(s, substr string) int {
	s = toLower(s)
	substr = toLower(substr)

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}

	return -1
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func ScanValue(columnType string, value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	baseType := ParseType(columnType)
	mapping, ok := GetTypeMapping(baseType)
	if !ok {
		return value, nil
	}

	switch v := value.(type) {
	case []byte:
		return scanBytes(v, mapping)
	default:
		return value, nil
	}
}

func scanBytes(b []byte, mapping TypeMapping) (interface{}, error) {
	if mapping.IsNumeric {
		return string(b), nil
	}

	return string(b), nil
}

func IsColumnNullable(nullable string) bool {
	return nullable == "YES"
}

func IsAutoIncrement(extra string) bool {
	return containsIgnoreCase(extra, "auto_increment")
}

func containsIgnoreCase(s, substr string) bool {
	return indexIgnoreCase(s, substr) != -1
}

func FormatDSN(host, port, user, password, dbname string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, dbname)
}
