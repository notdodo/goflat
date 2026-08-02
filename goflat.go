package goflat

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

var ErrInvalidType = errors.New("not a valid JSON input")

// FlattenerConfig holds configuration options for flattening.
type FlattenerConfig struct {
	Prefix      string
	Separator   string
	OmitEmpty   bool
	OmitNil     bool
	KeysToLower bool
}

func defaultConfiguration() FlattenerConfig {
	return FlattenerConfig{
		Prefix:      "",
		Separator:   ".",
		OmitEmpty:   true,
		OmitNil:     true,
		KeysToLower: false,
	}
}

func resolveConfig(config []FlattenerConfig) FlattenerConfig {
	if len(config) > 0 {
		return config[0]
	}
	return defaultConfiguration()
}

// shouldInclude checks whether a value should be kept based on OmitEmpty/OmitNil config.
func shouldInclude(val reflect.Value, config FlattenerConfig) bool {
	return (!config.OmitEmpty || !isEmptyValue(val)) && (!config.OmitNil || !isNilValue(val))
}

// buildKey joins prefix and name with the configured separator.
func buildKey(prefix, name, separator string) string {
	if prefix == "" {
		return name
	}
	return prefix + separator + name
}

// indexKey joins prefix with an array index.
func indexKey(prefix, separator string, i int) string {
	if prefix == "" {
		return strconv.Itoa(i)
	}
	return prefix + separator + strconv.Itoa(i)
}

// FlatStruct flattens a Go struct into a map with flattened keys.
func FlatStruct(input any, config ...FlattenerConfig) map[string]any {
	cfg := resolveConfig(config)
	result := make(map[string]any)
	flattenFields(reflect.ValueOf(input), cfg.Prefix, result, cfg)
	if cfg.KeysToLower {
		keysToLower(&result)
	}
	return result
}

// FlatJSON flattens a JSON string into a flattened JSON string.
func FlatJSON(jsonStr string, config ...FlattenerConfig) (string, error) {
	flattenedMap, err := FlatJSONToMap(jsonStr, config...)
	if err != nil {
		return "", err
	}
	flattenedJSON, err := json.Marshal(flattenedMap)
	if err != nil {
		return "", ErrInvalidType
	}
	return string(flattenedJSON), nil
}

// FlatJSONToMap flattens a JSON string into a map with flattened keys.
func FlatJSONToMap(jsonStr string, config ...FlattenerConfig) (map[string]any, error) {
	cfg := resolveConfig(config)

	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, ErrInvalidType
	}

	flattenedMap := make(map[string]any)
	flatten(cfg.Prefix, data, flattenedMap, cfg)
	if cfg.KeysToLower {
		keysToLower(&flattenedMap)
	}
	return flattenedMap, nil
}

// flatten flattens a nested JSON structure into a map with flattened keys.
func flatten(prefix string, value any, result map[string]any, config FlattenerConfig) {
	switch v := value.(type) {
	case map[string]any:
		for key, val := range v {
			flatten(buildKey(prefix, key, config.Separator), val, result, config)
		}
	case []any:
		flattenArray(prefix, v, result, config)
	default:
		if v == nil && (config.OmitNil || config.OmitEmpty) {
			return
		}
		if v != nil {
			val := reflect.ValueOf(v)
			if !shouldInclude(val, config) {
				return
			}
		}
		result[prefix] = v
	}
}

// flattenArray flattens a JSON array into a map with flattened keys.
func flattenArray(prefix string, arr []any, result map[string]any, config FlattenerConfig) {
	for i, v := range arr {
		flatten(indexKey(prefix, config.Separator, i), v, result, config)
	}
}

// resolveValue unwraps pointer and interface wrappers to get the underlying value.
func resolveValue(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return val
		}
		val = val.Elem()
	}
	return val
}

// isResolvedNil returns true if a resolved value is invalid or nil.
func isResolvedNil(resolved reflect.Value) bool {
	return !resolved.IsValid() || isNilValue(resolved)
}

