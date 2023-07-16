// The reason that so many of these functions seem like duplication is simply for the better error
// messages so each driver doesn't need to deal with validation on these simple field types

package drivers

import (
	"os"
	"strings"
)

// Config is a struct with contains config values.
// Those values may not be used by all drivers. Each driver should validate the config values by itself.
type Config struct {
	User    string
	Pass    string
	Host    string
	Port    int
	DBName  string
	SSLMode string

	BlackList      []string
	WhiteList      []string
	Schema         string
	AddEnumTypes   bool
	EnumNullPrefix string

	ForeignKeys []ForeignKey

	// Concurrency defines amount of threads to use when loading tables info.
	Concurrency int

	// For mysql
	TinyIntAsInt bool
}

// DefaultInt retrieves a non-zero int or the default value provided.
func DefaultInt(value int, def int) int {
	if value == 0 {
		return def
	}

	return value
}

// DefaultEnv grabs a value from the environment or a default.
// This is shared by drivers to get config for testing.
func DefaultEnv(key, def string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		val = def
	}
	return val
}

// TablesFromList takes a whitelist or blacklist and returns
// the table names.
func TablesFromList(list []string) []string {
	if len(list) == 0 {
		return nil
	}

	var tables []string
	for _, i := range list {
		splits := strings.Split(i, ".")

		if len(splits) == 1 {
			tables = append(tables, splits[0])
		}
	}

	return tables
}

// ColumnsFromList takes a whitelist or blacklist and returns
// the columns for a given table.
func ColumnsFromList(list []string, tablename string) []string {
	if len(list) == 0 {
		return nil
	}

	var columns []string
	for _, i := range list {
		splits := strings.Split(i, ".")

		if len(splits) != 2 {
			continue
		}

		if splits[0] == tablename || splits[0] == "*" {
			columns = append(columns, splits[1])
		}
	}

	return columns
}
