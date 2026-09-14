// Package helper provides the internal utilities shared by the API modules.
//
// It mirrors aiotieba.helper.
package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PackJSON serializes obj into a compact JSON string, mirroring pack_json.
func PackJSON(obj any) string {
	b, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(b)
}

// ParseJSON decodes data into v, mirroring parse_json.
func ParseJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// ParseJSONMap decodes a JSON object.
//
// Numbers are kept as json.Number so that large identifiers (user ids, fids)
// keep full precision; the Python client relies on Python's arbitrary precision
// integers here.
func ParseJSONMap(data []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("helper: decoding the JSON response: %w", err)
	}
	if out == nil {
		return nil, fmt.Errorf("helper: the JSON response is not an object")
	}
	return out, nil
}

// JSONInt returns m[key] as an int64. Missing keys yield 0, mirroring the
// Python int(m[key]) conversion used on numeric fields.
func JSONInt(m map[string]any, key string) int64 {
	v, ok := m[key]
	if !ok {
		return 0
	}
	return toInt64(v)
}

// JSONStr returns m[key] as a string.
func JSONStr(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// JSONBool returns m[key] as a bool.
func JSONBool(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != "" && t != "0" && t != "false"
	default:
		return toInt64(t) != 0
	}
}

// JSONMap returns m[key] as a nested object, or nil when it is missing or has
// another type.
func JSONMap(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	nested, _ := v.(map[string]any)
	return nested
}

// JSONSlice returns m[key] as a list, or nil when it is missing or has another
// type.
func JSONSlice(m map[string]any, key string) []any {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	list, _ := v.([]any)
	return list
}

// HasJSONKey reports whether the key exists, mirroring `"key" in data_map`.
func HasJSONKey(m map[string]any, key string) bool {
	_, ok := m[key]
	return ok
}

func toInt64(v any) int64 {
	switch t := v.(type) {
	case nil:
		return 0
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			f, ferr := t.Float64()
			if ferr != nil {
				return 0
			}
			return int64(f)
		}
		return n
	case float64:
		return int64(t)
	case float32:
		return int64(t)
	case int:
		return int64(t)
	case int64:
		return t
	case int32:
		return int64(t)
	case bool:
		if t {
			return 1
		}
		return 0
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}

// JSONRaw returns the raw decoded value of m[key].
//
// It is needed where the Python code compares JSON scalars with `==`, which is
// type sensitive: the string "123" never equals the number 123.
func JSONRaw(m map[string]any, key string) (any, bool) {
	v, ok := m[key]
	return v, ok
}

// JSONScalarEqual reports whether a and b are the same JSON scalar, mirroring
// Python's `==` on values decoded by json.loads. Composite values are never
// equal.
func JSONScalarEqual(a, b any) bool {
	switch a.(type) {
	case nil, bool, string, json.Number, float64, float32, int, int64, int32:
	default:
		return false
	}
	return a == b
}

// BoolInt converts b to 1 or 0, mirroring the `int(is_x)` conversions used by
// the Python API modules.
func BoolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// AnyInt64 converts an arbitrary decoded JSON value to an int64.
func AnyInt64(v any) int64 { return toInt64(v) }

// AnyFloat64 converts an arbitrary decoded JSON value to a float64.
func AnyFloat64(v any) float64 {
	switch t := v.(type) {
	case nil:
		return 0
	case json.Number:
		f, err := t.Float64()
		if err != nil {
			return 0
		}
		return f
	case float64:
		return t
	case float32:
		return float64(t)
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return float64(toInt64(t))
	}
}

// IsPortrait reports whether portrait matches the portrait format, mirroring
// is_portrait.
func IsPortrait(portrait any) bool {
	s, ok := portrait.(string)
	return ok && strings.HasPrefix(s, "tb.")
}

// IsUserName reports whether userName matches the user name format, mirroring
// is_user_name.
func IsUserName(userName any) bool {
	s, ok := userName.(string)
	return ok && !strings.HasPrefix(s, "tb.")
}

// DefaultDatetime returns the default datetime used by the API modules,
// mirroring default_datetime.
func DefaultDatetime() time.Time {
	return time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
}

// Timeout derives a context that is cancelled after delay, mirroring the
// timeout helper of the Python client.
func Timeout(parent context.Context, delay time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, delay)
}
