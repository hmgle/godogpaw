package engine

// modify from Pikafish (https://github.com/official-pikafish/Pikafish)

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type GenType = int

const (
	CAPTURES GenType = iota
	QUIETS
	QUIET_CHECKS
	EVASIONS
	PSEUDO_LEGAL
	LEGAL
)

type Key = uint64

type Zobrist struct {
	psq  [PIECE_NB][SQUARE_NB]Key
	side Key
}

var zkey Zobrist

func init() {
	now := time.Now()
	r := rand.New(rand.NewSource(1070372))
	for pc := 0; pc < PIECE_NB; pc++ {
		for s := SQ_A0; s <= SQ_I9; s++ {
			zkey.psq[pc][s] = r.Uint64()
		}
	}
	zkey.side = r.Uint64()
	log.Printf("zkey init cast time: %v\n", time.Since(now))
}

type StateInfo struct {
	// Copied when making a move
	Material      [COLOR_NB]Value
	Check10       [COLOR_NB]int16
	Rule60        int
	PliesFromNull int

	// Not copied when making a move (will be recomputed anyhow)
	key             Key
	checkersBB      Bitboard
	blockersForKing [COLOR_NB]Bitboard
	pinners         [COLOR_NB]Bitboard
	checkSquares    [PIECE_TYPE_NB]Bitboard
	needSlowCheck   bool
	capturedPiece   Piece
	move            MoveNG
}

type StateInfoStack []*StateInfo

func NewStateInfoStack() StateInfoStack                 { return make([]*StateInfo, 0) }
func (stack StateInfoStack) Top() *StateInfo            { return stack[len(stack)-1] }
func (stack StateInfoStack) Prev() *StateInfo           { return stack[len(stack)-2] }
func (stack StateInfoStack) PrevCnt(cnt int) *StateInfo { return stack[len(stack)-cnt-1] }
func (stack *StateInfoStack) Push(st *StateInfo)        { *stack = append(*stack, st) }
func (stack *StateInfoStack) Pop() *StateInfo {
	top := (*stack)[len(*stack)-1]
	(*stack) = (*stack)[:len(*stack)-1]
	return top
}

type PositionNG struct {
	// Data members
	Board      [SQUARE_NB]Piece
	ByTypeBB   [PIECE_TYPE_NB]Bitboard
	ByColorBB  [COLOR_NB]Bitboard
	PieceCount [PIECE_NB]int
	KingSQ     [COLOR_NB]Square
	St         StateInfoStack

	SideToMove Color
	GamePly    int
	Nodes      int

	// Bloom filter for fast repetition filtering
	Filter BloomFilter

	// Board for chasing detection
	idBoard [SQUARE_NB]int

	History HistoryTable
	Evals   [MAX_MOVES]Value
	Killers [MAX_MOVES][2]MoveNG
}

func (p *PositionNG) PieceOn(s Square) Piece {
	// assert(is_ok(s));
	return p.Board[s]
}

func (p *PositionNG) Empty(s Square) bool {
	return p.Board[s] == NO_PIECE
}

func (p *PositionNG) MovedPiece(m MoveNG) Piece {
	return p.PieceOn(FromSQ(m))
}

func (p *PositionNG) PiecesAllColor(pt ...PieceType) Bitboard {
	if len(pt) == 0 {
		pt = []PieceType{ALL_PIECES}
	}
	if len(pt) == 1 {
		return p.ByTypeBB[pt[0]]
	}
	b := From64(0)
	for _, v := range pt {
		b = b.Or(p.ByTypeBB[v])
	}
	return b
}

func (p *PositionNG) Pieces(c Color, pts ...PieceType) Bitboard {
	b := p.ByColorBB[c]
	if len(pts) > 0 {
		bts := From64(0)
		for _, pt := range pts {
			bts = bts.Or(p.ByTypeBB[pt])
		}
		b = b.And(bts)
	}
	return b
}

func (p *PositionNG) Square(pt PieceType, c Color) Square {
	return Lsb(p.Pieces(c, pt))
}

func (p *PositionNG) CheckSquares(pt PieceType) Bitboard {
	return p.St.Top().checkSquares[pt]
}

func (p *PositionNG) Checkers() Bitboard {
	return p.St.Top().checkersBB
}

// CheckersTo 返回 c 方攻击 s 位置的位板
func (p *PositionNG) CheckersTo(c Color, s Square, occupied Bitboard) Bitboard {
	return (PawnAttacksTo[c][s].And(p.PiecesAllColor(PAWN)).
		Or(AttacksBB(KNIGHT_TO, s, occupied).And(p.PiecesAllColor(KNIGHT))).
		Or(AttacksBB(ROOK, s, occupied).And(p.PiecesAllColor(KING, ROOK))).
		Or(AttacksBB(CANNON, s, occupied).And(p.PiecesAllColor(CANNON)))).And(p.Pieces(c))
}

func (p *PositionNG) CheckersTo2(c Color, s Square) Bitboard {
	return p.CheckersTo(c, s, p.PiecesAllColor(ALL_PIECES))
}

func (p *PositionNG) BlockersForKing(c Color) Bitboard {
	return p.St.Top().blockersForKing[c]
}

