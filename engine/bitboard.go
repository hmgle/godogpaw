package engine

import (
	"log"
	"math/bits"
	"time"
	"unsafe"
)

type Bitboard struct {
	Lo uint64
	Hi uint64
}

// From64 converts v to a Bitboard value.
func From64(v uint64) Bitboard {
	return Bitboard{
		Lo: v,
		Hi: 0,
	}
}

func FromHL(h, l uint64) Bitboard {
	return Bitboard{
		Lo: l,
		Hi: h,
	}
}

// PopCount population count.
func (u Bitboard) PopCount() uint {
	return uint(bits.OnesCount64(u.Hi) + bits.OnesCount64(u.Lo))
}

func (u Bitboard) Val64() uint64 {
	return u.Lo
}

func (u Bitboard) IsNotZero() bool {
	return u != Bitboard{}
}

// SubWrap returns u-v with wraparound semantics; for example,
// Zero.SubWrap(From64(1)) == Max.
func (u Bitboard) SubWrap(v Bitboard) Bitboard {
	lo, borrow := bits.Sub64(u.Lo, v.Lo, 0)
	hi, _ := bits.Sub64(u.Hi, v.Hi, borrow)
	return Bitboard{lo, hi}
}

// SubWrap64 returns u-v with wraparound semantics; for example,
// Zero.SubWrap64(1) == Max.
func (u Bitboard) SubWrap64(v uint64) Bitboard {
	lo, borrow := bits.Sub64(u.Lo, v, 0)
	hi := u.Hi - borrow
	return Bitboard{lo, hi}
}

// Sub64 returns u-v.
func (u Bitboard) Sub64(v uint64) Bitboard {
	lo, borrow := bits.Sub64(u.Lo, v, 0)
	hi, borrow := bits.Sub64(u.Hi, 0, borrow)
	if borrow != 0 {
		panic("underflow")
	}
	return Bitboard{lo, hi}
}

// Mul returns u*v, panicking on overflow.
func (u Bitboard) Mul(v Bitboard) Bitboard {
	hi, lo := bits.Mul64(u.Lo, v.Lo)
	p0, p1 := bits.Mul64(u.Hi, v.Lo)
	p2, p3 := bits.Mul64(u.Lo, v.Hi)
	hi, c0 := bits.Add64(hi, p1, 0)
	hi, c1 := bits.Add64(hi, p3, c0)
	if (u.Hi != 0 && v.Hi != 0) || p0 != 0 || p2 != 0 || c1 != 0 {
		panic("overflow")
	}
	return Bitboard{lo, hi}
}

// MulWrap returns u*v with wraparound semantics; for example,
// Max.MulWrap(Max) == 1.
func (u Bitboard) MulWrap(v Bitboard) Bitboard {
	hi, lo := bits.Mul64(u.Lo, v.Lo)
	hi += u.Hi*v.Lo + u.Lo*v.Hi
	return Bitboard{lo, hi}
}

// Lsh returns u<<n.
func (u Bitboard) Lsh(n uint) (s Bitboard) {
	if n > 64 {
		s.Lo = 0
		s.Hi = u.Lo << (n - 64)
	} else {
		s.Lo = u.Lo << n
		s.Hi = u.Hi<<n | u.Lo>>(64-n)
	}
	return
}

// Rsh returns u>>n.
func (u Bitboard) Rsh(n uint) (s Bitboard) {
	if n > 64 {
		s.Lo = u.Hi >> (n - 64)
		s.Hi = 0
	} else {
		s.Lo = u.Lo>>n | u.Hi<<(64-n)
		s.Hi = u.Hi >> n
	}
	return
}

// And returns u&v.
func (u Bitboard) And(v Bitboard) Bitboard {
	return Bitboard{u.Lo & v.Lo, u.Hi & v.Hi}
}

func (u Bitboard) Or(v Bitboard) Bitboard {
	return Bitboard{u.Lo | v.Lo, u.Hi | v.Hi}
}

// Xor returns u^v.
func (u Bitboard) Xor(v Bitboard) Bitboard {
	return Bitboard{u.Lo ^ v.Lo, u.Hi ^ v.Hi}
}

func (u Bitboard) Not() Bitboard {
	return Bitboard{^u.Lo, ^u.Hi}
}

