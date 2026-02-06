package siv

import (
	"errors"
	"iter"
)

var (
	ErrInvalid = errors.New("handle is invalid")
	ErrExpired = errors.New("handle has expired")
)

// SIV is an implementation of Jean Tampon's [Stable Index Vector]. This
// data structure achieves O(1) time complexity in appending, insertion,
// deletion and fast access, and is stable.
//
// The empty structure is ready to use with zero capacity. Alternatively,
// initiate an instance a certain capacity with [WithCapacity].
//
// [Stable Index Vector]: https://github.com/johnBuffer/StableIndexVector
type SIV[T any] struct {
	data    []T
	indices []int
	meta    []metadata
}

// Handle is a reference to an item stored in SIV. See [SIV.Get].
type Handle[T any] metadata

type metadata struct {
	rid int
	vid int
}

func WithCapacity[T any](cap int) *SIV[T] {
	return &SIV[T]{
		data:    make([]T, 0, cap),
		indices: make([]int, 0, cap),
		meta:    make([]metadata, 0, cap),
	}
}

// Get returns the item represented by the handle. In case of error,
// ErrInvalid indicates h is malformed, while ErrExpired indicates
// the desired item has been deleted.
func (s *SIV[T]) Get(h Handle[T]) (item T, err error) {
	id, err2 := s.findID(h)
	if err2 != nil {
		err = err2
		return
	}
	return s.data[id], nil
}

// Set updates the value of the item represented by h, returning
// the previous value.
func (s *SIV[T]) Set(h Handle[T], v T) (old T, err error) {
	id, err2 := s.findID(h)
	if err2 != nil {
		err = err2
		return
	}
	old, s.data[id] = s.data[id], v
	return
}

func (s *SIV[T]) Len() int {
	return len(s.data)
}

func (s *SIV[T]) Cap() int {
	return cap(s.data)
}

// Put adds an item to the SIV, returning a handle to it.
func (s *SIV[T]) Put(item T) Handle[T] {
	id := len(s.data)
	if len(s.meta) > len(s.data) {
		s.data = append(s.data, item)
		s.meta[id].vid++
		return Handle[T](s.meta[id])
	}
	s.data = append(s.data, item)
	s.indices = append(s.indices, id)
	s.meta = append(s.meta, metadata{id, 0})
	return Handle[T]{id, 0}
}

// Pop removes and returns the last item in the SIV.
// The returned item is not necessarily the last added one.
// It panics if the SIV is empty.
func (s *SIV[T]) Pop() T {
	if len(s.data) == 0 {
		panic("siv: no item to pop")
	}
	it, _ := s.Remove(Handle[T](s.meta[len(s.data)-1]))
	return it
}

// Remove removes the item represented by the handle from the SIV.
func (s *SIV[T]) Remove(h Handle[T]) (item T, err error) {
	id1, err2 := s.findID(h)
	if err2 != nil {
		err = err2
		return
	}
	id2 := len(s.data) - 1
	item = s.data[id1]
	rid1, rid2 := h.rid, s.meta[id2].rid
	if id1 != id2 {
		s.data[id1], s.data[id2] = s.data[id2], s.data[id1]
		s.meta[id1], s.meta[id2] = s.meta[id2], s.meta[id1]
		s.indices[rid1], s.indices[rid2] = s.indices[rid2], s.indices[rid1]
	}
	s.meta[id2].vid++
	s.data = s.data[:len(s.data)-1]
	return
}

func (s *SIV[T]) findID(h Handle[T]) (int, error) {
	if h.rid < 0 || h.rid >= len(s.indices) {
		return 0, ErrInvalid
	}
	id := s.indices[h.rid]
	if m := s.meta[id]; m.vid != h.vid {
		return 0, ErrExpired
	}
	return id, nil
}

// Iter returns an iterator over the items in the same order as
// they are stored in the underlying array.
func (s *SIV[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s.data {
			if !yield(v) {
				return
			}
		}
	}
}

// Iter2 returns an iterator over the items and their corresponding
// handles, ordered in the same way as Iter. Note that using the handles
// to modify items mid-iteration might cause unintended side effects.
func (s *SIV[T]) Iter2() iter.Seq2[Handle[T], T] {
	return func(yield func(Handle[T], T) bool) {
		for i, v := range s.data {
			h := Handle[T](s.meta[i])
			if !yield(h, v) {
				return
			}
		}
	}
}