func (p *PositionNG) blockersForKing(sliders Bitboard, s Square, pinners *Bitboard) Bitboard {
	blockers := From64(0)
	*pinners = From64(0)

	// Snipers are pieces that attack 's' when a piece and other pieces are removed
	snipers := (AttacksBBEmptyOcc(ROOK, s).And(p.PiecesAllColor(ROOK).Or(p.PiecesAllColor(CANNON).Or(p.PiecesAllColor(KING))))).
		Or((AttacksBBEmptyOcc(KNIGHT, s).And(p.PiecesAllColor(KNIGHT)))).And(sliders)
	occupancy := p.PiecesAllColor(ALL_PIECES).Xor(snipers.And(p.PiecesAllColor(CANNON).Not()))
	for snipers != (Bitboard{}) {
		sniperSq := PopLsb(&snipers)
		isCannon := TypeOf(p.PieceOn(sniperSq)) == CANNON
		var b Bitboard
		if isCannon {
			b = BetweenBB[s][sniperSq].And(p.PiecesAllColor(ALL_PIECES).Xor(SquareBB[sniperSq]))
		} else {
			b = BetweenBB[s][sniperSq].And(occupancy)
		}
		if (b != Bitboard{}) && ((!isCannon && !MoreThanOne(b)) || (isCannon && b.PopCount() == 2)) {
			blockers = blockers.Or(b)
			if b.And(p.Pieces(ColorOf(p.PieceOn(s)))) != (Bitboard{}) {
				*pinners = (*pinners).Or(SquareBB[sniperSq])
			}
		}
	}
	return blockers
}

func (pos *PositionNG) AttackersTo(s Square, occupied Bitboard) Bitboard {
	return PawnAttacksTo[WHITE][s].And(pos.Pieces(WHITE, PAWN)).
		Or(PawnAttacksTo[BLACK][s].And(pos.Pieces(BLACK, PAWN))).
		Or(AttacksBB(KNIGHT, s, occupied).And(pos.PiecesAllColor(KNIGHT))).
		Or(AttacksBB(ROOK, s, occupied).And(pos.PiecesAllColor(ROOK))).
		Or(AttacksBB(CANNON, s, occupied).And(pos.PiecesAllColor(CANNON))).
		Or(AttacksBB(BISHOP, s, occupied).And(pos.PiecesAllColor(BISHOP))).
		Or(AttacksBB(ADVISOR, s, occupied).And(pos.PiecesAllColor(ADVISOR))).
		Or(AttacksBB(KING, s, occupied).And(pos.PiecesAllColor(KING)))
}

func (pos *PositionNG) Pinners(c Color) Bitboard {
	return pos.St.Top().pinners[c]
}

func (pos *PositionNG) Legal(m MoveNG) bool {
	us := pos.SideToMove
	from := FromSQ(m)
	to := ToSQ(m)
	occupied := pos.PiecesAllColor(ALL_PIECES).Xor(SquareBB[from]).Or(SquareBB[to])
	var ksq Square
	if TypeOf(pos.MovedPiece(m)) == KING {
		ksq = to
	} else {
		ksq = pos.KingSQ[us]
	}

	// A non-king move is always legal when not moving the king or a pinned piece if we don't need slow check
	if !pos.St.Top().needSlowCheck && ksq != to && pos.BlockersForKing(us).And(SquareBB[from]) == (Bitboard{}) {
		return true
	}
	// If the moving piece is a king, check whether the destination square is
	// attacked by the opponent.
	if TypeOf(pos.PieceOn(from)) == KING {
		return pos.CheckersTo(notColor(us), to, occupied) == Bitboard{}
	}
	// A non-king move is legal if the king is not under attack after the move.
	return (pos.CheckersTo(notColor(us), ksq, occupied).And(SquareBB[to].Not())) == Bitboard{}
}

func (pos *PositionNG) PseudoLegal(m MoveNG) bool {
	us := pos.SideToMove
	from, to := FromSQ(m), ToSQ(m)
	pc := pos.MovedPiece(m)

	// If the 'from' square is not occupied by a piece belonging to the side to
	// move, the move is obviously not legal.
	if pc == NO_PIECE || ColorOf(pc) != us {
		return false
	}
	// The destination square cannot be occupied by a friendly piece
	if pos.Pieces(us).And(SquareBB[to]).IsNotZero() {
		return false
	}

	// Handle the special cases
	if TypeOf(pc) == PAWN {
		return PawnAttacks[us][from].And(SquareBB[to]).IsNotZero()
	} else if TypeOf(pc) == CANNON && !pos.Capture(m) {
		return AttacksBB(ROOK, from, pos.PiecesAllColor(ALL_PIECES)).And(SquareBB[to]).IsNotZero()
	} else {
		return AttacksBB(TypeOf(pc), from, pos.PiecesAllColor(ALL_PIECES)).And(SquareBB[to]).IsNotZero()
	}
}

func (pos *PositionNG) GivesCheck(m MoveNG) bool {
	from := FromSQ(m)
	to := ToSQ(m)
	ksq := pos.KingSQ[notColor(pos.SideToMove)]
	pt := TypeOf(pos.MovedPiece(m))

	// Is there a direct check?
	if pt == CANNON {
		if AttacksBB(CANNON, to, pos.PiecesAllColor(ALL_PIECES).Xor(SquareBB[from]).Or(SquareBB[to])).And(SquareBB[ksq]) != (Bitboard{}) {
			return true
		}
	} else if pos.CheckSquares(pt).And(SquareBB[to]) != (Bitboard{}) {
		return true
	}
	// Is there a discovered check?
	if AttacksBBEmptyOcc(ROOK, ksq).And(pos.Pieces(pos.SideToMove, CANNON)) != (Bitboard{}) {
		return pos.CheckersTo(pos.SideToMove, ksq, pos.PiecesAllColor(ALL_PIECES).Xor(SquareBB[from]).Or(SquareBB[to])).And(SquareBB[from].Not()) != (Bitboard{})
	} else if (pos.BlockersForKing(notColor(pos.SideToMove)).And(SquareBB[from]) != (Bitboard{})) && !Aligned(from, to, ksq) {
		return true
	}
	return false
}

