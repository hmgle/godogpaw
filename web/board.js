// board.js — Canvas-based Xiangqi board rendering

const BOARD_COLS = 9;
const BOARD_ROWS = 10;
const CELL_SIZE = 57;
const PADDING = 30;
const PIECE_RADIUS = 24;

// Piece type constants matching engine encoding
const NO_PIECE = 0;
const W_ROOK = 1, W_ADVISOR = 2, W_CANNON = 3, W_PAWN = 4, W_KNIGHT = 5, W_BISHOP = 6, W_KING = 7;
const B_ROOK = 9, B_ADVISOR = 10, B_CANNON = 11, B_PAWN = 12, B_KNIGHT = 13, B_BISHOP = 14, B_KING = 15;

// Chinese characters for each piece
const PIECE_CHARS = {
    [W_ROOK]: '俥', [W_ADVISOR]: '仕', [W_CANNON]: '砲', [W_PAWN]: '兵',
    [W_KNIGHT]: '傌', [W_BISHOP]: '相', [W_KING]: '帅',
    [B_ROOK]: '车', [B_ADVISOR]: '士', [B_CANNON]: '炮', [B_PAWN]: '卒',
    [B_KNIGHT]: '马', [B_BISHOP]: '象', [B_KING]: '将',
};

function pieceColor(pc) {
    if (pc >= 1 && pc <= 7) return 'white'; // RED side (WHITE internally)
    if (pc >= 9 && pc <= 15) return 'black';
    return null;
}

class BoardRenderer {
    constructor(canvas) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d');
        this.flipped = false;
        this.selectedSq = -1;
        this.legalTargets = [];
        this.lastFrom = -1;
        this.lastTo = -1;
        this.boardData = new Array(90).fill(0);

