// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package runner_test

import (
	"testing"

	"github.com/trustap/journey_tester/pkg/journey_tester/runner"
)

func TestGetMapStringValue(t *testing.T) {
	p := []string{"id"}
	d := &map[string]any{
		"id": "dave",
	}
	expected := "dave"
	actual, err := runner.GetStringValueAtPath(p, d)
	if err != nil {
		t.Errorf("error extracting value: %v", err)
	}
	if actual != expected {
		t.Errorf("failed to extract value, expected: %v, actual: %v", expected, actual)
	}
}

func TestGetSliceIndexValue(t *testing.T) {
	p := []string{"1"}
	d := &[]string{
		"id",
		"dave",
		"bar",
	}
	expected := "dave"
	actual, err := runner.GetStringValueAtPath(p, d)
	if err != nil {
		t.Errorf("error extracting value: %v", err)
	}
	if actual != expected {
		t.Errorf("failed to extract value, expected: %v, actual: %v", expected, actual)
	}
}

func TestGetNestedValue(t *testing.T) {
	p := []string{"user", "id"}
	d := &map[string]any{
		"user": map[string]any{
			"id": "dave",
		},
	}
	expected := "dave"
	actual, err := runner.GetStringValueAtPath(p, d)
	if err != nil {
		t.Errorf("error extracting value: %v", err)
	}
	if actual != expected {
		t.Errorf("failed to extract value, expected: %v, actual: %v", expected, actual)
	}
}

func TestGetStringValue(t *testing.T) {
	var p []string
	d := "dave"
	expected := "dave"
	actual, err := runner.GetStringValueAtPath(p, d)
	if err != nil {
		t.Errorf("error extracting value: %v", err)
	}
	if actual != expected {
		t.Errorf("failed to extract value, expected: %v, actual: %v", expected, actual)
	}
}

func TestGetIntValue(t *testing.T) {
	var p []string
	d := 4
	expected := "4"
	actual, err := runner.GetStringValueAtPath(p, d)
	if err != nil {
		t.Errorf("error extracting value: %v", err)
	}
	if actual != expected {
		t.Errorf("failed to extract value, expected: %v, actual: %v", expected, actual)
	}
}

func TestGetFloatValue(t *testing.T) {
	var p []string
	d := 4.2
	expected := "4.200000"
	actual, err := runner.GetStringValueAtPath(p, d)
	if err != nil {
		t.Errorf("error extracting value: %v", err)
	}
	if actual != expected {
		t.Errorf("failed to extract value, expected: %v, actual: %v", expected, actual)
	}
}

func TestNestedMixedTypes(t *testing.T) {
	p := []string{"id", "1"}
	d := &map[string]any{
		"id": &[]string{
			"id",
			"dave",
			"bar",
		},
	}
	expected := "dave"
	actual, err := runner.GetStringValueAtPath(p, d)
	if err != nil {
		t.Errorf("error extracting value: %v", err)
	}
	if actual != expected {
		t.Errorf("failed to extract value, expected: %v, actual: %v", expected, actual)
	}
}