func (pos *PositionNG) Capture(m MoveNG) bool {
	return pos.Empty(ToSQ(m))
}

func (pos *PositionNG) GenerateMoves(us Color, pt PieceType, typ GenType, movieList []MoveNG, target Bitboard) (size uint8) {
	bb := pos.Pieces(us, pt)
	for bb != (Bitboard{}) {
		from := PopLsb(&bb)
		b := From64(0)
		if pt != CANNON {
			if pt != PAWN {
				b = AttacksBB(pt, from, pos.PiecesAllColor(ALL_PIECES)).And(target)
			} else {
				b = PawnAttacks[us][from].And(target)
			}
		} else {
			// Generate cannon capture moves.
			if typ != QUIETS && typ != QUIET_CHECKS {
				b = b.Or(AttacksBB(CANNON, from, pos.PiecesAllColor(ALL_PIECES)).And(pos.Pieces(notColor(us))))
			}
			// Generate cannon quite moves.
			if typ != CAPTURES {
				b = b.Or(AttacksBB(ROOK, from, pos.PiecesAllColor(ALL_PIECES)).And(pos.PiecesAllColor(ALL_PIECES).Not()))
			}
			// Restrict to target if in evasion generation
			if typ == EVASIONS {
				b = b.And(target)
			}
		}
		// To check, you either move freely a blocker or make a direct check.
		if typ == QUIET_CHECKS {
			if pt == CANNON {
				// TODO HollowCannonDiscover
				opponentKingSquare := pos.KingSQ[notColor(us)]
				b = b.And(LineBB[from][opponentKingSquare].Not().And(pos.CheckSquares(pt)))
			} else {
				if pos.BlockersForKing(notColor(us)).And(SquareBB[from]) != (Bitboard{}) {
					opponentKingSquare := pos.KingSQ[notColor(us)]
					b = b.And(LineBB[from][opponentKingSquare].Not())
				} else {
					b = b.And(pos.CheckSquares(pt))
				}
			}
		}
		for b != (Bitboard{}) {
			movieList[size] = MakeMove(from, PopLsb(&b))
			size++
		}
	}
	return
}

func (pos *PositionNG) GenerateMovesWithoutKing(us Color, typ GenType, movieList []MoveNG, target Bitboard) (size uint8) {
	size = pos.GenerateMoves(us, ROOK, typ, movieList, target)
	size += pos.GenerateMoves(us, ADVISOR, typ, movieList[size:], target)
	size += pos.GenerateMoves(us, CANNON, typ, movieList[size:], target)
	size += pos.GenerateMoves(us, PAWN, typ, movieList[size:], target)
	size += pos.GenerateMoves(us, KNIGHT, typ, movieList[size:], target)
	size += pos.GenerateMoves(us, BISHOP, typ, movieList[size:], target)
	return
}

func (pos *PositionNG) GenerateAll(us Color, typ GenType, movieList []MoveNG) (size uint8) {
	ksq := pos.KingSQ[us]
	var target Bitboard
	if typ == PSEUDO_LEGAL {
		target = pos.Pieces(us).Not()
	} else if typ == CAPTURES {
		target = pos.Pieces(notColor(us))
	} else { // QUIETS || QUIET_CHECKS
		target = pos.PiecesAllColor(ALL_PIECES).Not()
	}
	size = pos.GenerateMovesWithoutKing(us, typ, movieList, target)

	if typ != EVASIONS && (typ != QUIET_CHECKS || pos.BlockersForKing(notColor(us)).And(SquareBB[ksq]).IsNotZero()) {
		b := AttacksBBEmptyOcc(KING, ksq).And(target)
		if typ == QUIET_CHECKS {
			opponentKingSquare := pos.KingSQ[notColor(us)]
			b = b.And(AttacksBBEmptyOcc(ROOK, opponentKingSquare))
		}
		for b != (Bitboard{}) {
			movieList[size] = MakeMove(ksq, PopLsb(&b))
			size++
		}
	}

	return
}

func (pos *PositionNG) Generate(typ GenType, movieList []MoveNG) (size uint8) {
	us := pos.SideToMove

	if typ == EVASIONS {
		return pos.GenerateEVASIONS(movieList)
	}
	if typ == LEGAL {
		return pos.GenerateLEGAL(movieList)
	}

	// TODO HollowCannonDiscover
	if us == WHITE {
		return pos.GenerateAll(WHITE, typ, movieList)
	}
	return pos.GenerateAll(BLACK, typ, movieList)
}

