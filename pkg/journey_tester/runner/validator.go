// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package runner

import (
	"fmt"
	"strconv"
)

// Validate returns `nil` if all of the entries in exp are found in act, and the
// type and values match.
func Validate(allowExtraProperties bool, expectedBody, jsonBody map[string]any) error {
	return mapsMatch(allowExtraProperties, expectedBody, jsonBody)
}

func mapsMatch(allowExtraProperties bool, expectedBody, jsonBody map[string]any) error {
	if !allowExtraProperties {
		for key := range jsonBody {
			_, exists := expectedBody[key]
			if !exists {
				return fmt.Errorf("found unexpected property '%s' in response body", key)
			}
		}
	}

	for key, exp := range expectedBody {
		act, ok := jsonBody[key]
		if !ok {
			return fmt.Errorf("expected field '%s' wasn't found in object: %v", key, jsonBody)
		}

		if exp != nil {
			err := valuesMatch(allowExtraProperties, exp, act, key)
			if err != nil {
				return fmt.Errorf("'%s' didn't match: %w", key, err)
			}
		}
	}
	return nil
}

type ExpType string

var (
	Null    ExpType = "null"
	NotNull ExpType = "not_null"
)

func valuesMatch(allowExtraProperties bool, exp, act any, identifier string) error {
	switch exp {
	case Null:
		if act != nil {
			return fmt.Errorf("'%s' type: expected 'null', got '%T'", identifier, act)
		}
		return nil
	case NotNull:
		if act == nil {
			return fmt.Errorf("'%s' type: unexpected 'null'", identifier)
		}
		return nil
	}

	switch typedExp := exp.(type) {
	case nil:
		return nil
	case string:
		typedAct, ok := act.(string)
		if !ok {
			return fmt.Errorf("'%s' type: expected 'string', got '%T'", identifier, act)
		}
		if typedExp != typedAct {
			return fmt.Errorf("'%s': expected '%s', got '%s'", identifier, typedExp, typedAct)
		}
	case int:
		typedAct, ok := act.(float64)
		if !ok {
			return fmt.Errorf("'%s' type: expected 'float64', got '%T'", identifier, act)
		}
		expFloat64 := float64(typedExp)
		if expFloat64 != typedAct {
			return fmt.Errorf("'%s': expected %f, got %f", identifier, expFloat64, typedAct)
		}
	case float64:
		typedAct, ok := act.(float64)
		if !ok {
			return fmt.Errorf("'%s' type: expected 'float64', got '%T'", identifier, act)
		}
		if typedExp != typedAct {
			return fmt.Errorf("'%s': expected %f, got %f", identifier, typedExp, typedAct)
		}
	case bool:
		typedAct, ok := act.(bool)
		if !ok {
			return fmt.Errorf("'%s' type: expected 'bool', got '%T'", identifier, act)
		}
		if typedExp != typedAct {
			return fmt.Errorf("'%s': expected '%v', got '%v'", identifier, typedExp, typedAct)
		}
	case map[string]any:
		typedAct, ok := act.(map[string]any)
		if !ok {
			return fmt.Errorf("'%s' type: expected 'map', got '%T'", identifier, act)
		}
		if err := mapsMatch(allowExtraProperties, typedExp, typedAct); err != nil {
			return fmt.Errorf("'%s': expected '%v', got '%v': %w", identifier, typedExp, typedAct, err)
		}
	case []any:
		typedAct, ok := act.([]any)
		if !ok {
			return fmt.Errorf("'%s' type: expected 'map', got '%T'", identifier, act)
		}
		if err := slicesMatch(allowExtraProperties, typedExp, typedAct); err != nil {
			return fmt.Errorf("'%s': expected '%v', got '%v': %w", identifier, typedExp, typedAct, err)
		}
	default:
		return fmt.Errorf("'%s' expectation has unsupported type: %T", identifier, exp)
	}
	return nil
}

func slicesMatch(allowExtraProperties bool, exps, acts []any) error {
	if len(exps) != len(acts) {
		return fmt.Errorf("length of slices do not match, expected: %v, got %v", len(exps), len(acts))
	}
	for i, exp := range exps {
		if exp != nil {
			err := valuesMatch(allowExtraProperties, exp, acts[i], strconv.Itoa(i))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
