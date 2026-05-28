// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package runner_test

import (
	"testing"

	"github.com/trustap/journey_tester/pkg/journey_tester/runner"
)

func TestValidateBasic(t *testing.T) {
	act := map[string]any{
		"status": "foo",
	}
	exp := map[string]any{
		"status": "foo",
	}
	err := runner.Validate(false, exp, act)
	if err != nil {
		t.Errorf("validate failed: %v", err)
	}
}

func TestValidateEmptyExp(t *testing.T) {
	act := map[string]any{
		"status": "foo",
	}
	exp := map[string]any{}
	err := runner.Validate(true, exp, act)
	if err != nil {
		t.Errorf("validate failed: %v", err)
	}
}

func TestValidateWrongFieldName(t *testing.T) {
	act := map[string]any{
		"status": "foo",
	}
	exp := map[string]any{
		"state": "foo",
	}
	err := runner.Validate(false, exp, act)
	if err == nil {
		t.Errorf("validate should have failed but didn't")
	}
}

func TestValidateWrongFieldType(t *testing.T) {
	act := map[string]any{
		"status": "4",
	}
	exp := map[string]any{
		"status": 4,
	}
	err := runner.Validate(false, exp, act)
	if err == nil {
		t.Errorf("validate should have failed but didn't")
	}
}

func TestValidateMissingSchemaField(t *testing.T) {
	act := map[string]any{}
	exp := map[string]any{
		"status": "4",
	}
	err := runner.Validate(false, exp, act)
	if err == nil {
		t.Errorf("validate should have failed but didn't")
	}
}

func TestValidateWrongValue(t *testing.T) {
	act := map[string]any{
		"status": "foo",
	}
	exp := map[string]any{
		"status": "bar",
	}
	err := runner.Validate(false, exp, act)
	if err == nil {
		t.Errorf("validate should have failed but didn't")
	}
}

func TestValidateAllPrimitiveJSONTypes(t *testing.T) {
	act := map[string]any{
		"status":  "foo",
		"id":      3.0,
		"ishappy": true,
	}
	exp := map[string]any{
		"status":  "foo",
		"id":      3,
		"ishappy": true,
	}
	err := runner.Validate(false, exp, act)
	if err != nil {
		t.Errorf("validate failed: %v", err)
	}
}

func TestValidateAllPrimitiveJSONTypesWrongValueFails(t *testing.T) {
	act := map[string]any{
		"status":  "foo",
		"id":      3.0,
		"ishappy": true,
	}
	exp := map[string]any{
		"status":  "foo",
		"id":      4,
		"ishappy": true,
	}
	err := runner.Validate(false, exp, act)
	if err == nil {
		t.Errorf("validate should have failed but didn't")
	}
}

func TestValidateNestedSuccess(t *testing.T) {
	act := map[string]any{
		"outer": "level",
		"status": map[string]any{
			"current": "new",
			"inner":   "part",
		},
	}
	exp := map[string]any{
		"outer": "level",
		"status": map[string]any{
			"current": "new",
			"inner":   "part",
		},
	}
	err := runner.Validate(false, exp, act)
	if err != nil {
		t.Errorf("validate failed: %v", err)
	}
}

func TestValidateNestedFailureValue(t *testing.T) {
	act := map[string]any{
		"outer": "level",
		"status": map[string]any{
			"current": "new",
			"inner":   "part",
		},
	}
	exp := map[string]any{
		"outer": "level",
		"status": map[string]any{
			"current": "new",
			"inner":   "bit",
		},
	}
	err := runner.Validate(false, exp, act)
	if err == nil {
		t.Errorf("validate should have failed but didn't")
	}
}

func TestValidateNestedFailureMissing(t *testing.T) {
	act := map[string]any{
		"outer": "level",
		"status": map[string]any{
			"current": "new",
		},
	}
	exp := map[string]any{
		"outer": "level",
		"status": map[string]any{
			"current": "new",
			"inner":   "bit",
		},
	}
	err := runner.Validate(false, exp, act)
	if err == nil {
		t.Errorf("validate should have failed but didn't")
	}
}

func TestValidateNestedSuccessExtra(t *testing.T) {
	act := map[string]any{
		"outer": "level",
		"status": map[string]any{
			"current": "new",
			"inner":   "bit",
		},
	}
	exp := map[string]any{
		"outer": "level",
		"status": map[string]any{
			"current": "new",
		},
	}
	err := runner.Validate(true, exp, act)
	if err != nil {
		t.Errorf("validate failed: %v", err)
	}
}

func TestValidateNestedFailureMissingMap(t *testing.T) {
	act := map[string]any{
		"outer": "level",
	}
	exp := map[string]any{
		"outer": "level",
		"status": map[string]any{
			"current": "new",
		},
	}
	err := runner.Validate(false, exp, act)
	if err == nil {
		t.Errorf("validate should have failed but didn't")
	}
}

func TestValidateFloatJSONKind(t *testing.T) {
	act := map[string]any{
		"status": 4.0,
	}
	exp := map[string]any{
		"status": 4,
	}
	err := runner.Validate(false, exp, act)
	if err != nil {
		t.Errorf("validate failed: %v", err)
	}
}

func TestValidateExpectFloat(t *testing.T) {
	act := map[string]any{
		"status": 4.0,
	}
	exp := map[string]any{
		"status": 4.0,
	}
	err := runner.Validate(false, exp, act)
	if err != nil {
		t.Errorf("validate failed: %v", err)
	}
}
