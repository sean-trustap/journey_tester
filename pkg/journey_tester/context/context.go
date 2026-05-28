// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package context

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	journey_tester_report "github.com/trustap/journey_tester/pkg/journey_tester/report"
	"gopkg.in/yaml.v3"
)

type Context struct {
	Conf     *Config
	Extracts *Extracts
}

type Config struct {
	values             map[string]string
	printer            journey_tester_report.Printer
	testEmailNamespace string
}

func (c *Config) Get(key string) string {
	if val, ok := c.values[key]; ok {
		return val
	}
	// Try to create value only if `key` not found in `local config`.
	if val, ok := c.createEmailAndPasswordConfigValues(key); ok {
		return val
	}

	c.printer.PrintError(fmt.Sprint("missing extracted value ", key))
	panic(fmt.Sprintf("missing extracted value for key: %v", key))
}

func (c *Config) GetInt(key string) int {
	str := c.Get(key)
	i, err := strconv.ParseFloat(str, 64)
	if err != nil {
		c.printer.PrintError(fmt.Sprint("missing extracted value ", key))
		panic(fmt.Sprintf("failed to get int value key: %v value: %v", key, str))
	}
	return int(i)
}

func (c *Config) GetDuration(key string) time.Duration {
	val := c.Get(key)
	d, err := time.ParseDuration(val)
	if err != nil {
		c.printer.PrintError(fmt.Sprintf("error parsing duration key: %s value: %s", key, val))
		panic(fmt.Sprintf("failed to get duration value key: %s value: %s", key, val))
	}
	return d
}

// `createEmailAndPasswordConfigValues` returns a new email or password corresponding to
// a `key` value of `user.<key>.email` and `user.<key>.password`.
func (c *Config) createEmailAndPasswordConfigValues(key string) (string, bool) {
	split := strings.Split(key, ".")
	ln := len(split)

	if ln == 3 && split[0] == "users" && split[2] == "email" {
		return split[1] + "." + c.testEmailNamespace, true
	}
	if ln == 3 && split[0] == "users" && split[2] == "password" {
		return "password", true
	}

	return "", false
}

func newConfig(values map[string]string, printer journey_tester_report.Printer, testEmailNamespace string) *Config {
	return &Config{
		values:             values,
		printer:            printer,
		testEmailNamespace: testEmailNamespace,
	}
}

func (c *Context) Copy() *Context {
	newContext := &Context{
		Extracts: CopyExt(c.Extracts),
		Conf:     c.Conf,
	}
	return newContext
}

func NewContext(
	configFile string,
	sampleConfigFile string,
	printer journey_tester_report.Printer,
) (*Context, error) {
	expValues, err := readConfig(sampleConfigFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read sample config: %w", err)
	}

	values, err := readConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read config: %w", err)
	}

	for k := range expValues {
		if _, ok := values[k]; !ok {
			return nil, fmt.Errorf("local config doesn't contain '%s'", k)
		}
	}

	// Verifying that the local configuration file doesn't contain any
	// variables that aren't in the sample helps to avoid using such
	// variables in tests, and forgetting to add new variables back into the
	// sample.
	for k, v := range values {
		if _, ok := expValues[k]; !ok {
			return nil, fmt.Errorf("local config contains unknown value '%s'", k)
		}

		if v == "" {
			return nil, fmt.Errorf("local config contains empty value for '%s'", k)
		}
	}

	// Value of test_email_namespace is mandatory for createEmailAndPasswordConfigValues()
	// functionality.
	testEmailNamespace, ok := values["test_email_namespace"]
	if !ok {
		return nil, fmt.Errorf("local config doesn't contain '%s'", "test_email_namespace")
	}

	return &Context{
		Extracts: NewExtracts(),
		Conf:     newConfig(values, printer, testEmailNamespace),
	}, nil
}

func readConfig(path string) (map[string]string, error) {
	configYaml, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read YAML config: %w", err)
	}

	nestedConfig := map[string]any{}
	err = yaml.Unmarshal(configYaml, nestedConfig)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse YAML config: %w", err)
	}

	values, err := flattenMap(nestedConfig)
	if err != nil {
		return nil, fmt.Errorf("couldn't flatten YAML config: %w", err)
	}

	return values, nil
}

func flattenMap(in map[string]any) (map[string]string, error) {
	out := map[string]string{}
	for key, val := range in {
		switch subMap := val.(type) {
		case map[string]any:
			flatMap, err := flattenMap(subMap)
			if err != nil {
				var errInSubCall error
				if e := (&nilValueError{}); errors.As(err, &e) {
					errInSubCall = &nilValueError{key: key + "." + e.key}
				} else {
					errInSubCall = fmt.Errorf("unexpected error type: %w", err)
				}
				return nil, errInSubCall
			}

			for subKey, subVal := range flatMap {
				out[key+"."+subKey] = subVal
			}
		default:
			v := fmt.Sprintf("%v", val)
			out[key] = v
		}
	}
	return out, nil
}

type nilValueError struct {
	key string
}

func (e *nilValueError) Error() string {
	return fmt.Sprintf("value for key '%s' is `nil`", e.key)
}
