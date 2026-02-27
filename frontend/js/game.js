/**
 * Game module: snake game logic, canvas rendering, game loop.
 * Depends on globals: state, API, toast, showScreen, getSkin (from store.js)
 */

const box = 20;
const gridSize = 400;
const gridCells = gridSize / box;

const canvas = document.getElementById('game');
const ctx = canvas.getContext('2d');

function directionOffset(d) {
    if (d === 'LEFT') return [-1, 0];
    if (d === 'RIGHT') return [1, 0];
    if (d === 'UP') return [0, -1];
    if (d === 'DOWN') return [0, 1];
    return [0, 0];
}

function roundRect(ctx, x, y, w, h, r) {
    ctx.beginPath();
    ctx.moveTo(x + r, y);
    ctx.lineTo(x + w - r, y);
    ctx.quadraticCurveTo(x + w, y, x + w, y + r);
    ctx.lineTo(x + w, y + h - r);
    ctx.quadraticCurveTo(x + w, y + h, x + w - r, y + h);
    ctx.lineTo(x + r, y + h);
    ctx.quadraticCurveTo(x, y + h, x, y + h - r);
    ctx.lineTo(x, y + r);
    ctx.quadraticCurveTo(x, y, x + r, y);
    ctx.closePath();
    ctx.fill();
}

function drawSnake() {
    const skin = getSkin(state.equippedSkin);
    const isGradient = skin.gradient;
    state.snake.forEach((seg, i) => {
        const isHead = i === 0;
        let fill = isHead ? skin.head : skin.body;
        if (isGradient) {
            const t = i / Math.max(state.snake.length, 1);
            const r = Math.floor(255 * (1 - t) + 255 * t * 0.76);
            const g = Math.floor(107 * (1 - t) + 212 * t * 0.5);
            const b = Math.floor(115 * (1 - t) + 170 * t);
            fill = `rgb(${r},${g},${b})`;
        }
        ctx.fillStyle = fill;
        roundRect(ctx, seg.x + 1, seg.y + 1, box - 2, box - 2, 6);
        if (isHead) {
            ctx.fillStyle = skin.eye || '#1a1a20';
            const [dx, dy] = directionOffset(state.direction);
            const ex = seg.x + box / 2 + (dx * 6);
            const ey = seg.y + box / 2 + (dy * 6);
            ctx.beginPath();
            ctx.arc(ex - 3, ey - 3, 3, 0, Math.PI * 2);
            ctx.arc(ex + 3, ey - 3, 3, 0, Math.PI * 2);
            ctx.fill();
        }
    });
}

function placeFood() {
    const used = new Set(state.snake.map(s => `${s.x},${s.y}`));
    let x, y;
    do {
        x = (Math.floor(Math.random() * (gridCells - 2)) + 1) * box;
        y = (Math.floor(Math.random() * (gridCells - 2)) + 1) * box;
    } while (used.has(`${x},${y}`));
    state.food = { x, y };
}

function drawFood() {
    ctx.fillStyle = '#ff4757';
    roundRect(ctx, state.food.x + 2, state.food.y + 2, box - 4, box - 4, 8);
    ctx.fillStyle = 'rgba(255,255,255,0.5)';
    ctx.beginPath();
    ctx.arc(state.food.x + box / 2, state.food.y + 6, 3, 0, Math.PI * 2);
    ctx.fill();
}

function tick() {
    if (state.paused) return;
    state.nextDirection = state.nextDirection || state.direction;
    if (!state.nextDirection) return;
    const d = state.nextDirection;
    const [dx, dy] = directionOffset(d);
    const head = state.snake[0];
    const nx = head.x + dx * box;
    const ny = head.y + dy * box;
    state.direction = d;
    state.nextDirection = null;

    if (nx < 0 || nx >= gridSize || ny < 0 || ny >= gridSize) {
        loseLife();
        return;
    }
    if (state.snake.some(s => s.x === nx && s.y === ny)) {
        loseLife();
        return;
    }

    state.snake.unshift({ x: nx, y: ny });
    if (nx === state.food.x && ny === state.food.y) {
        state.score++;
        document.getElementById('score').textContent = state.score;
        placeFood();
    } else {
        state.snake.pop();
    }
}

function loseLife() {
    clearInterval(state.gameLoop);
    state.gameLoop = null;
    if (state.extraLives > 0) {
        state.extraLives--;
        state.lives++;
    }
    state.lives--;
    document.getElementById('lives').textContent = state.lives;
    if (state.lives <= 0) {
        gameOver();
        return;
    }
    state.snake = [{ x: 10 * box, y: 10 * box }];
    state.direction = null;
    placeFood();
    state.gameLoop = setInterval(gameStep, state.speed);
}

function gameOver() {
    state.gameLoop = null;
    document.getElementById('finalScore').textContent = state.score;
    if (state.score > state.highScore) {
        state.highScore = state.score;
        localStorage.setItem('snakeHighScore', String(state.highScore));
    }
    (async () => {
        try {
            const res = await apiPost('/api/earn', { score: state.score });
            const earned = res.earned ?? 0;
            state.balance = res.balance ?? state.balance;
            document.getElementById('coinsEarned').textContent = earned;
            document.getElementById('coinsEarnedLine').style.display = 'block';
        } catch (e) {
            document.getElementById('coinsEarnedLine').style.display = 'none';
        }
    })();
    document.getElementById('btnBuyLife').style.display = state.balance >= 50 ? 'block' : 'none';
    showScreen('gameOverScreen');
}

function gameStep() {
    tick();
    ctx.fillStyle = '#0a0a0c';
    ctx.fillRect(0, 0, gridSize, gridSize);
    drawFood();
    drawSnake();
}

function startGame(initialLives) {
    state.score = 0;
    if (initialLives !== undefined) {
        state.lives = initialLives;
        state.extraLives = 0;
    } else {
        state.lives = 3 + (state.extraLives || 0);
        state.extraLives = 0;
    }
    state.snake = [{ x: 10 * box, y: 10 * box }];
    state.direction = null;
    state.nextDirection = null;
    state.paused = false;
    document.getElementById('score').textContent = '0';
    document.getElementById('lives').textContent = state.lives;
    document.getElementById('pauseOverlay').classList.remove('visible');
    placeFood();
    if (state.gameLoop) clearInterval(state.gameLoop);
    state.gameLoop = setInterval(gameStep, state.speed);
    showScreen('gameScreen');
}

function togglePause() {
    state.paused = !state.paused;
    document.getElementById('pauseOverlay').classList.toggle('visible', state.paused);
}

document.addEventListener('keydown', e => {
    if (state.screen !== 'gameScreen') return;
    const arrow = ['ArrowLeft', 'ArrowUp', 'ArrowRight', 'ArrowDown'].includes(e.key);
    if (e.key === ' ' || arrow) e.preventDefault();
    if (e.key === ' ') { togglePause(); return; }
    if (state.paused) return;
    const dir = state.nextDirection || state.direction;
    if (e.key === 'ArrowLeft' && dir !== 'RIGHT') state.nextDirection = 'LEFT';
    else if (e.key === 'ArrowUp' && dir !== 'DOWN') state.nextDirection = 'UP';
    else if (e.key === 'ArrowRight' && dir !== 'LEFT') state.nextDirection = 'RIGHT';
    else if (e.key === 'ArrowDown' && dir !== 'UP') state.nextDirection = 'DOWN';
});
