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

func TTClear() {
	age = 0
	clear(TT.Entries)
}

func UpdateAge() {
	age += 1
}

func TTSave(key Key, score int16, flag int8, ply, depth uint8, move MoveNG) {
	entry := &TT.Entries[key&TT.Mask]
	if entry.Key == 0 ||
		(entry.Age < age && flag == int8(TT_EXACT)) ||
		entry.Depth-2*(age-entry.Age)+boolTouint8(flag != int8(TT_EXACT) && entry.Flag == int8(TT_EXACT)) <= depth {
		// If the score we get from the transposition table is a checkmate score, we need
		// to do a little extra work. This is because we store checkmates in the table using
		// their distance from the node they're found in, not their distance from the root.
		// So if we found a checkmate-in-8 in a node that was 5 plies from the root, we need
		// to store the score as a checkmate-in-3. Then, if we read the checkmate-in-3 from
		// the table in a node that's 4 plies from the root, we need to return the score as
		// checkmate-in-7.
		if score > int16(VALUE_MATE_IN_MAX_PLY) {
			score += int16(ply)
		} else if score < -int16(VALUE_MATE_IN_MAX_PLY) {
			score -= int16(ply)
		}
		entry.Key = key
		entry.Score = score
		entry.Depth = depth
		entry.Flag = flag
		entry.Move = move
		entry.Age = age
	}
}

func TTProbe(key Key) *TTEntry {
	entry := &TT.Entries[key&TT.Mask]
	if entry.Key == key {
		return entry
	}
	return nil
}

func boolTouint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