        // Set canvas size
        this.canvas.width = PADDING * 2 + (BOARD_COLS - 1) * CELL_SIZE;
        this.canvas.height = PADDING * 2 + (BOARD_ROWS - 1) * CELL_SIZE;
    }

    // Convert engine square (file + rank*9) to canvas coordinates
    sqToCanvas(sq) {
        const file = sq % 9;
        const rank = Math.floor(sq / 9);
        let cx, cy;
        if (this.flipped) {
            cx = PADDING + (8 - file) * CELL_SIZE;
            cy = PADDING + rank * CELL_SIZE;
        } else {
            cx = PADDING + file * CELL_SIZE;
            cy = PADDING + (9 - rank) * CELL_SIZE;
        }
        return { x: cx, y: cy };
    }

    // Convert canvas pixel to engine square, or -1
    canvasToSq(px, py) {
        let bestDist = Infinity;
        let bestSq = -1;
        for (let sq = 0; sq < 90; sq++) {
            const { x, y } = this.sqToCanvas(sq);
            const dist = Math.sqrt((px - x) ** 2 + (py - y) ** 2);
            if (dist < PIECE_RADIUS + 4 && dist < bestDist) {
                bestDist = dist;
                bestSq = sq;
            }
        }
        return bestSq;
    }

    draw() {
        const ctx = this.ctx;
        const w = this.canvas.width;
        const h = this.canvas.height;

        // Background
        ctx.fillStyle = '#e8c97a';
        ctx.fillRect(0, 0, w, h);

        ctx.strokeStyle = '#4a3520';
        ctx.lineWidth = 1;

        // Draw grid lines
        for (let r = 0; r < BOARD_ROWS; r++) {
            const y = PADDING + r * CELL_SIZE;
            ctx.beginPath();
            ctx.moveTo(PADDING, y);
            ctx.lineTo(PADDING + (BOARD_COLS - 1) * CELL_SIZE, y);
            ctx.stroke();
        }
        for (let c = 0; c < BOARD_COLS; c++) {
            const x = PADDING + c * CELL_SIZE;
            if (c === 0 || c === BOARD_COLS - 1) {
                ctx.beginPath();
                ctx.moveTo(x, PADDING);
                ctx.lineTo(x, PADDING + (BOARD_ROWS - 1) * CELL_SIZE);
                ctx.stroke();
            } else {
                // Top half
                ctx.beginPath();
                ctx.moveTo(x, PADDING);
                ctx.lineTo(x, PADDING + 4 * CELL_SIZE);
                ctx.stroke();
                // Bottom half
                ctx.beginPath();
                ctx.moveTo(x, PADDING + 5 * CELL_SIZE);
                ctx.lineTo(x, PADDING + 9 * CELL_SIZE);
                ctx.stroke();
            }
        }

        // Draw palace diagonals
        this.drawPalaceDiagonals(ctx);

        // Draw river text
        this.drawRiver(ctx);

        // Draw last move highlight
        if (this.lastFrom >= 0 && this.lastTo >= 0) {
            this.drawSquareHighlight(ctx, this.lastFrom, 'rgba(255, 200, 0, 0.4)');
            this.drawSquareHighlight(ctx, this.lastTo, 'rgba(255, 200, 0, 0.4)');
        }

        // Draw selected square highlight
        if (this.selectedSq >= 0) {
            this.drawSquareHighlight(ctx, this.selectedSq, 'rgba(0, 180, 0, 0.35)');
        }

        // Draw legal target dots
        for (const tgt of this.legalTargets) {
            const { x, y } = this.sqToCanvas(tgt);
            ctx.beginPath();
            ctx.arc(x, y, 8, 0, Math.PI * 2);
            if (this.boardData[tgt] !== NO_PIECE) {
                // Capture: ring instead of dot
                ctx.lineWidth = 3;
                ctx.strokeStyle = 'rgba(0, 180, 0, 0.7)';
                ctx.stroke();
            } else {
                ctx.fillStyle = 'rgba(0, 180, 0, 0.5)';
                ctx.fill();
            }
        }

        // Draw pieces
        for (let sq = 0; sq < 90; sq++) {
            const pc = this.boardData[sq];
            if (pc === NO_PIECE) continue;
            this.drawPiece(ctx, sq, pc);
        }
    }

    drawPalaceDiagonals(ctx) {
        // In the engine: rank 0-2 is red palace, rank 7-9 is black palace
        // Palace columns: d,e,f (files 3,4,5)
        const palaces = [
            { rank0: 0, rank2: 2 }, // Red palace: rank 0,1,2
            { rank0: 7, rank2: 9 }, // Black palace: rank 7,8,9
        ];
        ctx.strokeStyle = '#4a3520';
        ctx.lineWidth = 1;
        for (const p of palaces) {
            // sq = file + rank * 9
            const tl = 3 + p.rank0 * 9; // d,rank0
            const tr = 5 + p.rank0 * 9; // f,rank0
            const bl = 3 + p.rank2 * 9; // d,rank2
            const br = 5 + p.rank2 * 9; // f,rank2
            const c1 = this.sqToCanvas(tl);
            const c2 = this.sqToCanvas(br);
            const c3 = this.sqToCanvas(tr);
            const c4 = this.sqToCanvas(bl);
            ctx.beginPath(); ctx.moveTo(c1.x, c1.y); ctx.lineTo(c2.x, c2.y); ctx.stroke();
            ctx.beginPath(); ctx.moveTo(c3.x, c3.y); ctx.lineTo(c4.x, c4.y); ctx.stroke();
        }
    }

    drawRiver(ctx) {
        const y = PADDING + 4.5 * CELL_SIZE;
        ctx.font = '22px serif';
        ctx.fillStyle = '#4a3520';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        if (this.flipped) {
            ctx.fillText('楚 河', PADDING + 2 * CELL_SIZE, y);
            ctx.fillText('汉 界', PADDING + 6 * CELL_SIZE, y);
        } else {
            ctx.fillText('楚 河', PADDING + 2 * CELL_SIZE, y);
            ctx.fillText('汉 界', PADDING + 6 * CELL_SIZE, y);
        }
    }

    drawSquareHighlight(ctx, sq, color) {
        const { x, y } = this.sqToCanvas(sq);
        ctx.fillStyle = color;
        ctx.fillRect(x - CELL_SIZE / 2, y - CELL_SIZE / 2, CELL_SIZE, CELL_SIZE);
    }

    drawPiece(ctx, sq, pc) {
        const { x, y } = this.sqToCanvas(sq);
        const isRed = pieceColor(pc) === 'white';

        // Shadow
        ctx.beginPath();
        ctx.arc(x + 1, y + 2, PIECE_RADIUS, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(0,0,0,0.15)';
        ctx.fill();

        // Piece circle
        ctx.beginPath();
        ctx.arc(x, y, PIECE_RADIUS, 0, Math.PI * 2);
        ctx.fillStyle = '#f5e6c8';
        ctx.fill();
        ctx.lineWidth = 2;
        ctx.strokeStyle = isRed ? '#c00' : '#222';
        ctx.stroke();

        // Inner circle
        ctx.beginPath();
        ctx.arc(x, y, PIECE_RADIUS - 4, 0, Math.PI * 2);
        ctx.strokeStyle = isRed ? '#c00' : '#222';
        ctx.lineWidth = 1;
        ctx.stroke();

        // Character
        const ch = PIECE_CHARS[pc] || '?';
        ctx.font = 'bold 22px "KaiTi", "STKaiti", "SimSun", serif';
        ctx.fillStyle = isRed ? '#c00' : '#222';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(ch, x, y + 1);
    }

    update(boardArr, lastFrom, lastTo) {
        this.boardData = boardArr;
        this.lastFrom = lastFrom;
        this.lastTo = lastTo;
        this.draw();
    }

    setSelected(sq, targets) {
        this.selectedSq = sq;
        this.legalTargets = targets || [];
        this.draw();
    }

    clearSelection() {
        this.selectedSq = -1;
        this.legalTargets = [];
        this.draw();
    }
}
