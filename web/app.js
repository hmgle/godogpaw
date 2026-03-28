// app.js — Game logic, Wasm loading, and UI control

let board;
let playerSide = 0; // 0 = WHITE/RED, 1 = BLACK
let aiThinking = false;
let searchDepth = 6;
let searchTimeMs = 5000;

const STATE_IDLE = 0;
const STATE_SELECTED = 1;
let interactionState = STATE_IDLE;

const REQUIRED_ENGINE_APIS = [
    'engineNewGame',
    'engineGetBoard',
    'engineGetLegalMovesFrom',
    'engineDoMoveBySquares',
    'engineUndoMove',
    'engineSearch',
];

function getMissingEngineApis() {
    return REQUIRED_ENGINE_APIS.filter(name => typeof window[name] !== 'function');
}

async function initWasm() {
    const go = new Go();
    const wasmUrl = new URL('godogpaw.wasm', document.baseURI);
    const result = await WebAssembly.instantiateStreaming(
        fetch(wasmUrl),
        go.importObject
    );

    let runtimeError = null;
    go.run(result.instance).catch(err => {
        runtimeError = err;
        console.error('Go runtime error:', err);
        setStatus('Engine crashed during startup.');
    });

    // Give Go a tick to publish its JS bridge functions.
    await new Promise(resolve => setTimeout(resolve, 0));

    if (runtimeError) {
        throw runtimeError;
    }

    const missing = getMissingEngineApis();
    if (missing.length > 0) {
        throw new Error(`Engine API missing after startup: ${missing.join(', ')}`);
    }
}

function setStatus(text) {
    document.getElementById('status-text').textContent = text;
}

function setThinking(on) {
    aiThinking = on;
    document.getElementById('app').classList.toggle('thinking', on);
    document.getElementById('btn-undo').disabled = on;
    document.getElementById('btn-new-game').disabled = on;
    // Disable/enable difficulty controls
    document.querySelectorAll('.preset-btn').forEach(b => b.disabled = on);
    document.querySelectorAll('#custom-panel input').forEach(el => el.disabled = on);
    if (on) {
        setStatus('AI thinking...');
    }
}

function refreshBoard() {
    const stateJson = engineGetBoard();
    const state = JSON.parse(stateJson);

    board.update(state.board, state.lastMoveFrom, state.lastMoveTo);

    if (state.isGameOver) {
        // The side to move has no legal moves — they lose in Xiangqi
        if (state.sideToMove === playerSide) {
            setStatus('Game over — you lose!');
        } else {
            setStatus('Game over — you win!');
        }
        return true;
    }

    if (state.inCheck) {
        setStatus(state.sideToMove === playerSide ? 'Check! Your turn.' : 'Check!');
    } else {
        setStatus(state.sideToMove === playerSide ? 'Your turn.' : '');
    }
    return false;
}

async function aiMove() {
    setThinking(true);
    // Yield to browser so it can repaint the board (showing the human move)
    // before the CPU-intensive WASM search blocks the main thread.
    await new Promise(r => setTimeout(r, 0));
    try {
        const moveStr = await engineSearch(searchDepth, searchTimeMs);
        if (!moveStr) {
            setStatus('AI has no moves — you win!');
            setThinking(false);
            return;
        }
    } catch (e) {
        console.error('AI search error:', e);
    }
    setThinking(false);
    board.clearSelection();
    const gameOver = refreshBoard();
    if (gameOver) return;
}

function startNewGame() {
    playerSide = document.getElementById('sel-side').value === 'black' ? 1 : 0;
    board.flipped = playerSide === 1;

    engineNewGame('');
    board.clearSelection();
    interactionState = STATE_IDLE;
    refreshBoard();

    // If player is black, AI moves first
    if (playerSide === 1) {
        aiMove();
    }
}

function undoMove() {
    if (aiThinking) return;
    // Undo two moves: AI move + player move
    const ok1 = engineUndoMove();
    if (!ok1) return;
    const ok2 = engineUndoMove();
    // If only one undo worked (start of game), that's fine
    board.clearSelection();
    interactionState = STATE_IDLE;
    refreshBoard();
}