func MoreThanOne(b Bitboard) bool {
	return b.And(b.SubWrap64(1)) != Bitboard{}
}

func (b Bitboard) String() string {
	s := "\n+---+---+---+---+---+---+---+---+---+\n"
	for r := RANK_9; r >= RANK_0; r-- {
		for f := FILE_A; f <= FILE_I; f++ {
			if b.And(SquareBB[MakeSquareNG(f, r)]).IsNotZero() {
				s += "| X "
			} else {
				s += "|   "
			}
		}
		s += "| " + string('0'+rune(r)) + "\n+---+---+---+---+---+---+---+---+---+\n"
	}
	s += "  a   b   c   d   e   f   g   h   i\n"
	return s
}

type (
	Square = int
	File   = int
	Rank   = int
)

const (
	SQ_A0 Square = iota
	SQ_B0
	SQ_C0
	SQ_D0
	SQ_E0
	SQ_F0
	SQ_G0
	SQ_H0
	SQ_I0

	SQ_A1
	SQ_B1
	SQ_C1
	SQ_D1
	SQ_E1
	SQ_F1
	SQ_G1
	SQ_H1
	SQ_I1

	SQ_A2
	SQ_B2
	SQ_C2
	SQ_D2
	SQ_E2
	SQ_F2
	SQ_G2
	SQ_H2
	SQ_I2

	SQ_A3
	SQ_B3
	SQ_C3
	SQ_D3
	SQ_E3
	SQ_F3
	SQ_G3
	SQ_H3
	SQ_I3

	SQ_A4
	SQ_B4
	SQ_C4
	SQ_D4
	SQ_E4
	SQ_F4
	SQ_G4
	SQ_H4
	SQ_I4

	SQ_A5
	SQ_B5
	SQ_C5
	SQ_D5
	SQ_E5
	SQ_F5
	SQ_G5
	SQ_H5
	SQ_I5

	SQ_A6
	SQ_B6
	SQ_C6
	SQ_D6
	SQ_E6
	SQ_F6
	SQ_G6
	SQ_H6
	SQ_I6

	SQ_A7
	SQ_B7
	SQ_C7
	SQ_D7
	SQ_E7
	SQ_F7
	SQ_G7
	SQ_H7
	SQ_I7

	SQ_A8
	SQ_B8
	SQ_C8
	SQ_D8
	SQ_E8
	SQ_F8
	SQ_G8
	SQ_H8
	SQ_I8

	SQ_A9
	SQ_B9
	SQ_C9
	SQ_D9
	SQ_E9
	SQ_F9
	SQ_G9
	SQ_H9
	SQ_I9

	SQ_NONE

	SQUARE_ZERO = 0
	SQUARE_NB   = 90
)

const (
	FILE_A File = iota
	FILE_B
	FILE_C
	FILE_D
	FILE_E
	FILE_F
	FILE_G
	FILE_H
	FILE_I
	FILE_NB
)

const (
	RANK_0 Rank = iota
	RANK_1
	RANK_2
	RANK_3
	RANK_4
	RANK_5
	RANK_6
	RANK_7
	RANK_8
	RANK_9
	RANK_NB
)

// Palace 九宫格
var Palace = Bitboard{
	Lo: 0xE07038,
	Hi: 0x70381C,
}

var (
	FileABB = Bitboard{
		Lo: 0x8040201008040201,
		Hi: 0x20100,
	}
	FileBBB = FileABB.Lsh(1)
	FileCBB = FileABB.Lsh(2)
	FileDBB = FileABB.Lsh(3)
	FileEBB = FileABB.Lsh(4)
	FileFBB = FileABB.Lsh(5)
	FileGBB = FileABB.Lsh(6)
	FileHBB = FileABB.Lsh(7)
	FileIBB = FileABB.Lsh(8)

	Rank0BB = From64(0x1FF)
	Rank1BB = Rank0BB.Lsh(uint(FILE_NB) * 1)
	Rank2BB = Rank0BB.Lsh(uint(FILE_NB) * 2)
	Rank3BB = Rank0BB.Lsh(uint(FILE_NB) * 3)
	Rank4BB = Rank0BB.Lsh(uint(FILE_NB) * 4)
	Rank5BB = Rank0BB.Lsh(uint(FILE_NB) * 5)
	Rank6BB = Rank0BB.Lsh(uint(FILE_NB) * 6)
	Rank7BB = Rank0BB.Lsh(uint(FILE_NB) * 7)
	Rank8BB = Rank0BB.Lsh(uint(FILE_NB) * 8)
	Rank9BB = Rank0BB.Lsh(uint(FILE_NB) * 9)
)

