package engine

import "github.com/bits-and-blooms/bitset"

const (
	rookVal    = 600
	knightVal  = 270
	cannonVal  = 285
	advisorVal = 120
	bishopVal  = 120
	pawnVal    = 30
	kingVal    = 6000
)

const exposedCannonVal = 55

var valMap = map[int]int{
	Rook:    rookVal,
	Knight:  knightVal,
	Cannon:  cannonVal,
	Advisor: advisorVal,
	Bishop:  bishopVal,
	Pawn:    pawnVal,
	King:    knightVal,
}

var (
	// RedRookPstValue 红车位置价值
	RedRookPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, -2, 10, 6, 14, 12, 14, 6, 10, -2, 0, 0, 0, 0, 0,
		0, 0, 8, 4, 8, 16, 8, 16, 8, 4, 8, 0, 0, 0, 0, 0,
		0, 0, 4, 8, 6, 14, 12, 14, 6, 8, 4, 0, 0, 0, 0, 0,
		0, 0, 6, 10, 8, 14, 14, 14, 8, 10, 6, 0, 0, 0, 0, 0,
		0, 0, 12, 16, 14, 20, 20, 20, 14, 16, 12, 0, 0, 0, 0, 0,
		0, 0, 12, 14, 12, 18, 18, 18, 12, 14, 12, 0, 0, 0, 0, 0,
		0, 0, 12, 18, 16, 22, 22, 22, 16, 18, 12, 0, 0, 0, 0, 0,
		0, 0, 12, 12, 12, 18, 18, 18, 12, 12, 12, 0, 0, 0, 0, 0,
		0, 0, 16, 20, 18, 24, 26, 24, 18, 20, 16, 0, 0, 0, 0, 0,
		0, 0, 14, 14, 12, 18, 16, 18, 12, 14, 14, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// BlackRookPstValue 黑车位置价值
	BlackRookPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 14, 14, 12, 18, 16, 18, 12, 14, 14, 0, 0, 0, 0, 0,
		0, 0, 16, 20, 18, 24, 26, 24, 18, 20, 16, 0, 0, 0, 0, 0,
		0, 0, 12, 12, 12, 18, 18, 18, 12, 12, 12, 0, 0, 0, 0, 0,
		0, 0, 12, 18, 16, 22, 22, 22, 16, 18, 12, 0, 0, 0, 0, 0,
		0, 0, 12, 14, 12, 18, 18, 18, 12, 14, 12, 0, 0, 0, 0, 0,
		0, 0, 12, 16, 14, 20, 20, 20, 14, 16, 12, 0, 0, 0, 0, 0,
		0, 0, 6, 10, 8, 14, 14, 14, 8, 10, 6, 0, 0, 0, 0, 0,
		0, 0, 4, 8, 6, 14, 12, 14, 6, 8, 4, 0, 0, 0, 0, 0,
		0, 0, 8, 4, 8, 16, 8, 16, 8, 4, 8, 0, 0, 0, 0, 0,
		0, 0, -2, 10, 6, 14, 12, 14, 6, 10, -2, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// RedCannonPstValue 红炮位置价值
	RedCannonPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 2, 6, 6, 6, 2, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 2, 4, 6, 6, 6, 4, 2, 0, 0, 0, 0, 0, 0,
		0, 0, 4, 0, 8, 6, 10, 6, 8, 0, 4, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 2, 4, 2, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, -2, 0, 4, 2, 6, 2, 4, 0, -2, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 2, 8, 2, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, -2, 4, 10, 4, -2, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 2, 2, 0, -10, -8, -10, 0, 2, 2, 0, 0, 0, 0, 0,
		0, 0, 2, 2, 0, -4, -14, -4, 0, 2, 2, 0, 0, 0, 0, 0,
		0, 0, 6, 4, 0, -10, -12, -10, 0, 4, 6, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// BlackCannonPstValue 黑炮位置价值
	BlackCannonPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 6, 4, 0, -10, -12, -10, 0, 4, 6, 0, 0, 0, 0, 0,
		0, 0, 2, 2, 0, -4, -14, -4, 0, 2, 2, 0, 0, 0, 0, 0,
		0, 0, 2, 2, 0, -10, -8, -10, 0, 2, 2, 0, 0, 0, 0, 0,
		0, 0, 0, 0, -2, 4, 10, 4, -2, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 2, 8, 2, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, -2, 0, 4, 2, 6, 2, 4, 0, -2, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 2, 4, 2, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 4, 0, 8, 6, 10, 6, 8, 0, 4, 0, 0, 0, 0, 0,
		0, 0, 0, 2, 4, 6, 6, 6, 4, 2, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 2, 6, 6, 6, 2, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// RedKnightPstValue 红马位置价值
	RedKnightPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, -4, 0, 0, 0, 0, 0, -4, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 2, 4, 4, -2, 4, 4, 2, 0, 0, 0, 0, 0, 0,
		0, 0, 4, 2, 8, 8, 4, 8, 8, 2, 4, 0, 0, 0, 0, 0,
		0, 0, 2, 6, 8, 6, 10, 6, 8, 6, 2, 0, 0, 0, 0, 0,
		0, 0, 4, 12, 16, 14, 12, 14, 16, 12, 4, 0, 0, 0, 0, 0,
		0, 0, 6, 16, 14, 18, 16, 18, 14, 16, 6, 0, 0, 0, 0, 0,
		0, 0, 8, 24, 18, 24, 20, 24, 18, 24, 8, 0, 0, 0, 0, 0,
		0, 0, 12, 14, 16, 20, 18, 20, 16, 14, 12, 0, 0, 0, 0, 0,
		0, 0, 4, 10, 28, 16, 8, 16, 28, 10, 4, 0, 0, 0, 0, 0,
		0, 0, 4, 8, 16, 12, 4, 12, 16, 8, 4, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// BlackKnightPstValue 黑马位置价值
	BlackKnightPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 4, 8, 16, 12, 4, 12, 16, 8, 4, 0, 0, 0, 0, 0,
		0, 0, 4, 10, 28, 16, 8, 16, 28, 10, 4, 0, 0, 0, 0, 0,
		0, 0, 12, 14, 16, 20, 18, 20, 16, 14, 12, 0, 0, 0, 0, 0,
		0, 0, 8, 24, 18, 24, 20, 24, 18, 24, 8, 0, 0, 0, 0, 0,
		0, 0, 6, 16, 14, 18, 16, 18, 14, 16, 6, 0, 0, 0, 0, 0,
		0, 0, 4, 12, 16, 14, 12, 14, 16, 12, 4, 0, 0, 0, 0, 0,
		0, 0, 2, 6, 8, 6, 10, 6, 8, 6, 2, 0, 0, 0, 0, 0,
		0, 0, 4, 2, 8, 8, 4, 8, 8, 2, 4, 0, 0, 0, 0, 0,
		0, 0, 0, 2, 4, 4, -2, 4, 4, 2, 0, 0, 0, 0, 0, 0,
		0, 0, 0, -4, 0, 0, 0, 0, 0, -4, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// RedBishopPstValue 红相位置价值
	RedBishopPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, -7, 0, 0, 0, 6, 0, 0, 0, -7, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, -2, 0, 0, 0, -2, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// BlackBishopPstValue 黑象位置价值
	BlackBishopPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, -2, 0, 0, 0, -2, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, -7, 0, 0, 0, 6, 0, 0, 0, -7, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// RedAdvisorPstValue 红士位置价值
	RedAdvisorPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, -2, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, -2, 0, -2, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// BlackAdvisorPstValue 黑士位置价值
	BlackAdvisorPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, -2, 0, -2, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, -2, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// RedKingPstValue 红帅位置价值
	RedKingPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, -5, 0, -5, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, -10, -10, -10, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, -15, -15, -15, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// BlackKingPstValue 黑将位置价值
	BlackKingPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, -15, -15, -15, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, -10, -10, -10, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, -5, 0, -5, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// RedPawnPstValue 红兵位置价值
	RedPawnPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, -2, 0, 4, 0, -2, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 2, 0, 8, 0, 8, 0, 8, 0, 2, 0, 0, 0, 0, 0,
		0, 0, 6, 12, 18, 18, 20, 18, 18, 12, 6, 0, 0, 0, 0, 0,
		0, 0, 10, 20, 30, 34, 40, 34, 30, 20, 10, 0, 0, 0, 0, 0,
		0, 0, 14, 26, 42, 60, 80, 60, 42, 26, 14, 0, 0, 0, 0, 0,
		0, 0, 18, 36, 56, 80, 120, 80, 56, 36, 18, 0, 0, 0, 0, 0,
		0, 0, 0, 3, 6, 9, 12, 9, 6, 3, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	// BlackPawnPstValue 黑兵位置价值
	BlackPawnPstValue = [...]int{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 3, 6, 9, 12, 9, 6, 3, 0, 0, 0, 0, 0, 0,
		0, 0, 18, 36, 56, 80, 120, 80, 56, 36, 18, 0, 0, 0, 0, 0,
		0, 0, 14, 26, 42, 60, 80, 60, 42, 26, 14, 0, 0, 0, 0, 0,
		0, 0, 10, 20, 30, 34, 40, 34, 30, 20, 10, 0, 0, 0, 0, 0,
		0, 0, 6, 12, 18, 18, 20, 18, 18, 12, 6, 0, 0, 0, 0, 0,
		0, 0, 2, 0, 8, 0, 8, 0, 8, 0, 2, 0, 0, 0, 0, 0,
		0, 0, 0, 0, -2, 0, 4, 0, -2, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
)

func rookValue(sq int, isRed bool) int {
	if isRed {
		return RedRookPstValue[sq]
	}
	return BlackRookPstValue[sq]
}

func knightValue(sq int, isRed bool) int {
	if isRed {
		return RedKnightPstValue[sq]
	}
	return BlackKnightPstValue[sq]
}

func cannonValue(sq int, isRed bool) int {
	if isRed {
		return RedCannonPstValue[sq]
	}
	return BlackCannonPstValue[sq]
}

func advisorValue(sq int, isRed bool) int {
	if isRed {
		return RedAdvisorPstValue[sq]
	}
	return BlackAdvisorPstValue[sq]
}

func bishopValue(sq int, isRed bool) int {
	if isRed {
		return RedBishopPstValue[sq]
	}
	return BlackBishopPstValue[sq]
}

func kingValue(sq int, isRed bool) int {
	if isRed {
		return RedKingPstValue[sq]
	}
	return BlackKingPstValue[sq]
}

func pawnValue(sq int, isRed bool) int {
	if isRed {
		return RedPawnPstValue[sq]
	}
	return BlackPawnPstValue[sq]
}

// 缺士车加分
func (p *Position) rookAwardVal(isRed bool) int {
	// TODO
	return 0
}

// 缺象炮加分
func (p *Position) cannonAwardVal(isRed bool) int {
	// TODO
	return 0
}

// knightDexterity 马罚分.
func (p *Position) knightDexterity(sq int, isRed bool) int {
	var (
		punishVal int
		selfPs    *bitset.BitSet
		sidePs    *bitset.BitSet
	)
	if isRed {
		selfPs, sidePs = p.Red, p.Black
	} else {
		selfPs, sidePs = p.Black, p.Red
	}
	blocksDeltas := []int{0x01, -0x01, 0x10, -0x10}
	for _, blockDelta := range blocksDeltas {
		if selfPs.Test(uint(sq + blockDelta)) {
			punishVal -= 5
		} else if sidePs.Test(uint(sq + blockDelta)) {
			punishVal -= 10
		}
	}
	return punishVal
}

// isExposedCannon 是否空头.
func (p *Position) isExposedCannon(cannonSq uint, isRed bool) int {
	var beCheckKingSq uint
	if isRed {
		blackKing := p.Kings.Intersection(p.Black)
		beCheckKingSq, _ = blackKing.NextSet(0)
	} else {
		beCheckKingSq, _ = p.Kings.NextSet(0)
	}
	// XXX debug
	if beCheckKingSq == 0 {
		return 0
	}
	rookAttacks := RookAttacks[int(beCheckKingSq)]
	if !rookAttacks.Test(cannonSq) {
		return 0
	}
	if File(int(cannonSq)) == File(int(beCheckKingSq)) {
		if !p.IsAnyPieceBetweenFile(int(cannonSq), int(beCheckKingSq)) {
			return exposedCannonVal
		}
	} else if Rank(int(cannonSq)) == Rank(int(beCheckKingSq)) {
		if !p.IsAnyPieceBetweenRank(int(cannonSq), int(beCheckKingSq)) {
			return exposedCannonVal
		}
	}
	return 0
}

func (p *Position) initEval() {
	p.redStrengthVal = 0
	p.blackStrengthVal = 0
	p.redPstVal = 0
	p.blackPstVal = 0

	redRooks := p.Rooks.Intersection(p.Red)
	for sq, e := redRooks.NextSet(0); e; sq, e = redRooks.NextSet(sq + 1) {
		p.redStrengthVal += rookVal
		p.redPstVal += RedRookPstValue[sq]
	}
	blackRooks := p.Rooks.Intersection(p.Black)
	for sq, e := blackRooks.NextSet(0); e; sq, e = blackRooks.NextSet(sq + 1) {
		p.blackStrengthVal += rookVal
		p.blackPstVal += BlackRookPstValue[sq]
	}

	redCannons := p.Cannons.Intersection(p.Red)
	for sq, e := redCannons.NextSet(0); e; sq, e = redCannons.NextSet(sq + 1) {
		p.redStrengthVal += cannonVal
		p.redPstVal += RedCannonPstValue[sq]
	}
	blackCannons := p.Cannons.Intersection(p.Black)
	for sq, e := blackCannons.NextSet(0); e; sq, e = blackCannons.NextSet(sq + 1) {
		p.blackStrengthVal += cannonVal
		p.blackPstVal += BlackCannonPstValue[sq]
	}

	redKnights := p.Knights.Intersection(p.Red)
	for sq, e := redKnights.NextSet(0); e; sq, e = redKnights.NextSet(sq + 1) {
		p.redStrengthVal += knightVal
		p.redPstVal += RedKingPstValue[sq]
	}
	blackKnights := p.Knights.Intersection(p.Black)
	for sq, e := blackKnights.NextSet(0); e; sq, e = blackKnights.NextSet(sq + 1) {
		p.blackStrengthVal += knightVal
		p.blackPstVal += BlackKingPstValue[sq]
	}

	redPawns := p.Pawns.Intersection(p.Red)
	for sq, e := redPawns.NextSet(0); e; sq, e = redPawns.NextSet(sq + 1) {
		p.redStrengthVal += pawnVal
		p.redPstVal += RedPawnPstValue[sq]
	}
	blackPawns := p.Pawns.Intersection(p.Black)
	for sq, e := blackPawns.NextSet(0); e; sq, e = blackPawns.NextSet(sq + 1) {
		p.blackStrengthVal += pawnVal
		p.blackPstVal += BlackPawnPstValue[sq]
	}

	redBishops := p.Bishops.Intersection(p.Red)
	for sq, e := redBishops.NextSet(0); e; sq, e = redBishops.NextSet(sq + 1) {
		p.redStrengthVal += bishopVal
		p.redPstVal += RedBishopPstValue[sq]
	}
	blackBishops := p.Bishops.Intersection(p.Black)
	for sq, e := blackBishops.NextSet(0); e; sq, e = blackBishops.NextSet(sq + 1) {
		p.blackStrengthVal += bishopVal
		p.blackPstVal += BlackBishopPstValue[sq]
	}

	redAdvisors := p.Advisors.Intersection(p.Red)
	for sq, e := redAdvisors.NextSet(0); e; sq, e = redAdvisors.NextSet(sq + 1) {
		p.redStrengthVal += advisorVal
		p.redPstVal += RedAdvisorPstValue[sq]
	}
	blackAdvisors := p.Advisors.Intersection(p.Black)
	for sq, e := blackAdvisors.NextSet(0); e; sq, e = blackAdvisors.NextSet(sq + 1) {
		p.blackStrengthVal += advisorVal
		p.blackPstVal += BlackAdvisorPstValue[sq]
	}

	redKings := p.Kings.Intersection(p.Red)
	for sq, e := redKings.NextSet(0); e; sq, e = redKings.NextSet(sq + 1) {
		p.redStrengthVal += kingVal
		p.redPstVal += RedKingPstValue[sq]
	}
	blackKings := p.Kings.Intersection(p.Black)
	for sq, e := blackKings.NextSet(0); e; sq, e = blackKings.NextSet(sq + 1) {
		p.blackStrengthVal += kingVal
		p.blackPstVal += BlackKingPstValue[sq]
	}
}

func (p *Position) Evaluate() int {
	eval := p.redStrengthVal + p.redPstVal - p.blackStrengthVal - p.blackPstVal

	redCannons := p.Cannons.Intersection(p.Red)
	for sq, e := redCannons.NextSet(0); e; sq, e = redCannons.NextSet(sq + 1) {
		eval += p.isExposedCannon(sq, true)
	}
	blackCannons := p.Cannons.Intersection(p.Black)
	for sq, e := blackCannons.NextSet(0); e; sq, e = blackCannons.NextSet(sq + 1) {
		eval -= p.isExposedCannon(sq, false)
	}

	redKnights := p.Knights.Intersection(p.Red)
	for sq, e := redKnights.NextSet(0); e; sq, e = redKnights.NextSet(sq + 1) {
		eval += p.knightDexterity(int(sq), true)
	}
	blackKnights := p.Knights.Intersection(p.Black)
	for sq, e := blackKnights.NextSet(0); e; sq, e = blackKnights.NextSet(sq + 1) {
		eval -= p.knightDexterity(int(sq), false)
	}

	if p.IsRedMove {
		return eval
	}
	return -eval
}
