import {
  type Block,
  BOARD_HEIGHT,
  BOARD_WIDTH,
  type BoardGrid,
  type Tetromino,
  WALL_KICK_I,
  WALL_KICK_JLSTZ,
} from '@/types/tetris';
import type { FrameHistory, ServerState } from '@/hooks/useTetrisBattle.ts';

export function rotateRight(shape: number[][]): number[][] {
  const row = shape.length;
  const col = shape[0].length;

  const rotated = Array(row)
    .fill(null)
    .map(() => Array(col).fill(0));
  for (let i = 0; i < row; i++) {
    for (let j = 0; j < col; j++) {
      rotated[i][j] = shape[row - 1 - j][i];
    }
  }

  return rotated;
}
export function rotateLeft(shape: number[][]): number[][] {
  const row = shape.length;
  const col = shape[0].length;

  const rotated = Array(row)
    .fill(null)
    .map(() => Array(col).fill(0));
  for (let i = 0; i < row; i++) {
    for (let j = 0; j < col; j++) {
      rotated[i][j] = shape[j][col - 1 - i];
    }
  }
  return rotated;
}
export function GravityCurve(level: number): number {
  if (level == 0) return 800;
  const framesPerCell = Math.max(1, Math.floor((53 - level) / 2) + 1);
  return framesPerCell * (1000 / 60); // ms per cell
}
export function generateBlocks_7bag(n: number): Tetromino[] {
  const block_list: Tetromino[] = [];
  const shuffleArray = (array: Tetromino[]) => {
    for (let i = array.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1)); // Chọn một chỉ số ngẫu nhiên
      [array[i], array[j]] = [array[j], array[i]]; // Hoán đổi hai phần tử
    }
    return array;
  };
  const bag: Tetromino[] = [1, 2, 3, 4, 5, 6, 7];
  while (n > 0) {
    shuffleArray(bag);
    block_list.push(...bag);
    n -= 7;
  }
  return block_list;
}
export function generateBlocks_classic(n: number): Tetromino[] {
  const block_list: Tetromino[] = [];
  const mapper: Tetromino[] = [1, 2, 3, 4, 5, 6, 7];
  let previous = -1;
  let next;
  while (n > 0) {
    do {
      next = Math.floor(Math.random() * 7);
    } while (next === previous);
    block_list.push(mapper[next]);
    previous = next;
    n--;
  }
  return block_list;
}
export function createEmptyBoard(): BoardGrid {
  return Array(BOARD_HEIGHT)
    .fill(null)
    .map(() => Array(BOARD_WIDTH).fill({ value: 0 }));
}

export const hasCollision = (board: BoardGrid, block: Block, cRow: number, cCol: number) => {
  for (let r = 0; r < block.shape.length; r++) {
    for (let c = 0; c < block.shape[r].length; c++) {
      if (block.shape[r][c] !== 0) {
        const targetRow = r + cRow;
        const targetCol = c + cCol;
        if (
          targetCol >= BOARD_WIDTH ||
          targetCol < 0 ||
          targetRow >= BOARD_HEIGHT ||
          targetRow < 0 ||
          board[targetRow][targetCol].value !== 0
        ) {
          return true;
        }
      }
    }
  }
  return false;
};

export const findLandingPosition = (
  curBoard: BoardGrid,
  block: Block,
  cRow: number,
  cCol: number,
): number => {
  let landingRow = cRow;
  for (let testRow = cRow; testRow <= BOARD_HEIGHT; testRow++) {
    if (hasCollision(curBoard, block, testRow + 1, cCol)) {
      landingRow = testRow;
      break;
    }
  }
  return landingRow;
};

