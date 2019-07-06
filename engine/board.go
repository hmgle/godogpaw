package engine

import (
	"strings"

	"github.com/willf/bitset"
)

/*
F0 F1 F2 F3 F4 F5 F6 F7 F8 F9 FA FB FC FD FE FF
E0 E1 E2 E3 E4 E5 E6 E7 E8 E9 EA EB EC ED EE EF
D0 D1 D2 D3 D4 D5 D6 D7 D8 D9 DA DB DC DD DE DF
C0 C1 C2 C3 C4 C5 C6 C7 C8 C9 CA CB CC CD CE CF
B0 B1 B2-B3-B4-B5-B6-B7-B8-B9-BA BB BC BD BE BF
      |  |  |  | \|/ |  |  |  |
A0 A1 A2-A3-A4-A5-A6-A7-A8-A9-AA AB AC AD AE AF
      |  |  |  | /|\ |  |  |  |
90 91 92-93-94-95-96-97-98-99-9A 9B 9C 9D 9E 9F
      |  |  |  |  |  |  |  |  |
80 81 82-83-84-85-86-87-88-89-8A 8B 8C 8D 8E 8F
      |  |  |  |  |  |  |  |  |
70 71 72-73-74-75-76-77-78-79-7A 7B 7C 7D 7E 7F
      |                       |
60 61 62-63-64-65-66-67-68-69-6A 6B 6C 6D 6E 6F
      |  |  |  |  |  |  |  |  |
50 51 52-53-54-55-56-57-58-59-5A 5B 5C 5D 5E 5F
      |  |  |  |  |  |  |  |  |
40 41 42-43-44-45-46-47-48-49-4A 4B 4C 4D 4E 4F
      |  |  |  | \|/ |  |  |  |
30 31 32-33-34-35-36-37-38-39-3A 3B 3C 3D 3E 3F
      |  |  |  | /|\ |  |  |  |
20 21 22-23-24-25-26-27-28-29-2A-2B 2C 2D 2E 2F
10 11 12 13 14 15 16 17 18 19 1A 1B 1C 1D 1E 1F
00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F
*/

const (
	fileNames = "##abcdefghi#####"
	rankNames = "##0123456789####"
)

// SquareName 返回位置的字符表示.
// 如 sq = 0x22, 返回 "a0".
// 如 sq = 0x23, 返回 "b0".
// 如 sq = 0x32, 返回 "a1".
func SquareName(sq int) string {
	file := fileNames[sq&0x0F]
	rank := rankNames[sq>>4]
	return string(file) + string(rank)
}

func MakeSquare(file, rank int) int {
	return rank<<4 | file
}

const (
	SquareNone = -1
)

func ParseSquare(s string) int {
	if len(s) < 2 || s == "-" {
		return SquareNone
	}
	file := strings.Index(fileNames, s[0:1])
	rank := strings.Index(rankNames, s[1:2])
	return MakeSquare(file, rank)
}

// File 列.
func File(sq int) int {
	return sq & 0x0F
}

// Rank 行.
func Rank(sq int) int {
	return sq >> 4
}

var (
	// GlobalBoard 全局 board.
	GlobalBoard = bitset.New(256)
	// RedBoard 红方 board.
	RedBoard = bitset.New(256)
	// BlackBoard 黑方 board.
	BlackBoard = bitset.New(256)

	// BoardMask 棋盘
	BoardMask = bitset.New(256)

	// FileMasks 列屏蔽
	FileMasks = []*bitset.BitSet{}

	// RankMasks 行屏蔽
	RankMasks = []*bitset.BitSet{}

	// RookAttacks 车攻击位置
	RookAttacks = make(map[int]*bitset.BitSet)

	// KnightAttacks 马攻击位置
	KnightAttacks = make(map[int]*bitset.BitSet)
)

const (
	// Empty 无
	Empty int = iota
	// King 帅
	King
	// Rook 车
	Rook
	// Knight 马
	Knight
	// Cannon 炮
	Cannon
	// Advisor 士
	Advisor
	// Bishop 象
	Bishop
	// Pawn 兵
	Pawn
)

// Piece 位置信息.
type Piece struct {
	Position uint8
	Name     int8
	Color    bool // is red?
}

func init() {
	for rank := 2; rank <= 0x0b; rank++ {
		for file := 2; file <= 0x0a; file++ {
			sq := MakeSquare(file, rank)
			BoardMask.Set(uint(sq))
		}
	}
	// 初始化列屏蔽位.
	for file := 0; file <= 0x0a; file++ {
		newBitSet := bitset.New(256)
		for rank := 2; rank <= 0x0b; rank++ {
			sq := MakeSquare(file, rank)
			newBitSet.Set(uint(sq))
		}
		FileMasks = append(FileMasks, newBitSet)
	}
	// 初始化行屏蔽位.
	for rank := 0; rank <= 0x0b; rank++ {
		newBitSet := bitset.New(256)
		for file := 0; file <= 0x0a; file++ {
			sq := MakeSquare(file, rank)
			newBitSet.Set(uint(sq))
		}
		RankMasks = append(RankMasks, newBitSet)
	}

	for rank := 2; rank <= 0x0b; rank++ {
		for file := 2; file <= 0x0a; file++ {
			sq := MakeSquare(file, rank)
			tmpBitSet := bitset.New(256)
			// 与 sq 同一行
			for i := sq & 0xf0; i <= sq+0x0f; i++ {
				tmpBitSet.Set(uint(i))
			}
			// 与 sq 同一列
			for i := sq & 0x0f; i <= 0xff; i = i + 0x1f {
				tmpBitSet.Set(uint(i))
			}
			tmpBitSet.InPlaceIntersection(BoardMask)
			// 清除本格
			tmpBitSet.Clear(uint(sq))
			RookAttacks[sq] = tmpBitSet
		}
	}

	for rank := 2; rank <= 0x0b; rank++ {
		for file := 2; file <= 0x0a; file++ {
			sq := MakeSquare(file, rank)
			tmpBitSet := bitset.New(256)
			tmpBitSet.Set(uint(sq + 0x10*2 + 1))
			tmpBitSet.Set(uint(sq + 0x10*2 - 1))
			tmpBitSet.Set(uint(sq + 0x10 + 2))
			tmpBitSet.Set(uint(sq + 0x10 - 2))
			tmpBitSet.Set(uint(sq - 0x10*2 + 1))
			tmpBitSet.Set(uint(sq - 0x10*2 - 1))
			tmpBitSet.Set(uint(sq - 0x10 + 2))
			tmpBitSet.Set(uint(sq - 0x10 - 2))
			tmpBitSet.InPlaceIntersection(BoardMask)
			KnightAttacks[sq] = tmpBitSet
		}
	}
}