func (pos *PositionNG) GenerateEVASIONS(movieList []MoveNG) (size uint8) {
	// If there are more than one checker, use slow version
	if MoreThanOne(pos.Checkers()) {
		return pos.Generate(PSEUDO_LEGAL, movieList)
	}
	us := pos.SideToMove
	ksq := pos.KingSQ[us]
	checksq := Lsb(pos.Checkers())
	pt := TypeOf(pos.PieceOn(checksq))

	// Generate evasions for king, capture and non capture moves
	b := AttacksBBEmptyOcc(KING, ksq).And(pos.Pieces(us).Not())
	// For all the squares attacked by slider checkers. We will remove them from
	// the king evasions in order to skip known illegal moves, which avoids any
	// useless legality checks later on.
	if pt == ROOK || pt == CANNON {
		b = b.And(LineBB[checksq][ksq].Not().Or(pos.Pieces(notColor(us))))
	}
	for b != (Bitboard{}) {
		movieList[size] = MakeMove(ksq, PopLsb(&b))
		size++
	}

	// Generate move away hurdle piece evasions for cannon
	if pt == CANNON {
		hurdle := BetweenBB[ksq][checksq].And(pos.Pieces(us))
		if hurdle != (Bitboard{}) {
			hurdleSq := PopLsb(&hurdle)
			pt = TypeOf(pos.PieceOn(hurdleSq))
			if pt == PAWN {
				b = PawnAttacks[us][hurdleSq].And(LineBB[checksq][hurdleSq].Not().And(pos.Pieces(us).Not()))
			} else if pt == CANNON {
				b = (AttacksBB(ROOK, hurdleSq, pos.PiecesAllColor(ALL_PIECES)).And(LineBB[checksq][hurdleSq].Not()).And(pos.PiecesAllColor(ALL_PIECES).Not())).Or(
					AttacksBB(CANNON, hurdleSq, pos.PiecesAllColor(ALL_PIECES)).And(pos.Pieces(notColor(us))))
			} else {
				b = AttacksBB(pt, hurdleSq, pos.PiecesAllColor(ALL_PIECES)).And(LineBB[checksq][hurdleSq].Not()).And(pos.Pieces(us).Not())
			}
			for b != (Bitboard{}) {
				movieList[size] = MakeMove(hurdleSq, PopLsb(&b))
				size++
			}
		}
	}
	// Generate blocking evasions or captures of the checking piece
	target := BetweenBB[ksq][checksq].And(pos.Pieces(us).Not())
	if us == WHITE {
		return size + pos.GenerateMovesWithoutKing(WHITE, EVASIONS, movieList[size:], target)
	}
	return size + pos.GenerateMovesWithoutKing(BLACK, EVASIONS, movieList[size:], target)
}

func (pos *PositionNG) GenerateLEGAL(movieList []MoveNG) (size uint8) {
	if pos.Checkers() != (Bitboard{}) {
		size = pos.GenerateEVASIONS(movieList)
	} else {
		size = pos.Generate(PSEUDO_LEGAL, movieList)
	}
	cursor := uint8(0)
	allLegal := true
	for i := uint8(0); i < size; i++ {
		if pos.Legal(movieList[i]) {
			if !allLegal {
				movieList[cursor] = movieList[i]
			}
			cursor++
		} else if allLegal {
			allLegal = false
		}
	}
	return cursor
}

func (pos *PositionNG) PosIsOk() bool {
	if (pos.SideToMove != WHITE && pos.SideToMove != BLACK) ||
		pos.PieceOn(pos.Square(KING, WHITE)) != W_KING ||
		pos.PieceOn(pos.Square(KING, BLACK)) != B_KING {
		fmt.Println("pos_is_ok: Default")
		return false
	}
	if pos.PieceCount[W_KING] != 1 ||
		pos.PieceCount[B_KING] != 1 ||
		pos.CheckersTo2(pos.SideToMove, pos.Square(KING, notColor(pos.SideToMove))) != (Bitboard{}) {
		fmt.Println("pos_is_ok: Kings")
		return false
	}
	if pos.Pieces(WHITE, PAWN).And(PawnBB[WHITE].Not()).IsNotZero() ||
		pos.Pieces(BLACK, PAWN).And(PawnBB[BLACK].Not()).IsNotZero() ||
		pos.PieceCount[W_PAWN] > 5 ||
		pos.PieceCount[B_PAWN] > 5 {
		fmt.Println("pos_is_ok: Pawns")
		return false
	}
	if pos.Pieces(WHITE).And(pos.Pieces(BLACK)).IsNotZero() ||
		pos.Pieces(WHITE).Or(pos.Pieces(BLACK)) != pos.PiecesAllColor(ALL_PIECES) ||
		pos.Pieces(WHITE).PopCount() > 16 ||
		pos.Pieces(BLACK).PopCount() > 16 {
		fmt.Println("pos_is_ok: Bitboards")
		return false
	}
	for p1 := PAWN; p1 <= KING; p1++ {
		for p2 := PAWN; p2 <= KING; p2++ {
			if p1 != p2 && pos.PiecesAllColor(p1).And(pos.PiecesAllColor(p2)).IsNotZero() {
				fmt.Println("pos_is_ok: Bitboards")
				return false
			}
		}
	}

	pieces := []Piece{
		W_ROOK, W_ADVISOR, W_CANNON, W_PAWN, W_KNIGHT, W_BISHOP, W_KING,
		B_ROOK, B_ADVISOR, B_CANNON, B_PAWN, B_KNIGHT, B_BISHOP, B_KING,
	}
	for _, pc := range pieces {
		if pos.PieceCount[pc] != int(pos.Pieces(ColorOf(pc), TypeOf(pc)).PopCount()) ||
			pos.PieceCount[pc] != Count(pos.Board[:], pc) {
			fmt.Printf("pos_is_ok: Pieces[%v]\n", pc)
			return false
		}
	}

	return true
}

func parsePiece(ch rune) Piece {
	i := strings.IndexRune(" RACPNBK racpnbk", ch)
	if i <= 0 {
		i = strings.IndexRune(" RACPHEK racphek", ch)
		if i <= 0 {
			return NO_PIECE
		}
	}
	return Piece(i)
}

