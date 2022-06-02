package goflat

import (
	"encoding/json"
	"errors"
	"strconv"
)

var ErrInvalidType = errors.New("not a valid JSON input")

// Convert a JSON string to a flat JSON string
// Only accepted inputs are: JSON objects or array of JSON objects
func Flat(jsonStr, prefix, separator string) (string, error) {
	var unflattenInterface interface{}
	err := json.Unmarshal([]byte(jsonStr), &unflattenInterface)
	if err != nil {
		return "", err
	}

	switch unflattenInterface.(type) {
	case []interface{}: // [{"a": 1}]
		var unflattenArrayMap []map[string]interface{}

		if err := json.Unmarshal([]byte(jsonStr), &unflattenArrayMap); err != nil {
			return "", err
		}
		return flattenArray(unflattenArrayMap, prefix, separator)
	case map[string]interface{}: // {"a": 1}
		var unflattenMap map[string]interface{}

		if err := json.Unmarshal([]byte(jsonStr), &unflattenMap); err != nil {
			return "", err
		}
		return flattenMap(unflattenMap, prefix, separator)
	default:
		return "", ErrInvalidType
	}
}

func flatten(unflat interface{}, prefix, separator string, top bool) (flatMap map[string]interface{}) {
	flatMap = make(map[string]interface{})

	switch unv := unflat.(type) {
	case map[string]interface{}:
		for key, value := range unv {
			for k, v := range flatten(value, key, separator, false) {
				newKey := prefix
				if top {
					newKey += k
				} else {
					newKey += separator + k
				}
				flatMap[newKey] = v
			}
		}
	case []interface{}:
		for i, val := range unv {
			for k, v := range flatten(val, prefix, separator, false) {
				flatMap[k+separator+strconv.Itoa(i)] = v
			}
		}
	default:
		flatMap[prefix] = unflat
	}

	return
}

func flattenArray(unflattenArrayMap []map[string]interface{}, prefix, separator string) (string, error) {
	var outputArrayMap []map[string]interface{}

	for _, unflattenMap := range unflattenArrayMap {
		flatMap := flatten(unflattenMap, prefix, separator, true)
		outputArrayMap = append(outputArrayMap, flatMap)
	}

	flatByteMap, err := json.Marshal(&outputArrayMap)
	if err != nil {
		return "", err
	}
	return string(flatByteMap), nil
}

func flattenMap(unflattenMap map[string]interface{}, prefix, separator string) (string, error) {
	flatMap := flatten(unflattenMap, prefix, separator, true)

	flatByteMap, err := json.Marshal(&flatMap)
	if err != nil {
		return "", err
	}
	return string(flatByteMap), nil
}
