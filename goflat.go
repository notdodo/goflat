package goflat

import (
	"errors"
	"strconv"

	json "github.com/ohler55/ojg/oj"
)

var ErrInvalidType = errors.New("not a valid JSON input")

// Convert a JSON struct to a flat Map
func FlatStruct(str any, prefix, separator string) (flattenMap map[string]interface{}, err error) {
	jsonBytes, err := json.Marshal(str)
	if err != nil {
		return nil, err
	}

	flattenStr, err := Flat(string(jsonBytes), prefix, separator)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(flattenStr), &flattenMap)
	return flattenMap, nil
}

// Convert a JSON string to a flat JSON string
// Only accepted inputs are: JSON objects or array of JSON objects
func Flat(jsonStr, prefix, separator string) (string, error) {
	unflattenInterface, err := json.ParseString(jsonStr)
	if err != nil {
		return "", err
	}

	switch v := unflattenInterface.(type) {
	case []interface{}: // [{"a": 1}]
		var flattenArrayMap []map[string]interface{}
		for _, element := range v {
			flattenArrayMap = append(flattenArrayMap, flattenMap(element.(map[string]interface{}), prefix, separator))
		}
		return json.JSON(flattenArrayMap), nil
	case map[string]interface{}: // {"a": 1}
		flattenArrayMap := flattenMap(v, prefix, separator)
		return json.JSON(flattenArrayMap), nil
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

	return json.JSON(&outputArrayMap), nil
}

func flattenMap(unflattenMap map[string]interface{}, prefix, separator string) map[string]interface{} {
	return flatten(unflattenMap, prefix, separator, true)
}
