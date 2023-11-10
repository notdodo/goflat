package goflat

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	json "github.com/ohler55/ojg/oj"
)

var ErrInvalidType = errors.New("not a valid JSON input")

// FlattenerConfig holds configuration options for flattening.
type FlattenerConfig struct {
	Prefix    string
	Separator string
	OmitEmpty bool
	OmitNil   bool
	SortKeys  bool
}

// DefaultFlattenerConfig returns a FlattenerConfig with default values.
func defaultConfiguration(config ...FlattenerConfig) FlattenerConfig {
	return FlattenerConfig{
		Prefix:    "",
		Separator: ".",
		OmitEmpty: false,
		OmitNil:   false,
		SortKeys:  false,
	}
}

func FlatStruct(input interface{}, config ...FlattenerConfig) map[string]interface{} {
	cfg := defaultConfiguration()
	if len(config) > 0 {
		cfg = config[0]
	}

	result := make(map[string]interface{})
	flattenFields(reflect.ValueOf(input), cfg.Prefix, result, cfg)
	if cfg.SortKeys {
		keys := make(map[string]struct{})
		for key := range result {
			keys[key] = struct{}{}
		}
		return sortKeysAndReturnResult(result, keys)
	}
	return result
}

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
	flatten(cfg.Separator, cfg.Prefix, data, flattenedMap, cfg)
	if cfg.SortKeys {
		keys := make(map[string]struct{})
		for key := range flattenedMap {
			keys[key] = struct{}{}
		}
		flattenedMap = sortKeysAndReturnResult(flattenedMap, keys)
	}

	flattenedJSON, err := json.Marshal(flattenedMap)
	if err != nil {
		return "", ErrInvalidType
	}
	return string(flattenedJSON), nil
}

func sortKeysAndReturnResult(result map[string]interface{}, keys map[string]struct{}) map[string]interface{} {
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

func flatten(separator, prefix string, value interface{}, result map[string]interface{}, config FlattenerConfig) {
	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + config.Separator + key
			}
			flatten(config.Separator, fullKey, val, result, config)
		}
	case []interface{}:
		flattenArray(prefix, v, result, config)
	default:
		if !(config.OmitEmpty && isEmptyValue(reflect.ValueOf(v))) && !(config.OmitNil && isNilValue(reflect.ValueOf(v))) {
			result[prefix] = v
		}
	}
}

func flattenArray(prefix string, arr []interface{}, result map[string]interface{}, config FlattenerConfig) {
	for i, v := range arr {
		fullKey := fmt.Sprintf("%s%s%s%d", config.Prefix, prefix, config.Separator, i)
		if strings.Index(fullKey, config.Separator) == 0+len(config.Prefix) {
			fullKey = fullKey[1:]
		}
		flatten("", fullKey, v, result, config)
	}
}

func flattenFields(val reflect.Value, prefix string, result map[string]interface{}, config FlattenerConfig) {
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name
		fieldValue := field.Interface()

		if !(config.OmitEmpty && isEmptyValue(field)) && !(config.OmitNil && isNilValue(field)) {
			if field.Kind() == reflect.Struct {
				flattenFields(field, prefix+fieldName+config.Separator, result, config)
			} else if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
				flattenArrayField(prefix, fieldName, field, result, config)
			} else {
				result[prefix+fieldName] = fieldValue
			}
		}
	}
}

func flattenArrayField(prefix, fieldName string, field reflect.Value, result map[string]interface{}, config FlattenerConfig) {
	for i := 0; i < field.Len(); i++ {
		item := field.Index(i).Interface()
		key := fmt.Sprintf("%s%s%d", prefix+fieldName+config.Separator, config.Separator, i)
		if !(config.OmitEmpty && isEmptyValue(reflect.ValueOf(item))) && !(config.OmitNil && isNilValue(reflect.ValueOf(item))) {
			result[key] = item
		}
	}
}

func isEmptyValue(field reflect.Value) bool {
	zero := reflect.Zero(field.Type())
	return reflect.DeepEqual(field.Interface(), zero.Interface())
}

func isNilValue(field reflect.Value) bool {
	return field.Kind() == reflect.Ptr && field.IsNil()
}
