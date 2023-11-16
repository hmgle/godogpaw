package engine

type (
	PieceType = int
	Piece     = int
)

const (
	NO_PIECE_TYPE PieceType = iota
	ROOK                    // 1
	ADVISOR
	CANNON
	PAWN
	KNIGHT
	BISHOP
	KING
	KNIGHT_TO

	ALL_PIECES = 0

	PIECE_TYPE_NB = 8
)

const (
	NO_PIECE Piece = iota
	W_ROOK         // 1
	W_ADVISOR
	W_CANNON
	W_PAWN
	W_KNIGHT
	W_BISHOP
	W_KING // 7

	B_ROOK    = ROOK + 8 // 9
	B_ADVISOR = ROOK + 8 + 1
	B_CANNON  = ROOK + 8 + 2
	B_PAWN    = ROOK + 8 + 3
	B_KNIGHT  = ROOK + 8 + 4
	B_BISHOP  = ROOK + 8 + 5
	B_KING    = ROOK + 8 + 6 // 15

	PIECE_NB = ROOK + 8 + 7 // 16
)

func TypeOf(pc Piece) PieceType {
	return pc & 7
}

func ColorOf(pc Piece) Color {
	return Color(pc >> 3)
}

type Direction = int

const (
	NORTH Direction = 9
	EAST            = 1
	SOUTH           = -NORTH
	WEST            = -EAST

	NORTH_EAST = NORTH + EAST
	SOUTH_EAST = SOUTH + EAST
	SOUTH_WEST = SOUTH + WEST
	NORTH_WEST = NORTH + WEST
)

func IsOKSquare(s Square) bool {
	return s >= SQ_A0 && s <= SQ_I9
}

type Color = int8

const (
	WHITE Color = iota
	BLACK
	COLOR_NB = 2
)

func notColor(c Color) Color {
	if c == WHITE {
		return BLACK
	}
	return WHITE
}

type Phase = int

const (
	PHASE_ENDGAME Phase = 0
	PHASE_MIDGAME Phase = 128

	MG       Phase = 0
	EG       Phase = 1
	PHASE_NB Phase = 2
)

const (
	MAX_MOVES int16 = 128
	MAX_PLY   int16 = 246
)

type Value = int32

const (
	VALUE_ZERO      Value = 0
	VALUE_DRAW      Value = 0
	VALUE_KNOWN_WIN Value = 10000
	VALUE_MATE      Value = 32000
	VALUE_INFINITE  Value = 32001
	VALUE_NONE      Value = 32002

	VALUE_MATE_IN_MAX_PLY  = VALUE_MATE - Value(MAX_PLY)
	VALUE_MATED_IN_MAX_PLY = -VALUE_MATE_IN_MAX_PLY

	// Mg: mid_game eg: end_game
	RookValueMg    Value = 1245
	RookValueEg    Value = 1540
	AdvisorValueMg Value = 229
	AdvisorValueEg Value = 187
	CannonValueMg  Value = 653
	CannonValueEg  Value = 632
	PawnValueMg    Value = 80
	PawnValueEg    Value = 129
	KnightValueMg  Value = 574
	KnightValueEg  Value = 747
	BishopValueMg  Value = 308
	BishopValueEg  Value = 223
)

var PieceValue [PHASE_NB][PIECE_NB]Value = [PHASE_NB][PIECE_NB]Value{
	{
		VALUE_ZERO, RookValueMg, AdvisorValueMg, CannonValueMg, PawnValueMg, KnightValueMg, BishopValueMg, VALUE_ZERO, VALUE_ZERO, RookValueMg, AdvisorValueMg, CannonValueMg, PawnValueMg, KnightValueMg, BishopValueMg, VALUE_ZERO,
	},
	{
		VALUE_ZERO, RookValueEg, AdvisorValueEg, CannonValueEg, PawnValueEg, KnightValueEg, BishopValueEg, VALUE_ZERO, VALUE_ZERO, RookValueEg, AdvisorValueEg, CannonValueEg, PawnValueEg, KnightValueEg, BishopValueEg, VALUE_ZERO,
	},
}

type MoveNG = int

const (
	MOVE_NONE MoveNG = iota
	MOVE_NULL MoveNG = 129
)

func IsOKMove(m MoveNG) bool {
	return m != MOVE_NONE && m != MOVE_NULL
}

func FromSQ(m MoveNG) Square {
	if !IsOKMove(m) {
		panic(m)
	}
	return m >> 7
}

func ToSQ(m MoveNG) Square {
	if !IsOKMove(m) {
		panic(m)
	}
	return m & 0x7F
}

func MakeMove(from, to Square) MoveNG {
	return (from << 7) + to
}

func MakePieceNG(c Color, pt PieceType) Piece {
	return Piece(c<<3) + pt
}

func MakeSquareNG(f File, r Rank) Square {
	return r*FILE_NB + f
}

// From ethereal
const (
	HISTORY_GOOD int8 = iota
	HISTORY_TOTAL
	HISTORY_NB
)

type HistoryTable [COLOR_NB][SQUARE_NB][SQUARE_NB]Value
