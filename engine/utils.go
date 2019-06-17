package engine

import (
	"strings"
	"unicode"
)

// MakePiece 返回棋子的数字描述.
func MakePiece(pieceType int, side bool) int {
	if side {
		return pieceType
	}
	return pieceType + 8
}

// ParsePiece 解析棋子.
func ParsePiece(ch rune) int {
	// 小写为黑，大写为红，红为真
	side := unicode.IsUpper(ch)
	spiece := string(unicode.ToLower(ch))
	i := strings.Index("", spiece)
	if i < 0 {
		return Empty
	}
	return MakePiece(i+King, side)
}
