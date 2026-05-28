// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package hashset

type ImmutableStringMap[V any] struct {
	strm StringMap[V]
}

func ImmutableStringMapFromMap[V any](v map[string]V) *ImmutableStringMap[V] {
	return &ImmutableStringMap[V]{strm: StringMapFromMap(v)}
}

func (i *ImmutableStringMap[V]) Get(v string) (V, bool) {
	return i.strm.Get(v)
}