func (pos *PositionNG) PutPiece(pc Piece, s Square) {
	pos.Board[s] = pc
	pos.ByTypeBB[TypeOf(pc)] = pos.ByTypeBB[TypeOf(pc)].Or(SquareBB[s])
	pos.ByTypeBB[ALL_PIECES] = pos.ByTypeBB[ALL_PIECES].Or(SquareBB[s])
	pos.ByColorBB[ColorOf(pc)] = pos.ByColorBB[ColorOf(pc)].Or(SquareBB[s])
	pos.PieceCount[pc]++
	pos.PieceCount[MakePieceNG(ColorOf(pc), ALL_PIECES)]++
}

func (pos *PositionNG) RemovePiece(s Square) {
	pc := pos.Board[s]
	pos.ByTypeBB[ALL_PIECES] = pos.ByTypeBB[ALL_PIECES].Xor(SquareBB[s])
	pos.ByTypeBB[TypeOf(pc)] = pos.ByTypeBB[TypeOf(pc)].Xor(SquareBB[s])
	pos.ByColorBB[ColorOf(pc)] = pos.ByColorBB[ColorOf(pc)].Xor(SquareBB[s])
	pos.Board[s] = NO_PIECE
	pos.PieceCount[pc]--
	pos.PieceCount[MakePieceNG(ColorOf(pc), ALL_PIECES)]--
}

func (pos *PositionNG) MovePiece(from, to Square) {
	pc := pos.Board[from]
	tpc := TypeOf(pc)
	fromTo := SquareBB[from].Or(SquareBB[to])
	pos.ByTypeBB[ALL_PIECES] = pos.ByTypeBB[ALL_PIECES].Xor(fromTo)
	pos.ByTypeBB[tpc] = pos.ByTypeBB[TypeOf(pc)].Xor(fromTo)
	pos.ByColorBB[ColorOf(pc)] = pos.ByColorBB[ColorOf(pc)].Xor(fromTo)
	pos.Board[from] = NO_PIECE
	pos.Board[to] = pc
	if tpc == KING {
		pos.KingSQ[pos.SideToMove] = to
	}
}

func (pos *PositionNG) DoMove(m MoveNG, newSt *StateInfo) {
	pos.doMove(m, newSt, pos.GivesCheck(m))
}

func (pos *PositionNG) doMove(m MoveNG, newSt *StateInfo, givesCheck bool) {
	st := pos.St.Top()

	pos.Nodes++
	// Update the bloom filter
	pos.Filter.Incr(st.key)

	k := st.key ^ zkey.side
	newSt.Material = st.Material
	newSt.Check10 = st.Check10
	newSt.Rule60 = st.Rule60
	newSt.PliesFromNull = st.PliesFromNull
	pos.St.Push(newSt)
	st = newSt
	st.move = m

	// Increment ply counters. Clamp to 10 checks for each side in rule 60
	// In particular, rule60 will be reset to zero later on in case of a capture.
	pos.GamePly++
	if givesCheck {
		st.Check10[pos.SideToMove]++
	}
	if givesCheck && st.Check10[pos.SideToMove] > 10 {
		st.Rule60 -= 1
	} else {
		st.Rule60 += 1
	}
	st.PliesFromNull++

	us := pos.SideToMove
	them := notColor(us)
	from := FromSQ(m)
	to := ToSQ(m)
	pc := pos.PieceOn(from)
	captured := pos.PieceOn(to)

	if captured != NO_PIECE {
		capsq := to
		st.Material[them] -= PieceValue[MG][captured]

		// Update board and piece lists
		pos.RemovePiece(capsq)

		// Update hash key
		k ^= zkey.psq[captured][capsq]

		// Reset rule 60 counter
		st.Rule60 = 0
		st.Check10[WHITE] = 0
		st.Check10[BLACK] = 0
	}
	// Update hash key
	k ^= zkey.psq[pc][from] ^ zkey.psq[pc][to]

	pos.MovePiece(from, to)

	// Set capture piece
	st.capturedPiece = captured

	// Update the key with the final value
	st.key = k

	// Calculate checkers bitboard (if move gives check)
	if givesCheck {
		st.checkersBB = pos.CheckersTo2(us, pos.KingSQ[them])
	} else {
		st.checkersBB = From64(0)
	}

	pos.SideToMove = notColor(pos.SideToMove)

	// Update king attacks used for fast check detection
	pos.SetCheckInfo()
}

func (pos *PositionNG) UndoMove(m MoveNG) {
	pos.SideToMove = notColor(pos.SideToMove)

	from := FromSQ(m)
	to := ToSQ(m)

	pos.MovePiece(to, from) // Put the piece back at the source square

	st := pos.St.Top()
	if st.capturedPiece != NO_PIECE {
		capsq := to
		pos.PutPiece(st.capturedPiece, capsq) // Restore the captured piece
	}

	// Finally point our state pointer back to the previous state
	pos.St.Pop()
	pos.GamePly--

	// Update the bloom filter
	pos.Filter.Decr(pos.St.Top().key)
}

func (pos *PositionNG) DoNullMove(newSt *StateInfo) {
	st := pos.St.Top()

	// Update the bloom filter
	pos.Filter.Incr(st.key)

	newSt.Material = st.Material
	newSt.Check10 = st.Check10
	newSt.Rule60 = st.Rule60
	newSt.key = st.key
	newSt.checkersBB = st.checkersBB
	newSt.blockersForKing = st.blockersForKing
	newSt.pinners = st.pinners
	newSt.checkSquares = st.checkSquares
	newSt.needSlowCheck = st.needSlowCheck
	newSt.capturedPiece = st.capturedPiece
	newSt.move = st.move

	pos.St.Push(newSt)
	st = newSt
	st.key ^= zkey.side
	st.Rule60++
	st.PliesFromNull = 0
	pos.SideToMove = notColor(pos.SideToMove)
	pos.SetCheckInfo()
}