// structFieldName returns the json tag name if present, otherwise the Go field name.
// Returns empty string for fields with json:"-" to signal they should be skipped.
func structFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name
	}
	if tag == "-" {
		return ""
	}
	if idx := strings.Index(tag, ","); idx != -1 {
		tag = tag[:idx]
	}
	if tag == "" {
		return f.Name
	}
	return tag
}

// flattenFields flattens fields of a struct into a map with flattened keys.
func flattenFields(val reflect.Value, prefix string, result map[string]any, config FlattenerConfig) {
	val = resolveValue(val)
	if !val.IsValid() {
		return
	}

	switch val.Kind() {
	case reflect.Struct:
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldName := structFieldName(typ.Field(i))
			if fieldName == "" {
				continue // json:"-"
			}
			resolved := resolveValue(field)
			if isResolvedNil(resolved) {
				if config.OmitNil {
					continue
				}
			} else if resolved.Kind() == reflect.Slice || resolved.Kind() == reflect.Array {
				flattenArrayFields(prefix+fieldName, resolved, result, config)
				continue
			}
			if shouldInclude(field, config) {
				flattenFields(field, prefix+fieldName+config.Separator, result, config)
			}
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			field := val.MapIndex(key)
			fieldName := key.String()
			fullKey := prefix + fieldName
			resolved := resolveValue(field)
			if isResolvedNil(resolved) {
				if !config.OmitNil {
					result[fullKey] = nil
				}
				continue
			}
			if !shouldInclude(resolved, config) {
				continue
			}
			switch resolved.Kind() {
			case reflect.Struct, reflect.Map:
				flattenFields(resolved, fullKey+config.Separator, result, config)
			case reflect.Slice, reflect.Array:
				flattenArrayFields(fullKey, resolved, result, config)
			default:
				result[fullKey] = resolved.Interface()
			}
		}
	default:
		if !shouldInclude(val, config) || !val.CanInterface() {
			return
		}
		// Strip trailing separator from prefix
		if len(prefix) >= len(config.Separator) && strings.HasSuffix(prefix, config.Separator) {
			prefix = prefix[:len(prefix)-len(config.Separator)]
		}
		// If val is a JSON string, recursively flatten it
		if val.Kind() == reflect.String {
			if data := tryParseJSON(val.String()); data != nil {
				flatten(prefix, data, result, config)
				return
			}
		}
		result[prefix] = val.Interface()
	}
}

// flattenArrayFields flattens fields of a slice/array into a map with flattened keys.
func flattenArrayFields(base string, field reflect.Value, result map[string]any, config FlattenerConfig) {
	for i := 0; i < field.Len(); i++ {
		resolved := resolveValue(field.Index(i))
		if isResolvedNil(resolved) {
			if !config.OmitNil {
				result[indexKey(base, config.Separator, i)] = nil
			}
			continue
		}
		switch resolved.Kind() {
		case reflect.Struct, reflect.Map:
			flattenFields(resolved, indexKey(base, config.Separator, i)+config.Separator, result, config)
		default:
			if shouldInclude(resolved, config) {
				result[indexKey(base, config.Separator, i)] = resolved.Interface()
			}
		}
	}
}

func keysToLower(result *map[string]any) {
	newResult := make(map[string]any, len(*result))
	for k, v := range *result {
		newResult[strings.ToLower(k)] = v
	}
	*result = newResult
}

// isEmptyValue checks if a reflect.Value is empty.
func isEmptyValue(field reflect.Value) bool {
	if field.IsValid() && field.Type().Kind() == reflect.Bool {
		return false
	}
	if !field.IsValid() || !field.CanInterface() {
		return true
	}
	return field.IsZero()
}

// isNilValue checks if a reflect.Value is nil.
func isNilValue(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return field.IsNil()
	default:
		return false
	}
}

// tryParseJSON attempts to parse a string as JSON, returning the parsed value or nil.
// Only attempts parsing if the string looks like it could be JSON (starts with { or [).
func tryParseJSON(str string) any {
	trimmed := strings.TrimSpace(str)
	if len(trimmed) < 2 {
		return nil
	}
	if trimmed[0] != '{' && trimmed[0] != '[' {
		return nil
	}
	var data any
	if err := json.Unmarshal([]byte(str), &data); err != nil {
		return nil
	}
	return data
}