var (
	PawnFileBB             = FileABB.Or(FileCBB).Or(FileEBB).Or(FileGBB).Or(FileIBB)
	HalfBB     [2]Bitboard = [...]Bitboard{
		Rank0BB.Or(Rank1BB).Or(Rank2BB).Or(Rank3BB).Or(Rank4BB),
		Rank5BB.Or(Rank6BB).Or(Rank7BB).Or(Rank8BB).Or(Rank9BB),
	}
	PawnBB [2]Bitboard = [...]Bitboard{
		HalfBB[BLACK].Or((Rank3BB.Or(Rank4BB)).And(PawnFileBB)),
		HalfBB[WHITE].Or((Rank6BB.Or(Rank5BB)).And(PawnFileBB)),
	}
)

var (
	SquareBB       [SQUARE_NB]Bitboard
	SquareDistance [SQUARE_NB][SQUARE_NB]uint8

	LineBB        [SQUARE_NB][SQUARE_NB]Bitboard
	BetweenBB     [SQUARE_NB][SQUARE_NB]Bitboard
	PseudoAttacks [PIECE_TYPE_NB][SQUARE_NB]Bitboard
	PawnAttacks   [COLOR_NB][SQUARE_NB]Bitboard
	PawnAttacksTo [COLOR_NB][SQUARE_NB]Bitboard
)

func init() {
	now := time.Now()
	for s := SQ_A0; s <= SQ_I9; s++ {
		SquareBB[s] = From64(1).Lsh(uint(s))
	}
	for s1 := SQ_A0; s1 <= SQ_I9; s1++ {
		for s2 := SQ_A0; s2 <= SQ_I9; s2++ {
			SquareDistance[s1][s2] = uint8(max(DistanceF(s1, s2), DistanceR(s1, s2)))
		}
	}
	InitMagics(ROOK, RookTable[:], RookMagics[:], RookMagicsInit[:])
	InitMagics(CANNON, CannonTable[:], CannonMagics[:], RookMagicsInit[:])
	InitMagics(BISHOP, BishopTable[:], BishopMagics[:], BishopMagicsInit[:])
	InitMagics(KNIGHT, KnightTable[:], KnightMagics[:], KnightMagicsInit[:])
	InitMagics(KNIGHT_TO, KnightToTable[:], KnightToMagics[:], KnightToMagicsInit[:])
	for s1 := SQ_A0; s1 <= SQ_I9; s1++ {
		PawnAttacks[WHITE][s1] = _pawnAttacksBB(WHITE, s1)
		PawnAttacks[BLACK][s1] = _pawnAttacksBB(BLACK, s1)

		PawnAttacksTo[WHITE][s1] = _pawnAttacksToBB(WHITE, s1)
		PawnAttacksTo[BLACK][s1] = _pawnAttacksToBB(BLACK, s1)

		PseudoAttacks[ROOK][s1] = AttacksBB(ROOK, s1, From64(0))
		PseudoAttacks[BISHOP][s1] = AttacksBB(BISHOP, s1, From64(0))
		PseudoAttacks[KNIGHT][s1] = AttacksBB(KNIGHT, s1, From64(0))

		// Only generate pseudo attacks in the palace squares for king and advisor
		if Palace.And(SquareBB[s1]) != (Bitboard{}) {
			steps := []Direction{NORTH, SOUTH, WEST, EAST}
			for _, step := range steps {
				PseudoAttacks[KING][s1] = PseudoAttacks[KING][s1].Or(safeDestination(s1, step))
			}
			PseudoAttacks[KING][s1] = PseudoAttacks[KING][s1].And(Palace)
			steps = []Direction{NORTH_WEST, NORTH_EAST, SOUTH_WEST, SOUTH_EAST}
			for _, step := range steps {
				PseudoAttacks[ADVISOR][s1] = PseudoAttacks[ADVISOR][s1].Or(safeDestination(s1, step))
			}
			PseudoAttacks[ADVISOR][s1] = PseudoAttacks[ADVISOR][s1].And(Palace)
		}

		for s2 := SQ_A0; s2 <= SQ_I9; s2++ {
			if PseudoAttacks[ROOK][s1].And(SquareBB[s2]) != (Bitboard{}) {
				LineBB[s1][s2] = AttacksBB(ROOK, s1, From64(0)).
					And(AttacksBB(ROOK, s2, From64(0))).Or(SquareBB[s1]).Or(SquareBB[s2])
				BetweenBB[s1][s2] = AttacksBB(ROOK, s1, SquareBB[s2]).And(AttacksBB(ROOK, s2, SquareBB[s1]))
			}
			if PseudoAttacks[KNIGHT][s1].And(SquareBB[s2]) != (Bitboard{}) {
				BetweenBB[s1][s2] = BetweenBB[s1][s2].Or(LameLeaperPathWithDirection(KNIGHT_TO, s1, (s2 - s1)))
			}
			BetweenBB[s1][s2] = BetweenBB[s1][s2].Or(SquareBB[s2])
		}
	}
	log.Printf("cast time: %v\n", time.Since(now))
}