export const applyBlockOnBoard = (board: BoardGrid, cRow: number, cCol: number, block: Block) => {
  const landingRow = findLandingPosition(board, block, cRow, cCol);

  // --- render ghost ---
  if (landingRow !== cRow) {
    for (let r = 0; r < block.shape.length; r++) {
      for (let c = 0; c < block.shape[r].length; c++) {
        if (block.shape[r][c] !== 0) {
          const ghostRow = landingRow + r;
          const ghostCol = cCol + c;
          if (ghostRow >= 0 && ghostRow < BOARD_HEIGHT && ghostCol >= 0 && ghostCol < BOARD_WIDTH) {
            // chỉ render ghost ở ô trống
            if (board[ghostRow][ghostCol].value === 0) {
              board[ghostRow][ghostCol] = { value: 0, type: 'ghost' };
            }
          }
        }
      }
    }
  }

  // --- render block thật ---
  for (let r = 0; r < block.shape.length; r++) {
    for (let c = 0; c < block.shape[r].length; c++) {
      if (block.shape[r][c] !== 0) {
        const currentRow = cRow + r;
        const currentCol = cCol + c;
        if (
          currentRow >= 0 &&
          currentRow < BOARD_HEIGHT &&
          currentCol >= 0 &&
          currentCol < BOARD_WIDTH
        ) {
          board[currentRow][currentCol] = {
            value: block.shape[r][c],
          };
        }
      }
    }
  }
};

export const getWallKickData = (
  tetrominoType: Tetromino,
  fromState: number,
  toState: number,
): [number, number][] => {
  if (tetrominoType === 2) return [[0, 0]]; // O không cần wall kick
  const key = `${fromState}->${toState}`;
  return tetrominoType === 1 ? WALL_KICK_I[key] : WALL_KICK_JLSTZ[key];
};

/**
 * Infer tetromino type from a numeric shape matrix.
 * Scans the shape for the first non-zero value and converts it to a Tetromino
 * using the ReverseTetrominoMap. Returns undefined if no non-zero cell is found.
 */
export const getTetrominoTypeFromShape = (shape: number[][]): Tetromino => {
  for (let r = 0; r < shape.length; r++) {
    const row = shape[r];
    if (!row) continue;
    for (let c = 0; c < row.length; c++) {
      const v = row[c];
      if (v && v !== 0) {
        return v as Tetromino;
      }
    }
  }
  console.log('not found tetromino type from shape:', shape);
  return 1 as Tetromino; // default to I if not found
};

export const clearLines = (curBoard: BoardGrid) => {
  // const rows= curBoard.length;
  const cols = BOARD_WIDTH;
  const newBoard = curBoard.filter((row) => row.some((cell) => cell.value === 0));
  let j = newBoard.length - 1;
  let i = BOARD_HEIGHT - 1;
  for (; i >= 0 && j >= 0; i--, j--) {
    for (let k = 0; k < cols; k++) {
      curBoard[i][k] = newBoard[j][k];
    }
  }
  for (; i >= 0; i--) {
    curBoard[i].fill({ value: 0, type: '0' });
  }
};
export function deepCompareState(localFrame: FrameHistory, serverState: ServerState) {
  const local = localFrame.state;

  const localBoard = local.board;
  const serverBoard = serverState.boardState.board;
  if (serverBoard.length != localBoard.length) return false;
  for (let i = 0; i < localBoard.length; i++) {
    if (serverBoard[i].length !== localBoard[i].length) return false;
    for (let j = 0; j < localBoard[i].length; j++) {
      if (serverBoard[i][j].value !== localBoard[i][j].value) return false;
    }
  }
  const localBlock = local.activeBlock?.shape;
  const serverBlock = serverState.boardState.activeBlock?.shape;
  if (!serverBlock || serverBlock.length != localBlock!.length) return false;
  for (let i = 0; i < localBlock!.length; i++) {
    if (serverBlock[i].length !== localBlock![i].length) return false;
    for (let j = 0; j < localBlock![i].length; j++) {
      if (serverBlock[i][j] !== localBlock![i][j]) return false;
    }
  }
  if (
    local.cRow != serverState.boardState.cRow ||
    local.cCol != serverState.boardState.cCol ||
    local.canHold != serverState.boardState.canHold ||
    //localFrame.dropSpeed != serverState.boardState.dropSpeed ||
    local.blockIndex != serverState.boardState.blockIndex ||
    localFrame.onGround != serverState.onGround ||
    local.holdBlock != serverState.boardState.holdBlock ||
    !equalFloat(localFrame.gravityTime, serverState.gravityTime) ||
    !equalFloat(localFrame.lockTime, serverState.lockTime)
  ) {
    return false;
  }

  return true;
}
const equalFloat = (a: number, b: number, eps = 0.1) => Math.abs(a - b) <= eps;
