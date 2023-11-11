package goflat

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

var ErrInvalidType = errors.New("not a valid JSON input")

// `FlattenerConfig` holds configuration options for flattening.
type FlattenerConfig struct {
	Prefix    string
	Separator string
	OmitEmpty bool
	OmitNil   bool
	SortKeys  bool
}

// `DefaultFlattenerConfig ` returns a FlattenerConfig with default values.
func defaultConfiguration(config ...FlattenerConfig) FlattenerConfig {
	return FlattenerConfig{
		Prefix:    "",
		Separator: ".",
		OmitEmpty: false,
		OmitNil:   false,
		SortKeys:  false,
	}
}

// `FlatStruct ` flattens a Go struct into a map with flattened keys.
func FlatStruct(input interface{}, config ...FlattenerConfig) map[string]interface{} {
	cfg := defaultConfiguration()
	if len(config) > 0 {
		cfg = config[0]
	}

	result := make(map[string]interface{})
	flattenFields(reflect.ValueOf(input), cfg.Prefix, result, cfg)
	if cfg.SortKeys {
		return sortKeysAndReturnResult(result)
	}
	return result
}

// `FlatJSON ` flattens a JSON string into a flattened JSON string.
func FlatJSON(jsonStr string, config ...FlattenerConfig) (string, error) {
	cfg := defaultConfiguration()
	if len(config) > 0 {
		cfg = config[0]
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", ErrInvalidType
	}

	flattenedMap := make(map[string]interface{})
	flatten(cfg.Prefix, data, flattenedMap, cfg)
	if cfg.SortKeys {
		flattenedMap = sortKeysAndReturnResult(flattenedMap)
	}

	flattenedJSON, err := json.Marshal(flattenedMap)
	if err != nil {
		return "", ErrInvalidType
	}
	return string(flattenedJSON), nil
}

// `FlatJSONToMap ` flattens a JSON string into a map with flattened keys.
func FlatJSONToMap(jsonStr string, config ...FlattenerConfig) (map[string]interface{}, error) {
	cfg := defaultConfiguration()
	if len(config) > 0 {
		cfg = config[0]
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return nil, ErrInvalidType
	}

	flattenedMap := make(map[string]interface{})
	flatten(cfg.Prefix, data, flattenedMap, cfg)
	if cfg.SortKeys {
		flattenedMap = sortKeysAndReturnResult(flattenedMap)
	}

	return flattenedMap, nil
}

// `sortKeysAndReturnResult ` sorts keys in the flattened structure.
func sortKeysAndReturnResult(result map[string]interface{}) map[string]interface{} {
	keys := make(map[string]string)
	for key := range result {
		keys[key] = ""
	}

	sortedResult := make(map[string]interface{})
	sortedKeys := make([]string, 0, len(keys))
	for key := range keys {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	for _, key := range sortedKeys {
		sortedResult[key] = result[key]
	}
	return sortedResult
}

// `flatten ` flattens a nested structure into a map with flattened keys.
func flatten(prefix string, value interface{}, result map[string]interface{}, config FlattenerConfig) {
	switch v := value.(type) {
	case map[string]interface{}:
		// For each key-value pair in the map, recursively flatten the nested structure.
		for key, val := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + config.Separator + key
			}
			flatten(fullKey, val, result, config)
		}
	case []interface{}:
		// For each element in the array, recursively flatten the nested structure.
		flattenArray(prefix, v, result, config)
	default:
		// If the value is neither a map nor an array, add it to the result map.
		// Optionally omitting empty or nil values based on the configuration.
		if !(config.OmitEmpty && isEmptyValue(reflect.ValueOf(v))) && !(config.OmitNil && isNilValue(reflect.ValueOf(v))) {
			result[prefix] = v
		}
	}
}

// `flattenArray ` flattens an array into a map with flattened keys.
func flattenArray(prefix string, arr []interface{}, result map[string]interface{}, config FlattenerConfig) {
	for i, v := range arr {
		// Generate the full key for each element in the array.
		fullKey := fmt.Sprintf("%s%s%s%d", config.Prefix, prefix, config.Separator, i)
		// Remove leading separator if it's present.
		if strings.Index(fullKey, config.Separator) == 0+len(config.Prefix) {
			fullKey = fullKey[1:]
		}
		// Recursively flatten the nested structure for each array element.
		flatten(fullKey, v, result, config)
	}
}

// `flattenFields ` flattens fields of a struct into a map with flattened keys.
func flattenFields(val reflect.Value, prefix string, result map[string]interface{}, config FlattenerConfig) {
	typ := val.Type()
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}

	switch val.Kind() {
	case reflect.Struct:
		// For each field in the struct, recursively flatten the nested structure.
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldName := typ.Field(i).Name
			// Optionally omitting empty or nil values based on the configuration.
			if !(config.OmitEmpty && isEmptyValue(field)) && !(config.OmitNil && isNilValue(field)) {
				// Recursively flatten the nested structure for each struct field.
				flattenFields(field, prefix+fieldName+config.Separator, result, config)
			}
		}
	case reflect.Map:
		// For each key-value pair in the map, recursively flatten the nested structure.
		for _, key := range val.MapKeys() {
			field := val.MapIndex(key)
			fieldName := key.String()
			fullKey := prefix + fieldName
			// Optionally omitting empty or nil values based on the configuration.
			if !(config.OmitEmpty && isEmptyValue(field)) && !(config.OmitNil && isNilValue(field)) {
				if field.Kind() == reflect.Struct {
					// If the value is a struct, recursively flatten the nested structure.
					flattenFields(field, fullKey+config.Separator, result, config)
				} else if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
					// If the value is a slice or array, flatten each element in the collection.
					flattenArrayFields(fullKey, "", field, result, config)
				} else {
					// If the value is neither a struct nor a slice/array, add it to the result map.
					result[fullKey] = field.Interface()
				}
			}
		}
	default:
		// If the value is neither a struct nor a map, add it to the result map.
		// Optionally omitting empty or nil values based on the configuration.
		if !(config.OmitEmpty && isEmptyValue(val)) && !(config.OmitNil && isNilValue(val)) {
			result[prefix[:len(prefix)-1]] = val.Interface()
		}
	}
}

// `flattenArrayFields ` flattens fields of an array into a map with flattened keys.
func flattenArrayFields(prefix, fieldName string, field reflect.Value, result map[string]interface{}, config FlattenerConfig) {
	for i := 0; i < field.Len(); i++ {
		// Extract each element from the array and generate a key for it.
		item := field.Index(i).Interface()
		key := fmt.Sprintf("%s%s%d", prefix+fieldName+config.Separator, config.Separator, i)
		// Optionally omitting empty or nil values based on the configuration.
		if !(config.OmitEmpty && isEmptyValue(reflect.ValueOf(item))) && !(config.OmitNil && isNilValue(reflect.ValueOf(item))) {
			// Add the key-value pair to the result map.
			result[key] = item
		}
	}
}

// `isEmptyValue ` checks if a reflect.Value is empty.
func isEmptyValue(field reflect.Value) bool {
	if !field.IsValid() || !field.CanInterface() {
		return true
	}
	// Check if the field is equal to its zero value.
	zero := reflect.Zero(field.Type())
	return reflect.DeepEqual(field.Interface(), zero.Interface())
}

// `isNilValue ` checks if a reflect.Value is nil.
func isNilValue(field reflect.Value) bool {
	// Check if the field is a pointer and is nil.
	return field.Kind() == reflect.Ptr && field.IsNil()
}