func FileOf(s Square) File {
	return File(s % FILE_NB)
}

func RankOf(s Square) Rank {
	return Rank(s / FILE_NB)
}

func DistanceF(x, y Square) uint {
	f1 := FileOf(x)
	f2 := FileOf(y)
	if f1 < f2 {
		return uint(f2 - f1)
	}
	return uint(f1 - f2)
}

func DistanceR(x, y Square) uint {
	r1 := RankOf(x)
	r2 := RankOf(y)
	if r1 < r2 {
		return uint(r2 - r1)
	}
	return uint(r1 - r2)
}

func Distance(x, y Square) uint {
	return uint(SquareDistance[x][y])
}

// Magic holds all magic bitboards relevant data for a single square
type Magic struct {
	mask  Bitboard
	magic Bitboard
	shift uint

	attacks unsafe.Pointer
}

func (m *Magic) Index(occupied Bitboard) uint64 {
	return occupied.And(m.mask).MulWrap(m.magic).Rsh(m.shift).Val64()
}

func SlidingAttack(sq Square, occupied Bitboard, pt PieceType) (attack Bitboard) {
	if pt != ROOK && pt != CANNON {
		panic(pt)
	}
	for _, d := range []Direction{NORTH, SOUTH, EAST, WEST} {
		hurdle := false
		for s := sq + d; IsOKSquare(s) && Distance(s-d, s) == 1; s += d {
			if pt == ROOK || hurdle {
				attack = attack.Or(SquareBB[s])
			}
			if occupied.And(SquareBB[s]) != (Bitboard{}) {
				if pt == CANNON && !hurdle {
					hurdle = true
				} else {
					break
				}
			}
		}
	}
	return
}

func safeDestination(s Square, step Direction) Bitboard {
	to := s + step
	if IsOKSquare(to) && Distance(s, to) <= 2 {
		return SquareBB[to]
	}
	return From64(0)
}

func LameLeaperPathWithDirection(pt PieceType, s Square, d Direction) Bitboard {
	if pt != KNIGHT && pt != KNIGHT_TO && pt != BISHOP {
		panic(pt)
	}
	var b Bitboard
	to := s + d
	if !IsOKSquare(to) || Distance(s, to) >= 4 {
		return b
	}
	if pt == KNIGHT_TO {
		s, to = to, s
		d = -d
	}
	var (
		dr Direction
		df Direction
	)
	if d > 0 {
		dr = NORTH
	} else {
		dr = SOUTH
	}
	if FileOf(to) > FileOf(s) {
		df = EAST
	} else {
		df = WEST
	}
	diff := abs(FileOf(to)-FileOf(s)) - abs(RankOf(to)-RankOf(s))
	if diff > 0 {
		s += df
	} else if diff < 0 {
		s += dr
	} else {
		s += df + dr
	}
	return b.Or(SquareBB[s])
}

