// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package runner

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

func GetStringValueAtPath(path []string, data any) (string, error) {
	val := reflect.Indirect(reflect.ValueOf(data))
	switch val.Kind() {
	case reflect.Map:
		if len(path) == 0 {
			return "", fmt.Errorf("path is empty but current object is map and needs a key")
		}
		key := reflect.ValueOf(path[0]).String()
		for _, e := range val.MapKeys() {
			if e.String() == key {
				v := val.MapIndex(e)
				return GetStringValueAtPath(path[1:], v.Interface())
			}
		}
	case reflect.Slice:
		if len(path) == 0 {
			return "", fmt.Errorf("path is empty but current object is an array and needs an index")
		}
		index := reflect.ValueOf(path[0]).String()
		for i := 0; i < val.Len(); i++ {
			if index == strconv.Itoa(i) {
				return GetStringValueAtPath(path[1:], val.Index(i).Interface())
			}
		}
	case reflect.Array:
		return "", fmt.Errorf("array not supported")
	case reflect.Bool:
		if len(path) == 0 {
			return strconv.FormatBool(val.Bool()), nil
		}
		return "", fmt.Errorf("path hits raw type not struct")
	case reflect.String:
		if len(path) == 0 {
			return val.String(), nil
		}
		return "", fmt.Errorf("path hits raw type not struct")
	case reflect.Int:
		if len(path) == 0 {
			return strconv.Itoa(int(val.Int())), nil
		}
		return "", fmt.Errorf("path hits raw type not struct")
	case reflect.Float64:
		if len(path) == 0 {
			if isFloatInt(val.Float()) {
				return fmt.Sprintf("%d", int(val.Float())), nil
			}
			return fmt.Sprintf("%f", val.Float()), nil
		}
		return "", fmt.Errorf("path hits raw type not struct")
	case reflect.Float32:
		if len(path) == 0 {
			return fmt.Sprintf("%f", val.Float()), nil
		}
		return "", fmt.Errorf("path hits raw type not struct")
	default:
		return "", fmt.Errorf("error, path: %v, kind is %v", path, val.Kind().String())
	}
	return "", fmt.Errorf("error, path: %v", path)
}

func isFloatInt(floatValue float64) bool {
	return math.Mod(floatValue, 1.0) == 0
}
