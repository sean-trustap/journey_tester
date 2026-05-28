// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package context

import (
	"fmt"
	"strconv"
)

// Extracts holds extracted state. It should be used during a journey to get and set values between
// steps. It also holds state that is setup by the framework, such as the admin token.
type Extracts struct {
	values map[string]string
}

func NewExtracts() *Extracts {
	values := map[string]string{}
	return &Extracts{
		values: values,
	}
}

func CopyExt(original *Extracts) *Extracts {
	newExt := NewExtracts()
	for k, v := range original.values {
		newExt.values[k] = v
	}
	return newExt
}

func (c *Extracts) Set(key string, val string) {
	c.values[key] = val
}

func (c *Extracts) Get(key string) string {
	v, ok := c.values[key]
	if !ok {
		// We don't want the tests to check 'ok'. If the key is not set, it's a fundamental
		// problem with the journey and the journey should be aborted.
		panic(fmt.Sprintf("missing extracted value for key: %v", key))
	}
	return v
}

func (c *Extracts) GetInt(key string) int {
	str := c.Get(key)
	i, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to get int value key: %v value: %v", key, str))
	}
	return int(i)
}

func (c *Extracts) GetBool(key string) bool {
	str := c.Get(key)
	b, err := strconv.ParseBool(str)
	if err != nil {
		panic(fmt.Sprintf("failed to get int value key: %v value: %v", key, str))
	}
	return b
}

func (c *Extracts) Has(key string) bool {
	_, ok := c.values[key]
	return ok
}
