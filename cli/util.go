package cli

import (
	"errors"
	"reflect"
	"strings"
)

// errNeedQuery is returned when `jobs` is invoked with neither keywords nor a
// location.
var errNeedQuery = errors.New("provide search keywords or a --location")

// joinArgs joins positional arguments into a single query string.
func joinArgs(args []string) string { return strings.TrimSpace(strings.Join(args, " ")) }

// sliceLen returns the length of a slice value, or 1 for a single record.
func sliceLen(records any) int {
	rv := reflect.ValueOf(records)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Slice {
		return rv.Len()
	}
	if !rv.IsValid() {
		return 0
	}
	return 1
}
