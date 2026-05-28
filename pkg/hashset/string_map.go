// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package hashset

func NewStringMap[V any]() StringMap[V] {
	return StringMap[V]{data: make(map[string]V)}
}

type StringMap[V any] struct {
	data map[string]V
}

func StringMapFromMap[V any](vals map[string]V) StringMap[V] {
	hashmap := NewStringMap[V]()
	for key, vals := range vals {
		hashmap.Set(key, vals)
	}
	return hashmap
}

func (strmap StringMap[V]) Set(k string, v V) {
	strmap.data[k] = v
}

func (strmap StringMap[V]) Get(v string) (V, bool) {
	price, ok := strmap.data[v]
	return price, ok
}