func LameLeaperPath(pt PieceType, s Square) Bitboard {
	var b Bitboard
	var directions []Direction
	if pt == BISHOP {
		directions = []Direction{2 * NORTH_EAST, 2 * SOUTH_EAST, 2 * SOUTH_WEST, 2 * NORTH_WEST}
	} else {
		directions = []Direction{
			2*SOUTH + WEST, 2*SOUTH + EAST, SOUTH + 2*WEST, SOUTH + 2*EAST,
			NORTH + 2*WEST, NORTH + 2*EAST, 2*NORTH + WEST, 2*NORTH + EAST,
		}
	}
	for _, d := range directions {
		b = b.Or(LameLeaperPathWithDirection(pt, s, d))
	}
	if pt == BISHOP {
		if RankOf(s) > RANK_4 {
			b = b.And(HalfBB[1])
		} else {
			b = b.And(HalfBB[0])
		}
	}
	return b
}

func LameLeaperAttack(pt PieceType, s Square, occupied Bitboard) Bitboard {
	var b Bitboard
	var directions []Direction
	if pt == BISHOP {
		directions = []Direction{2 * NORTH_EAST, 2 * SOUTH_EAST, 2 * SOUTH_WEST, 2 * NORTH_WEST}
	} else {
		directions = []Direction{
			2*SOUTH + WEST, 2*SOUTH + EAST, SOUTH + 2*WEST, SOUTH + 2*EAST,
			NORTH + 2*WEST, NORTH + 2*EAST, 2*NORTH + WEST, 2*NORTH + EAST,
		}
	}
	for _, d := range directions {
		to := s + d
		if IsOKSquare(to) && Distance(s, to) < 4 && occupied.And(LameLeaperPathWithDirection(pt, s, d)) == (Bitboard{}) {
			b = b.Or(SquareBB[to])
		}
	}
	if pt == BISHOP {
		if RankOf(s) > RANK_4 {
			b = b.And(HalfBB[1])
		} else {
			b = b.And(HalfBB[0])
		}
	}
	return b
}

func rankBBR(r Rank) Bitboard {
	return Rank0BB.Lsh(uint(FILE_NB * r))
}

func rankBB(s Square) Bitboard {
	return rankBBR(RankOf(s))
}

func fileBBf(f File) Bitboard {
	return FileABB.Lsh(uint(f))
}

func fileBB(s Square) Bitboard {
	return fileBBf(FileOf(s))
}

func Shift(d Direction, b Bitboard) Bitboard {
	switch d {
	case NORTH:
		return b.And(Rank9BB.Not()).Lsh(uint(NORTH))
	case SOUTH:
		return b.Rsh(uint(NORTH))
	case EAST:
		return b.And(FileIBB.Not()).Lsh(EAST)
	case WEST:
		return b.And(FileABB.Not()).Rsh(EAST)
	case NORTH_EAST:
		return b.And(FileIBB.Not()).Lsh(uint(NORTH_EAST))
	case NORTH_WEST:
		return b.And(FileABB.Not()).Lsh(uint(NORTH_WEST))
	case SOUTH_EAST:
		return b.And(FileIBB.Not()).Rsh(uint(NORTH_WEST))
	case SOUTH_WEST:
		return b.And(FileABB.Not()).Rsh(uint(NORTH_EAST))
	default:
		return From64(0)
	}
}

func Aligned(s1, s2, s3 Square) bool {
	return LineBB[s1][s2].And(SquareBB[s3]) != Bitboard{}
}

func _pawnAttacksBB(c Color, s Square) Bitboard {
	b := SquareBB[s]
	var attack Bitboard
	if c == WHITE {
		attack = Shift(NORTH, b)
	} else {
		attack = Shift(SOUTH, b)
	}
	if (c == WHITE && RankOf(s) > RANK_4) || (c == BLACK && RankOf(s) < RANK_5) {
		attack = attack.Or(Shift(WEST, b)).Or(Shift(EAST, b))
	}
	return attack
}

func _pawnAttacksToBB(c Color, s Square) Bitboard {
	b := SquareBB[s]
	var attack Bitboard
	if c == WHITE {
		attack = Shift(SOUTH, b)
	} else {
		attack = Shift(NORTH, b)
	}
	if (c == WHITE && RankOf(s) > RANK_4) || (c == BLACK && RankOf(s) < RANK_5) {
		attack = attack.Or(Shift(WEST, b).Or(Shift(EAST, b)))
	}
	return attack
}

