package engine

import (
	"math/rand"
)

type transEntry struct {
	key    uint64
	forRed bool
	depth  uint8
	score  int16
	move   Move
}

const (
	tableSize = 1024 * 1024 * 5 / 16
)

var (
	// sideKey        uint64
	pieceSquareKey [7 * 2][256]uint64 // 7 种类型棋子，2 种颜色，256 个位置，存放计算置换表的随机数

	transTable []transEntry = make([]transEntry, tableSize)
)

func clearTransTable() {
	for i := range transTable {
		transTable[i] = transEntry{}
	}
}

func ProbeHash(forRed bool, key uint64, depth uint8) (e *transEntry, ok bool) {
	index := key % tableSize
	if transTable[index].key != key {
		return nil, false
	}
	if transTable[index].forRed != forRed {
		return nil, false
	}
	e = &transTable[index]
	if e.depth >= depth {
		return e, true
	}
	return nil, false
}

func RecordHash(forRed bool, key uint64, depth uint8, score int16, move Move) {
	index := key % tableSize
	if transTable[index].key == key && transTable[index].depth >= depth {
		return
	}
	transTable[index].key = key
	transTable[index].depth = depth
	transTable[index].forRed = forRed
	transTable[index].move = move
	transTable[index].score = score
}

func (p *Position) ProbeHash(depth uint8) (bestMove int32, score int, ok bool) {
	e, probeOk := ProbeHash(p.IsRedMove, p.Key, depth)
	if !probeOk {
		return
	}
	return int32(e.move), int(e.score), true
}

func (p *Position) RecordHash(depth uint8, score int16, move int32) {
	RecordHash(p.IsRedMove, p.Key, depth, score, Move(move))
}

// getPieceSquareKey 返回该棋子（piece, side）在 square 的随机数.
// side: 红行为真.
func getPieceSquareKey(piece int, side bool, square uint) uint64 {
	if side {
		return pieceSquareKey[piece-1][square]
	}
	return pieceSquareKey[piece+6][square]
}

// ComputeKey 计算该棋盘的置换表 key，仅初始化时用.
func (p *Position) ComputeKey() uint64 {
	var result uint64
	/*
		if p.IsRedMove {
			result ^= sideKey
		}
	*/
	for i := uint(0); i < 256; i++ {
		piece := p.WhatPiece(i)
		if piece != Empty {
			result ^= getPieceSquareKey(piece, p.Red.Test(i), i)
		}
	}
	// clearTransTable()
	return result
}

func init() {
	r := rand.New(rand.NewSource(0))
	// sideKey = r.Uint64()
	for i := 0; i < 14; i++ {
		for j := 0; j < 256; j++ {
			pieceSquareKey[i][j] = r.Uint64()
		}
	}
}