func (pos *PositionNG) UndoNullMove() {
	pos.St.Pop()
	pos.SideToMove = notColor(pos.SideToMove)

	// Update the bloom filter
	pos.Filter.Decr(pos.St.Top().key)
}

func (pos *PositionNG) Set(fenStr string) *PositionNG {
	st := new(StateInfo)
	pos.St = NewStateInfoStack()
	pos.St.Push(st)
	sq := SQ_A9

	tokens := strings.Split(fenStr, " ")
	if len(tokens) < 2 {
		log.Fatalf("bad fen: %s", fenStr)
	}
	// 1. Piece placement
	for _, token := range tokens[0] {
		if unicode.IsDigit(token) {
			sq += (int(token) - '0') * EAST
		} else if token == '/' {
			sq += 2 * SOUTH
		} else if pc := parsePiece(token); pc > 0 {
			pos.PutPiece(pc, sq)
			sq++
		} else {
			fmt.Printf("bad token: %v\n", string(token))
			sq++
		}
	}
	// 2. Active color
	if tokens[1] == "w" {
		pos.SideToMove = WHITE
	} else {
		pos.SideToMove = BLACK
	}

	if len(tokens) >= 5 {
		st.Rule60, _ = strconv.Atoi(tokens[4])
	}
	if len(tokens) >= 6 {
		pos.GamePly, _ = strconv.Atoi(tokens[5])
	}
	// Convert from fullmove starting from 1 to gamePly starting from 0,
	// handle also common incorrect FEN with fullmove = 0.
	pos.GamePly = max(2*(pos.GamePly-1), 0)
	if pos.SideToMove == BLACK {
		pos.GamePly += 1
	}
	pos.KingSQ[WHITE] = pos.Square(KING, WHITE)
	pos.KingSQ[BLACK] = pos.Square(KING, BLACK)

	pos.SetState()

	if !pos.PosIsOk() {
		log.Fatalf("pos_is_ok: fen: %s", fenStr)
	}
	return pos
}

func (pos *PositionNG) SetCheckInfo() {
	us := pos.SideToMove
	uksq := pos.KingSQ[us]
	oksq := pos.KingSQ[notColor(us)]
	st := pos.St.Top()
	st.blockersForKing[us] = pos.blockersForKing(pos.Pieces(notColor(us)), uksq, &(st.pinners[notColor(us)]))
	st.blockersForKing[notColor(us)] = pos.blockersForKing(pos.Pieces(us), oksq, &(st.pinners[us]))
	// We have to take special cares about the cannon and checks
	st.needSlowCheck = pos.Checkers().IsNotZero() || AttacksBBEmptyOcc(ROOK, uksq).And(pos.Pieces(notColor(us), CANNON)).IsNotZero()
	st.checkSquares[PAWN] = PawnAttacksTo[pos.SideToMove][oksq]
	st.checkSquares[KNIGHT] = AttacksBB(KNIGHT_TO, oksq, pos.PiecesAllColor(ALL_PIECES))
	st.checkSquares[CANNON] = AttacksBB(CANNON, oksq, pos.PiecesAllColor(ALL_PIECES))
	st.checkSquares[ROOK] = AttacksBB(ROOK, oksq, pos.PiecesAllColor(ALL_PIECES))
	st.checkSquares[BISHOP] = From64(0)
	st.checkSquares[ADVISOR] = From64(0)
	st.checkSquares[KING] = From64(0)
}

func (pos *PositionNG) SetState() {
	st := pos.St.Top()
	st.key = 0
	st.Material[WHITE] = VALUE_ZERO
	st.Material[BLACK] = VALUE_ZERO
	st.checkersBB = pos.CheckersTo2(notColor(pos.SideToMove), pos.Square(KING, pos.SideToMove))
	st.move = MOVE_NONE

	pos.SetCheckInfo()

	for b := pos.PiecesAllColor(ALL_PIECES); b != (Bitboard{}); {
		s := PopLsb(&b)
		pc := pos.PieceOn(s)
		st.key ^= zkey.psq[pc][s]
		if TypeOf(pc) != KING {
			st.Material[ColorOf(pc)] += PieceValue[MG][pc]
		}
	}
	if pos.SideToMove == BLACK {
		st.key ^= zkey.side
	}
}

func (pos *PositionNG) SeeGe(m MoveNG, threshold Value) bool {
	var occupied Bitboard
	return pos._SeeGe(m, &occupied, threshold)
}

