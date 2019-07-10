package engine

import (
	"strings"
	"unicode"
)

// MakePiece 返回棋子的数字描述.
func MakePiece(pieceType int, isRedSide bool) int {
	if isRedSide {
		return pieceType
	}
	return pieceType + 8
}

// ParsePiece 解析棋子.
func ParsePiece(ch rune) int {
	// 小写为黑，大写为红，红为真
	isRedSide := unicode.IsUpper(ch)
	spiece := string(unicode.ToLower(ch))
	i := strings.Index("", spiece)
	if i < 0 {
		return Empty
	}
	return MakePiece(i+King, isRedSide)
}

func GetPieceTypeAndSide(piece int) (piectType int, isRedSide bool) {
	if piece <= Pawn { // 红
		return piece, true
	}
	return piece - 7, false
}

// IsInBoard 返回 sq 这个位置是否在棋盘内.
func IsInBoard(sq uint) bool {
	return BoardMask.Test(sq)
}