function handleBoardClick(e) {
    if (aiThinking) return;

    const rect = board.canvas.getBoundingClientRect();
    const scaleX = board.canvas.width / rect.width;
    const scaleY = board.canvas.height / rect.height;
    const px = (e.clientX - rect.left) * scaleX;
    const py = (e.clientY - rect.top) * scaleY;
    const sq = board.canvasToSq(px, py);

    if (sq < 0) {
        board.clearSelection();
        interactionState = STATE_IDLE;
        return;
    }

    const stateJson = engineGetBoard();
    const state = JSON.parse(stateJson);

    // Not player's turn
    if (state.sideToMove !== playerSide) return;
    if (state.isGameOver) return;

    const pc = state.board[sq];
    const pcSide = pc === 0 ? -1 : (pc <= 7 ? 0 : 1);

    if (interactionState === STATE_IDLE) {
        // Click own piece to select
        if (pcSide === playerSide) {
            selectPiece(sq);
        }
    } else if (interactionState === STATE_SELECTED) {
        if (pcSide === playerSide) {
            // Click another own piece — switch selection
            selectPiece(sq);
        } else if (board.legalTargets.includes(sq)) {
            // Execute move
            const from = board.selectedSq;
            const ok = engineDoMoveBySquares(from, sq);
            if (ok) {
                board.clearSelection();
                interactionState = STATE_IDLE;
                const gameOver = refreshBoard();
                if (!gameOver) {
                    aiMove();
                }
            }
        } else {
            // Click empty non-target: deselect
            board.clearSelection();
            interactionState = STATE_IDLE;
        }
    }
}

function selectPiece(sq) {
    const targetsJson = engineGetLegalMovesFrom(sq);
    const targets = JSON.parse(targetsJson);
    board.setSelected(sq, targets);
    interactionState = STATE_SELECTED;
}

function initDifficultyControls() {
    const presetBtns = document.querySelectorAll('.preset-btn:not(#btn-custom)');
    const btnCustom = document.getElementById('btn-custom');
    const customPanel = document.getElementById('custom-panel');
    const sliderDepth = document.getElementById('slider-depth');
    const sliderTime = document.getElementById('slider-time');
    const valDepth = document.getElementById('val-depth');
    const valTime = document.getElementById('val-time');

    function setActivePreset(btn) {
        document.querySelectorAll('.preset-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
    }

    function updateTimeLabel(secs) {
        valTime.textContent = secs === 0 ? '不限时' : secs + '秒';
    }

    presetBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            setActivePreset(btn);
            searchDepth = parseInt(btn.dataset.depth, 10);
            searchTimeMs = parseInt(btn.dataset.time, 10);
            customPanel.classList.add('hidden');
            // Sync sliders with preset values
            sliderDepth.value = searchDepth;
            valDepth.textContent = searchDepth;
            sliderTime.value = Math.round(searchTimeMs / 1000);
            updateTimeLabel(Math.round(searchTimeMs / 1000));
        });
    });

    btnCustom.addEventListener('click', () => {
        setActivePreset(btnCustom);
        customPanel.classList.remove('hidden');
    });

    sliderDepth.addEventListener('input', () => {
        searchDepth = parseInt(sliderDepth.value, 10);
        valDepth.textContent = searchDepth;
    });

    sliderTime.addEventListener('input', () => {
        const secs = parseInt(sliderTime.value, 10);
        searchTimeMs = secs * 1000;
        updateTimeLabel(secs);
    });
}

// Init
(async function main() {
    const canvas = document.getElementById('board-canvas');
    board = new BoardRenderer(canvas);

    setStatus('Loading engine...');
    try {
        await initWasm();
    } catch (e) {
        console.error('Engine init error:', e);
        setStatus('Engine failed to load.');
        return;
    }
    setStatus('Engine ready.');

    canvas.addEventListener('click', handleBoardClick);
    document.getElementById('btn-new-game').addEventListener('click', startNewGame);
    document.getElementById('btn-undo').addEventListener('click', undoMove);
    initDifficultyControls();

    startNewGame();
})();
