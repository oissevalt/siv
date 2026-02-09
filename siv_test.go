package siv

import (
	"runtime"
	"slices"
	"testing"
)

func TestSIV(t *testing.T) {
	s := SIV[int]{}

	h1 := s.Put(10)
	h2 := s.Put(20)
	h3 := s.Put(30)

	expect(t, slices.Equal(s.data, []int{10, 20, 30}))

	var n int
	var err error

	n, err = s.Remove(h2)
	expect(t, n == 20 && err == nil)

	h4 := s.Put(40)

	expect(t, slices.Equal(s.data, []int{10, 30, 40}))
	expect(t, h4.vid == 2)

	n, err = s.Get(h3)
	expect(t, n == 30 && err == nil)

	n, err = s.Remove(h1)
	expect(t, n == 10 && err == nil)

	expect(t, slices.Equal(s.data, []int{40, 30}))
}

func expect(t *testing.T, cond bool) {
	if !cond {
		_, _, line, ok := runtime.Caller(1)
		if ok {
			t.Fatalf("assertion failed at line %d", line)
			return
		}
		t.Fatalf("assertion failed, no caller info available")
	}
}
