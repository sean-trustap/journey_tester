// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package hashset

type ImmutableSet[T comparable] struct {
	set Set[T]
}

func ImmutableSetFromSlice[T comparable](vals []T) *ImmutableSet[T] {
	return &ImmutableSet[T]{set: SetFromSlice(vals)}
}

func (s *ImmutableSet[T]) Has(v T) bool {
	return s.set.Has(v)
}

func (s *ImmutableSet[T]) AsSlice() []T {
	return s.set.AsSlice()
}
