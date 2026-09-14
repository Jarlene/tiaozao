(function () {
  'use strict';

  const BOARD_SIZE = 15;        // 15×15 棋盘
  const WIN_COUNT = 5;          // 五子连珠获胜
  const CELL_SIZE = 40;         // 格子边长（像素）
  const MARGIN = 30;            // 棋盘边距
  const BLACK = 1;
  const WHITE = 2;
  const PLAYER_NAME = { [BLACK]: '黑方', [WHITE]: '白方' };

  const canvas = document.getElementById('board');
  if (!canvas) return;
  const ctx = canvas.getContext('2d');

  // 棋盘状态
  let board = [];               // 0 = 空, 1 = 黑, 2 = 白
  let history = [];             // [{ row, col }] 落子记录（用于悔棋）
  let current = BLACK;          // 当前执子方
  let over = false;             // 游戏是否结束
  let lastMove = null;          // 最后落子 { row, col }

  const turnText = document.getElementById('turnText');
  const turnStone = document.querySelector('.turn-box .stone');
  const gameStatus = document.getElementById('gameStatus');
  const undoBtn = document.getElementById('undoBtn');
  const restartBtn = document.getElementById('restartBtn');

  function createEmptyBoard() {
    return Array.from({ length: BOARD_SIZE }, () => Array(BOARD_SIZE).fill(0));
  }

  function initGame() {
    board = createEmptyBoard();
    history = [];
    current = BLACK;
    over = false;
    lastMove = null;
    syncTurn();
    gameStatus.textContent = '';
    updateButtons();
    drawBoard();
  }

  function syncTurn() {
    if (over) return;
    turnStone.className = 'stone ' + (current === BLACK ? 'black' : 'white');
    turnText.textContent = PLAYER_NAME[current] + '落子';
  }

  function updateButtons() {
    undoBtn.disabled = history.length === 0;
  }

  // 计算交叉点像素坐标：第 i 条线（0..14）在 margin + i * cellSize
  function gridToPixel(index) {
    return MARGIN + index * CELL_SIZE;
  }

  // 将像素坐标转换为最近的交叉点索引，超出棋盘范围返回 -1
  function pixelToGrid(px) {
    const idx = Math.round((px - MARGIN) / CELL_SIZE);
    return idx >= 0 && idx < BOARD_SIZE ? idx : -1;
  }

  function drawBoard() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    // 木纹底色
    ctx.fillStyle = '#d9a862';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    // 网格线
    ctx.strokeStyle = '#6b4a24';
    ctx.lineWidth = 1;
    ctx.beginPath();
    for (let i = 0; i < BOARD_SIZE; i++) {
      const p = gridToPixel(i);
      ctx.moveTo(gridToPixel(0), p);
      ctx.lineTo(gridToPixel(BOARD_SIZE - 1), p);
      ctx.moveTo(p, gridToPixel(0));
      ctx.lineTo(p, gridToPixel(BOARD_SIZE - 1));
    }
    ctx.stroke();

    // 天元与四个星位
    ctx.fillStyle = '#6b4a24';
    const starPoints = [3, 7, 11];
    starPoints.forEach((r) => {
      starPoints.forEach((c) => {
        const x = gridToPixel(c);
        const y = gridToPixel(r);
        ctx.beginPath();
        ctx.arc(x, y, 4, 0, Math.PI * 2);
        ctx.fill();
      });
    });

    // 棋子
    for (let r = 0; r < BOARD_SIZE; r++) {
      for (let c = 0; c < BOARD_SIZE; c++) {
        if (board[r][c] !== 0) {
          drawStone(r, c, board[r][c]);
        }
      }
    }

    // 高亮最后落子
    if (lastMove) {
      const x = gridToPixel(lastMove.col);
      const y = gridToPixel(lastMove.row);
      ctx.strokeStyle = '#e74c3c';
      ctx.lineWidth = 2.5;
      ctx.beginPath();
      ctx.arc(x, y, CELL_SIZE / 2 - 6, 0, Math.PI * 2);
      ctx.stroke();
    }

  }

  function drawStone(row, col, player) {
    const x = gridToPixel(col);
    const y = gridToPixel(row);
    const radius = CELL_SIZE * 0.42;

    ctx.beginPath();
    ctx.arc(x, y, radius, 0, Math.PI * 2);
    if (player === BLACK) {
      ctx.fillStyle = '#1a1a1a';
      ctx.fill();
      ctx.strokeStyle = 'rgba(0,0,0,0.6)';
      ctx.lineWidth = 1;
      ctx.stroke();
    } else {
      ctx.fillStyle = '#f8f8f8';
      ctx.fill();
      ctx.strokeStyle = '#b0b0b0';
      ctx.lineWidth = 1;
      ctx.stroke();
    }

    // 高光
    ctx.beginPath();
    ctx.arc(x - radius * 0.3, y - radius * 0.35, radius * 0.18, 0, Math.PI * 2);
    ctx.fillStyle = player === BLACK ? 'rgba(255,255,255,0.35)' : 'rgba(255,255,255,0.9)';
    ctx.fill();
  }

  // 在 (row, col) 落子，返回是否合法
  function place(row, col) {
    if (over || row < 0 || row >= BOARD_SIZE || col < 0 || col >= BOARD_SIZE) return false;
    if (board[row][col] !== 0) return false; // 已占格不可重复落子

    board[row][col] = current;
    history.push({ row, col });
    lastMove = { row, col };

    if (hasWon(row, col, current)) {
      over = true;
      gameStatus.textContent = PLAYER_NAME[current] + '获胜！';
      turnText.textContent = '游戏结束';
    } else if (history.length === BOARD_SIZE * BOARD_SIZE) {
      over = true;
      gameStatus.textContent = '平局，棋盘已下满';
      turnText.textContent = '游戏结束';
    } else {
      current = current === BLACK ? WHITE : BLACK;
      syncTurn();
    }

    updateButtons();
    drawBoard();
    return true;
  }

  // 从 (row, col) 出发沿四个方向检查五连
  function hasWon(row, col, player) {
    const directions = [
      [1, 0],   // 竖直
      [0, 1],   // 水平
      [1, 1],   // 右下斜
      [1, -1],  // 左下斜
    ];
    return directions.some(([dr, dc]) => countInLine(row, col, player, dr, dc) >= WIN_COUNT);
  }

  function countInLine(row, col, player, dr, dc) {
    let count = 1;
    // 正方向
    for (let step = 1; step < WIN_COUNT; step++) {
      const r = row + dr * step;
      const c = col + dc * step;
      if (r < 0 || r >= BOARD_SIZE || c < 0 || c >= BOARD_SIZE || board[r][c] !== player) break;
      count++;
    }
    // 反方向
    for (let step = 1; step < WIN_COUNT; step++) {
      const r = row - dr * step;
      const c = col - dc * step;
      if (r < 0 || r >= BOARD_SIZE || c < 0 || c >= BOARD_SIZE || board[r][c] !== player) break;
      count++;
    }
    return count;
  }

  function undo() {
    if (over || history.length === 0) return;
    const move = history.pop();
    board[move.row][move.col] = 0;
    current = current === BLACK ? WHITE : BLACK;
    lastMove = history.length > 0 ? history[history.length - 1] : null;
    syncTurn();
    updateButtons();
    drawBoard();
  }

  // 点击/触摸事件：换算为交叉点坐标
  function getBoardCoord(event) {
    const rect = canvas.getBoundingClientRect();
    const scale = canvas.width / rect.width;
    const clientX = event.touches ? event.touches[0].clientX : event.clientX;
    const clientY = event.touches ? event.touches[0].clientY : event.clientY;
    const px = (clientX - rect.left) * scale;
    const py = (clientY - rect.top) * scale;
    return { row: pixelToGrid(py), col: pixelToGrid(px) };
  }

  canvas.addEventListener('click', (event) => {
    const { row, col } = getBoardCoord(event);
    if (row >= 0 && col >= 0) place(row, col);
  });

  canvas.addEventListener('touchstart', (event) => {
    event.preventDefault();
    const { row, col } = getBoardCoord(event);
    if (row >= 0 && col >= 0) place(row, col);
  }, { passive: false });

  undoBtn.addEventListener('click', undo);
  restartBtn.addEventListener('click', initGame);

  // 首次渲染
  initGame();
})();
