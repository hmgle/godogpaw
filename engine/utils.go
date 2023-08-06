package engine

import (
	"strings"
	"unicode"

	"github.com/bits-and-blooms/bitset"
)

// MakePiece 返回棋子的数字描述.
func MakePiece(pieceType int, isRedSide bool) int {
	if pieceType == Empty {
		return Empty
	}
	if isRedSide {
		return pieceType
	}
	return pieceType + 8
}

// ParsePiece 解析棋子.
func ParsePiece(ch rune) int {
	// 大写为红，小写为黑，红为真
	isRedSide := unicode.IsUpper(ch)
	spiece := string(unicode.ToLower(ch))
	i := strings.Index("krncabp", spiece)
	if i < 0 {
		i = strings.Index("krhcaep", spiece)
		if i < 0 {
			return Empty
		}
	}
	return MakePiece(i+King, isRedSide)
}

func GetPieceTypeAndSide(piece int) (piectType int, isRedSide bool) {
	if piece <= Pawn { // 红
		return piece, true
	}
	return piece - 8, false
}

// IsInBoard 返回 sq 这个位置是否在棋盘内.
func IsInBoard(sq uint) bool {
	return BoardMask.Test(sq)
}

// LegalPawnMvs 返回 sq 这个位置兵卒的合法着法位置.
func LegalPawnMvs(sq int, isRedSide bool) *bitset.BitSet {
	if isRedSide {
		return LegalRedPawnMvs[sq]
	}
	return LegalBlackPawnMvs[sq]
}

// LegalAdvisorMvs 返回 sq 这个位置士的合法着法位置.
func LegalAdvisorMvs(sq uint) *bitset.BitSet {
	movs := bitset.New(256)
	switch sq {
	case 0x25, 0x27, 0x45, 0x47:
		movs.Set(0x36)
	case 0x36:
		movs.Set(0x25).Set(0x27).Set(0x45).Set(0x47)
	case 0x95, 0x97, 0xb5, 0xb7:
		movs.Set(0xa6)
	case 0xa6:
		movs.Set(0x95).Set(0x97).Set(0xb5).Set(0xb7)
	}
	return movs
}
