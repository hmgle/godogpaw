package engine

import (
	"unsafe"
)

type TTEntry struct {
	Key   Key
	Score int16
	Depth uint8
	Flag  int8
	Move  MoveNG
	Age   uint8
}

const (
	TT_ALPHA int8 = 1 + iota
	TT_BETA
	TT_EXACT
)

type TransTable struct {
	Entries []TTEntry
	Mask    uint64
}

var (
	TT  = NewTranTable(16)
	age uint8
)

func roundPowerOfTwo(size int) int {
	x := 1
	for (x << 1) <= size {
		x <<= 1
	}
	return x
}

func NewTranTable(megabytes int) *TransTable {
	megabytes = min(4096, max(1, megabytes))
	size := 1024 * 1024 * megabytes / int(unsafe.Sizeof(TTEntry{}))
	return &TransTable{
		Entries: make([]TTEntry, size),
		Mask:    uint64(roundPowerOfTwo(size - 1)),
	}
}