func (pos *PositionNG) _SeeGe(m MoveNG, occupied *Bitboard, threshold Value) bool {
	from, to := FromSQ(m), ToSQ(m)

	swap := PieceValue[MG][pos.PieceOn(to)] - threshold
	if swap < 0 {
		return false
	}
	swap = PieceValue[MG][pos.PieceOn(from)] - swap
	if swap <= 0 {
		return true
	}
	*occupied = pos.PiecesAllColor(ALL_PIECES).Xor(SquareBB[from]).Xor(SquareBB[to]) // xoring to is important for pinned piece logic
	stm := pos.SideToMove
	attackers := pos.AttackersTo(to, *occupied)

	// Flying general
	if attackers.And(pos.Pieces(stm, KING)) != (Bitboard{}) {
		attackers = attackers.Or(AttacksBB(ROOK, to, (*occupied).And(pos.PiecesAllColor(ROOK).Not())).And(pos.Pieces(notColor(stm), KING)))
	}
	if attackers.And(pos.Pieces(notColor(stm), KING)) != (Bitboard{}) {
		attackers = attackers.Or(AttacksBB(ROOK, to, (*occupied).And(pos.PiecesAllColor(ROOK).Not())).And(pos.Pieces(stm, KING)))
	}
	nonCannons := attackers.And(pos.PiecesAllColor(CANNON).Not())
	cannons := attackers.And(pos.PiecesAllColor(CANNON))
	var (
		stmAttackers Bitboard
		bb           Bitboard
	)
	res := 1
	for {
		stm = notColor(stm)
		attackers = attackers.And(*occupied)

		// If stm has no more attackers then give up: stm loses
		stmAttackers = attackers.And(pos.Pieces(stm))
		if stmAttackers == (Bitboard{}) {
			break
		}

		// Don't allow pinned pieces to attack as long as there are
		// pinners on their original square.
		if pos.Pinners(notColor(stm)).And(*occupied) != (Bitboard{}) {
			stmAttackers = stmAttackers.And(pos.BlockersForKing(stm).Not())
			if stmAttackers == (Bitboard{}) {
				break
			}
		}
		res ^= 1
		// Locate and remove the next least valuable attacker, and add to the
		// bitboard 'attackers' any protential attackers when it is removed.
		bb = stmAttackers.And(pos.PiecesAllColor(PAWN))
		if bb != (Bitboard{}) {
			*occupied = (*occupied).Xor(LeastSignificantSquareBB(bb))
			swap = PawnValueMg - swap
			if swap < Value(res) {
				break
			}
			nonCannons = nonCannons.Or(AttacksBB(ROOK, to, *occupied).And(pos.PiecesAllColor(ROOK)))
			cannons = AttacksBB(CANNON, to, *occupied).And(pos.PiecesAllColor(CANNON))
			attackers = nonCannons.Or(cannons)
		} else if bb = stmAttackers.And(pos.PiecesAllColor(ADVISOR)); bb != (Bitboard{}) {
			*occupied = (*occupied).Xor(LeastSignificantSquareBB(bb))
			if swap = AdvisorValueMg - swap; swap < Value(res) {
				break
			}
			nonCannons = nonCannons.Or(AttacksBB(KNIGHT_TO, to, *occupied).And(pos.PiecesAllColor(KNIGHT)))
			attackers = nonCannons.Or(cannons)
		} else if bb = stmAttackers.And(pos.PiecesAllColor(BISHOP)); bb != (Bitboard{}) {
			*occupied = (*occupied).Xor(LeastSignificantSquareBB(bb))
			if swap = BishopValueMg - swap; swap < Value(res) {
				break
			}
		} else if bb = stmAttackers.And(pos.PiecesAllColor(CANNON)); bb != (Bitboard{}) {
			*occupied = (*occupied).Xor(LeastSignificantSquareBB(bb))
			if swap = CannonValueMg - swap; swap < Value(res) {
				break
			}
			cannons = AttacksBB(CANNON, to, *occupied).And(pos.PiecesAllColor(CANNON))
			attackers = nonCannons.Or(cannons)
		} else if bb = stmAttackers.And(pos.PiecesAllColor(KNIGHT)); bb != (Bitboard{}) {
			*occupied = (*occupied).Xor(LeastSignificantSquareBB(bb))
			if swap = KnightValueMg - swap; swap < Value(res) {
				break
			}
		} else if bb = stmAttackers.And(pos.PiecesAllColor(ROOK)); bb != (Bitboard{}) {
			*occupied = (*occupied).Xor(LeastSignificantSquareBB(bb))
			if swap = RookValueMg - swap; swap < Value(res) {
				break
			}
			nonCannons = nonCannons.Or(AttacksBB(ROOK, to, *occupied).And(pos.PiecesAllColor(ROOK)))
			cannons = AttacksBB(CANNON, to, *occupied).And(pos.PiecesAllColor(CANNON))
			attackers = nonCannons.Or(cannons)
		} else { // KING
			// If we "capture" with the king but opponent still has attackers,
			// reverse the result.
			if attackers.And(pos.Pieces(stm).Not()) != (Bitboard{}) {
				return res^1 != 0
			}
			return res != 0
		}
	}
	return res != 0
}

func (pos *PositionNG) IsRepetition() bool {
	// TODO
	st := pos.St.Top()
	if st.PliesFromNull < 7 {
		return false
	}
	for i := 0; i < 3; i++ {
		st1 := pos.St.PrevCnt(i)
		st2 := pos.St.PrevCnt(i + 2)
		if st1.key != st2.key {
			return false
		}
	}
	return true
}

func (pos *PositionNG) IsDraw() bool {
	// TODO
	if pos.St.Top().Rule60 >= 120 || pos.IsRepetition() {
		return true
	}
	return false
}

func (pos *PositionNG) Evaluate() (v Value) {
	// TODO
	return
}

func (pos *PositionNG) DetectChases(d, ply int) Value {
	// Grant each piece on board a unique id for each side
	whiteID := 0
	blackID := 0
	for s := SQ_A0; s <= SQ_I9; s++ {
		if pos.Board[s] != NO_PIECE {
			if ColorOf(pos.Board[s]) == WHITE {
				pos.idBoard[s] = whiteID
				whiteID++
			} else {
				pos.idBoard[s] = blackID
				blackID++
			}
		}
	}
	// us := pos.SideToMove
	// them := notColor(us)

	// // Rollback until we reached st - d
	// rooks := [COLOR_NB]uint16{0xFFFF, 0xFFFF}
	// chase := [COLOR_NB]uint16{0xFFFF, 0xFFFF}
	// var newChase [COLOR_NB]uint16
	// // newChase[us] = chase
	// // TODO
	return -1
}