func AttacksBBEmptyOcc(pt PieceType, s Square) Bitboard {
	//   assert((Pt != PAWN) && (is_ok(s)));
	return PseudoAttacks[pt][s]
}

func AttacksBB(pt PieceType, s Square, occupied Bitboard) Bitboard {
	if pt == PAWN {
		panic(pt)
	}
	if !IsOKSquare(s) {
		panic(s)
	}
	switch pt {
	case ROOK:
		return *(*Bitboard)(unsafe.Pointer(uintptr(RookMagics[s].attacks) + uintptr(RookMagics[s].Index(occupied))*unsafe.Sizeof(RookTable[0])))
	case CANNON:
		return *(*Bitboard)(unsafe.Pointer(uintptr(CannonMagics[s].attacks) + uintptr(CannonMagics[s].Index(occupied))*unsafe.Sizeof(CannonTable[0])))
	case BISHOP:
		return *(*Bitboard)(unsafe.Pointer(uintptr(BishopMagics[s].attacks) + uintptr(BishopMagics[s].Index(occupied))*unsafe.Sizeof(BishopTable[0])))
	case KNIGHT:
		return *(*Bitboard)(unsafe.Pointer(uintptr(KnightMagics[s].attacks) + uintptr(KnightMagics[s].Index(occupied))*unsafe.Sizeof(KnightTable[0])))
	case KNIGHT_TO:
		return *(*Bitboard)(unsafe.Pointer(uintptr(KnightToMagics[s].attacks) + uintptr(KnightToMagics[s].Index(occupied))*unsafe.Sizeof(KnightToTable[0])))
	default:
		return PseudoAttacks[pt][s]
	}
}

var (
	RookTable     [0x108000]Bitboard // To store rook attacks
	CannonTable   [0x108000]Bitboard // To store cannon attacks
	BishopTable   [0x228]Bitboard    // To store bishop attacks
	KnightTable   [0x380]Bitboard    // To store knight attacks
	KnightToTable [0x3E0]Bitboard    // To store by knight attacks
)

var (
	RookMagics     [SQUARE_NB]Magic
	CannonMagics   [SQUARE_NB]Magic
	BishopMagics   [SQUARE_NB]Magic
	KnightMagics   [SQUARE_NB]Magic
	KnightToMagics [SQUARE_NB]Magic
)

var RookAttackMap [SQUARE_NB]map[Bitboard]Bitboard

func InitMagics(pt PieceType, table []Bitboard, magics []Magic, magicsInit []Bitboard) {
	var b Bitboard
	var edges Bitboard
	var size int
	for s := SQ_A0; s <= SQ_I9; s++ {
		edges = (Rank0BB.Or(Rank9BB).And(rankBB(s).Not())).Or(FileABB.Or(FileIBB).And(fileBB(s).Not()))
		m := &magics[s]
		if pt == ROOK {
			m.mask = SlidingAttack(s, From64(0), pt)
		} else if pt == CANNON {
			m.mask = RookMagics[s].mask
		} else {
			m.mask = LameLeaperPath(pt, s)
		}
		if pt != KNIGHT_TO {
			m.mask = m.mask.And(edges.Not())
		}
		m.shift = 128 - m.mask.PopCount()
		m.magic = magicsInit[s]

		// Set the offset for the attacks table of the square. We have individual
		// table sizes for each square with "Fancy Magic Bitboards".
		if s == SQ_A0 {
			m.attacks = unsafe.Pointer(&table[0])
		} else {
			m.attacks = unsafe.Pointer(uintptr(magics[s-1].attacks) + uintptr(size)*unsafe.Sizeof(table[0]))
		}
		b = From64(0)
		size = 0
		for {
			if pt == ROOK || pt == CANNON {
				*(*Bitboard)(unsafe.Pointer(uintptr(m.attacks) + uintptr(m.Index(b))*unsafe.Sizeof(table[0]))) = SlidingAttack(s, b, pt)
			} else {
				*(*Bitboard)(unsafe.Pointer(uintptr(m.attacks) + uintptr(m.Index(b))*unsafe.Sizeof(table[0]))) = LameLeaperAttack(pt, s, b)
			}
			size++
			b = b.SubWrap(m.mask).And(m.mask)
			if b == (Bitboard{}) {
				break
			}
		}
	}
}
