import {
  type Block,
  BOARD_HEIGHT,
  BOARD_WIDTH,
  type BoardGrid,
  type RotationState,
  type Tetromino,
  WALL_KICK_I,
  WALL_KICK_JLSTZ,
} from '@/types/tetris';

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
  const bag: Tetromino[] = ['I', 'O', 'T', 'S', 'Z', 'J', 'L'];
  while (n > 0) {
    shuffleArray(bag);
    block_list.push(...bag);
    n -= 7;
  }
  return block_list;
}
export function generateBlocks_classic(n: number): Tetromino[] {
  const block_list: Tetromino[] = [];
  const mapper: Tetromino[] = ['I', 'O', 'T', 'S', 'Z', 'J', 'L'];
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
    .map(() => Array(BOARD_WIDTH).fill({ value: 0, type: '0' }));
}

export const hasCollision = (board: BoardGrid, block: Block, cRow: number, cCol: number) => {
  let isCollision = false;
  block.shape
    .filter((row) => row.some((isSet) => isSet))
    .forEach((row, rowIndex) => {
      row.forEach((cell, colIndex) => {
        if (cell) {
          if (
            colIndex + cCol >= BOARD_WIDTH ||
            colIndex + cCol < 0 ||
            rowIndex + cRow >= BOARD_HEIGHT ||
            board[rowIndex + cRow][colIndex + cCol].value !== 0
          ) {
            isCollision = true;
          }
        }
      });
    });
  return isCollision;
};

export const findLandingPosition = (
  curBoard: BoardGrid,
  block: Block,
  cRow: number,
  cCol: number,
): number => {
  const shapeHeight = block.shape.filter((row) => row.some((isSet) => isSet)).length;

  let landingRow = cRow; // Start from current row

  for (let testRow = cRow + 1; testRow <= BOARD_HEIGHT; testRow++) {
    if (hasCollision(curBoard, block, testRow, cCol)) {
      landingRow = testRow - 1; // Dừng ở row trước khi collision
      break;
    }
    if (testRow === BOARD_HEIGHT) {
      landingRow = BOARD_HEIGHT - shapeHeight;
    }
  }

  return landingRow;
};
export const applyBlockOnBoard = (board: BoardGrid, cRow: number, cCol: number, block: Block) => {
  const landingRow = findLandingPosition(board, block, cRow, cCol);

  //render block ghost
  if (landingRow !== cRow) {
    block.shape
      .filter((row) => row.some((isSet) => isSet))
      .forEach((row, rowIndex) => {
        row.forEach((cell, colIndex) => {
          if (cell) {
            const ghostRow = landingRow + rowIndex;
            const ghostCol = cCol + colIndex;
            if (ghostRow < BOARD_HEIGHT && ghostCol < BOARD_WIDTH) {
              board[ghostRow][ghostCol] = { value: 0, type: 'ghost' };
            }
          }
        });
      });
  }

  block.shape
    .filter((row) => row.some((isSet) => isSet))
    .forEach((row, rowIndex) => {
      row.forEach((cell, colIndex) => {
        if (cell) {
          const currentRow = cRow + rowIndex;
          const currentCol = cCol + colIndex;
          if (currentRow < BOARD_HEIGHT && currentCol < BOARD_WIDTH) {
            board[currentRow][currentCol] = { value: 1, type: cell.toString() };
          }
        }
      });
    });
};
export const getWallKickData = (
  tetrominoType: Tetromino,
  fromState: RotationState,
  toState: RotationState,
): [number, number][] => {
  if (tetrominoType === 'O') return [[0, 0]]; // O không cần wall kick
  const key = `${fromState}->${toState}`;

  console.log(`Getting wall kick for ${tetrominoType}: ${key}`); // Debug
  return tetrominoType === 'I' ? WALL_KICK_I[key] : WALL_KICK_JLSTZ[key];
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
