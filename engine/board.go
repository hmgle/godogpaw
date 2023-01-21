package engine

import (
	"strings"

	"github.com/bits-and-blooms/bitset"
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

// ParseSquare 返回字符表示的 sq.
// SquareName 的反函数.
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

	// AttackKingPawnSqs 威胁将帅的兵的位置
	AttackKingPawnSqs = make(map[int]*bitset.BitSet)

	// LegalKingMvs 将帅的合法着法位置
	LegalKingMvs = make(map[int]*bitset.BitSet)

	// LegalRedPawnMvs 兵的合法着法位置
	LegalRedPawnMvs [0xBB + 1]*bitset.BitSet
	// LegalBlackPawnMvs 卒的合法着法位置
	LegalBlackPawnMvs [0xBB + 1]*bitset.BitSet
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
			for i := sq & 0x0f; i <= 0xff; i += 0x10 {
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

	// 设置威胁帅的卒位置
	for rank := 2; rank <= 0x04; rank++ {
		for file := 0x05; file <= 0x07; file++ {
			sq := MakeSquare(file, rank)
			tmpBitSet := bitset.New(256)
			tmpBitSet.Set(uint(sq) - 1)
			tmpBitSet.Set(uint(sq) + 1)
			tmpBitSet.Set(uint(sq) + 0x10)
			AttackKingPawnSqs[sq] = tmpBitSet
		}
	}
	// 设置威胁将的兵位置
	for rank := 0x09; rank <= 0x0b; rank++ {
		for file := 0x05; file <= 0x07; file++ {
			sq := MakeSquare(file, rank)
			tmpBitSet := bitset.New(256)
			tmpBitSet.Set(uint(sq) - 1)
			tmpBitSet.Set(uint(sq) + 1)
			tmpBitSet.Set(uint(sq) - 0x10)
			AttackKingPawnSqs[sq] = tmpBitSet
		}
	}

	kingMask := bitset.New(256)
	for rank := 0x02; rank <= 0x04; rank++ {
		for file := 0x05; file <= 0x07; file++ {
			sq := MakeSquare(file, rank)
			kingMask.Set(uint(sq))
		}
	}
	for rank := 0x09; rank <= 0x0b; rank++ {
		for file := 0x05; file <= 0x07; file++ {
			sq := MakeSquare(file, rank)
			kingMask.Set(uint(sq))
		}
	}
	for rank := 0x02; rank <= 0x04; rank++ {
		for file := 0x05; file <= 0x07; file++ {
			sq := MakeSquare(file, rank)
			tmpBitSet := bitset.New(256)
			tmpBitSet.Set(uint(sq) - 1).Set(uint(sq) + 1).
				Set(uint(sq) + 0x10).Set(uint(sq) - 0x10)
			tmpBitSet.InPlaceIntersection(kingMask)
			LegalKingMvs[sq] = tmpBitSet
		}
	}
	for rank := 0x09; rank <= 0x0b; rank++ {
		for file := 0x05; file <= 0x07; file++ {
			sq := MakeSquare(file, rank)
			tmpBitSet := bitset.New(256)
			tmpBitSet.Set(uint(sq) - 1).Set(uint(sq) + 1).
				Set(uint(sq) + 0x10).Set(uint(sq) - 0x10)
			tmpBitSet.InPlaceIntersection(kingMask)
			LegalKingMvs[sq] = tmpBitSet
		}
	}

	// 兵的合法着法位置
	for rank := 5; rank <= 6; rank++ {
		for file := 2; file <= 0x0b; file++ {
			sq := MakeSquare(file, rank)
			tmpBitSet := bitset.New(256)
			tmpBitSet.Set(uint(sq) + 0x10)
			tmpBitSet.InPlaceIntersection(BoardMask)
			LegalRedPawnMvs[sq] = tmpBitSet
		}
	}
	for rank := 7; rank <= 0x0b; rank++ {
		for file := 2; file <= 0x0b; file++ {
			sq := MakeSquare(file, rank)
			tmpBitSet := bitset.New(256)
			tmpBitSet.Set(uint(sq) - 1)
			tmpBitSet.Set(uint(sq) + 1)
			tmpBitSet.Set(uint(sq) + 0x10)
			tmpBitSet.InPlaceIntersection(BoardMask)
			LegalRedPawnMvs[sq] = tmpBitSet
		}
	}
	// 卒的合法着法位置
	for rank := 8; rank >= 7; rank-- {
		for file := 2; file <= 0x0b; file++ {
			sq := MakeSquare(file, rank)
			tmpBitSet := bitset.New(256)
			tmpBitSet.Set(uint(sq) - 0x10)
			tmpBitSet.InPlaceIntersection(BoardMask)
			LegalBlackPawnMvs[sq] = tmpBitSet
		}
	}
	for rank := 6; rank >= 0x02; rank-- {
		for file := 2; file <= 0x0b; file++ {
			sq := MakeSquare(file, rank)
			tmpBitSet := bitset.New(256)
			tmpBitSet.Set(uint(sq) - 1)
			tmpBitSet.Set(uint(sq) + 1)
			tmpBitSet.Set(uint(sq) - 0x10)
			tmpBitSet.InPlaceIntersection(BoardMask)
			LegalBlackPawnMvs[sq] = tmpBitSet
		}
	}
}