func (pos *PositionNG) RuleJudge(result *Value, ply int) bool {
	st := pos.St.Top()
	end := min(max(0, 2*int(st.Check10[WHITE])-10)+st.Rule60+
		max(0, 2*int(st.Check10[BLACK])-10), st.PliesFromNull)

	if end >= 4 && pos.Filter.Value(st.key) >= 1 {
		cnt := 0
		stp := pos.St.PrevCnt(2)
		checkThem := st.checkersBB.And(stp.checkersBB)
		checkUs := pos.St.PrevCnt(1).checkersBB.And(pos.St.PrevCnt(3).checkersBB)
		for i := 4; i <= end; i += 2 {
			stp = pos.St.PrevCnt(i)
			checkThem = checkThem.And(stp.checkersBB)
			// Return a score if a position repeats once earlier but strictly
			// after the root, or repeats twice before or at the root.
			if stp.key == st.key {
				cnt++
			}
			if stp.key == st.key && (cnt == 2 || ply > i) {
				if checkThem == (Bitboard{}) && checkUs == (Bitboard{}) {
					// Copy the current position to a rollback struct, so we don't need to do those moves again
				}
			}
		}
	}
	return false
}

// Perft() is our utility to verify move generation. All the leaf nodes up
// to the given depth are generated and counted, and the sum is returned.
func (pos *PositionNG) Perft(depth uint, root bool) (nodes int) {
	var movieList [MAX_MOVES]MoveNG
	size := pos.GenerateLEGAL(movieList[:])
	if depth <= 1 {
		if root {
			for i, m := range movieList[:size] {
				fmt.Printf("%3d: %s%s: 1\n", i+1, squareStr(FromSQ(m)), squareStr(ToSQ(m)))
			}
		}
		return int(size)
	}
	for i := uint8(0); i < size; i++ {
		var st StateInfo
		pos.DoMove(movieList[i], &st)
		cnt := pos.Perft(depth-1, false)
		nodes += cnt
		pos.UndoMove(movieList[i])

		if root {
			fmt.Printf("%s%s: %d\n", squareStr(FromSQ(movieList[i])), squareStr(ToSQ(movieList[i])), cnt)
		}
	}
	return nodes
}

func PerftTest(depth uint, fen string) {
	var p PositionNG

	p.Set(fen)
	isOk := p.PosIsOk()
	log.Printf("fen: %s, p.PosIsOk: %+v\n", fen, isOk)

	startT := time.Now()
	nodes := p.Perft(depth, true)
	elapsed := time.Since(startT)
	log.Printf("fen: %s\n", fen)
	log.Printf("depth: %d, nodes: %d\ntime: %v\nNps: %d\n", depth, nodes, elapsed,
		int(float64(nodes)/elapsed.Seconds()))
}

func (pos *PositionNG) MoveStr(m MoveNG) (movStr string) {
	from := FromSQ(m)
	to := ToSQ(m)

	pc := pos.PieceOn(from)
	if pc == NO_PIECE {
		panic(m)
	}

	movStr += piectStr(pc) + " "
	movStr += fmt.Sprintf("%s(%d) -> %s(%d)", squareStr(from), from, squareStr(to), to)
	return
}

func squareStr(sq Square) (str string) {
	f := FileOf(sq)
	r := RankOf(sq)
	str = string('a' + rune(f))
	str += string('0' + rune(r))
	return
}

func Move2Str(m MoveNG) string {
	from := FromSQ(m)
	to := ToSQ(m)
	return squareStr(from) + squareStr(to)
}

func (pos *PositionNG) String() string {
	b := pos.PiecesAllColor(ALL_PIECES)
	s := "\n+---+---+---+---+---+---+---+---+---+\n"
	for r := RANK_9; r >= RANK_0; r-- {
		for f := FILE_A; f <= FILE_I; f++ {
			if b.And(SquareBB[MakeSquareNG(f, r)]).IsNotZero() {
				pc := pos.Board[MakeSquareNG(f, r)]
				pcStr := piectStr(pc)
				s += fmt.Sprintf("| %s", pcStr)
			} else {
				s += "|   "
			}
		}
		s += "| " + string('0'+rune(r)) + "\n+---+---+---+---+---+---+---+---+---+\n"
	}
	s += "  a   b   c   d   e   f   g   h   i\n"
	return s
}

func piectStr(pt Piece) string {
	switch pt {
	case W_ROOK:
		return "俥"
	case W_ADVISOR:
		return "仕"
	case W_CANNON:
		return "砲"
	case W_PAWN:
		return "兵"
	case W_KNIGHT:
		return "傌"
	case W_BISHOP:
		return "相"
	case W_KING:
		return "帅"

	case B_ROOK:
		return "车"
	case B_ADVISOR:
		return "士"
	case B_CANNON:
		return "炮"
	case B_PAWN:
		return "卒"
	case B_KNIGHT:
		return "马"
	case B_BISHOP:
		return "象"
	case B_KING:
		return "将"
	}
	return "NULL"
}

func pieceTypeStr(pt PieceType) string {
	switch pt {
	case ROOK:
		return "车"
	case ADVISOR:
		return "士"
	case CANNON:
		return "炮"
	case PAWN:
		return "兵"
	case KNIGHT:
		return "马"
	case BISHOP:
		return "象"
	case KING:
		return "将"
	case KNIGHT_TO:
		return "馬"
	}
	return "empty"
}
