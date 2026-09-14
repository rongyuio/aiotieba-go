// Package classdef holds the shared data models of the API responses.
//
// It mirrors the Python package aiotieba.api._classdef.
package classdef

// Containers is the generic base of the content lists. It mirrors
// aiotieba.api._classdef.container.Containers.
type Containers[T any] struct {
	Objs []T
}

// Len returns the number of items. It mirrors __len__.
func (c Containers[T]) Len() int { return len(c.Objs) }

// Empty reports whether the list is empty. It mirrors __bool__.
func (c Containers[T]) Empty() bool { return len(c.Objs) == 0 }
