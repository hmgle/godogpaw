package engine

import "math/rand"

var (
	sideKey        uint64
	pieceSquareKey [7 * 2][256]uint64 // 7 种类型棋子，2 种颜色，256 个位置，存放计算置换表的随机数
)

// PieceSquareKey 返回该棋子（piece, side）在 square 的随机数.
func PieceSquareKey(piece int, side bool, square uint) uint64 {
	return pieceSquareKey[MakePiece(piece, side)][square]
}

// ComputeKey 计算该棋盘的置换表 key，仅初始化时用.
func (p *Position) ComputeKey() uint64 {
	var result uint64
	if p.IsRedMove {
		result ^= sideKey
	}
	for i := uint(0); i < 256; i++ {
		piece := p.WhatPiece(i)
		if piece != Empty {
			result ^= PieceSquareKey(piece, p.Red.Test(i), i)
		}
	}
	return result
}

func init() {
	r := rand.New(rand.NewSource(0))
	sideKey = r.Uint64()
	for i := 0; i < 14; i++ {
		for j := 0; j < 256; j++ {
			pieceSquareKey[i][j] = r.Uint64()
		}
	}
}
