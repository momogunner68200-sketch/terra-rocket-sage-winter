package main

import (
	"testing"
	"unsafe"
)

func TestKeyInputSize(t *testing.T) {
	if unsafe.Sizeof(keyInput{}) != 40 {
		t.Fatalf("INPUT size = %d, want 40", unsafe.Sizeof(keyInput{}))
	}
}
