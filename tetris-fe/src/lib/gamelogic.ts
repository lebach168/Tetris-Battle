import {
  Block,
  BoardGrid,
  BOARD_HEIGHT,
  BOARD_WIDTH,
  TetrominoType,
  RotationState,
  WALL_KICK_I,
  WALL_KICK_JLSTZ,
} from "@/types/tetris";

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
  // if (level == 0) return 1000;
  const framesPerCell = Math.max(1, Math.floor((53 - level) / 2) + 1);
  return framesPerCell * (1000 / 60); // ms per cell
}
export function generateBlocks_7bag(n: number): TetrominoType[] {
  
  const block_list: TetrominoType[] = [];
  const shuffleArray = (array: TetrominoType[]) => {
    for (let i = array.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1)); // Chọn một chỉ số ngẫu nhiên
      [array[i], array[j]] = [array[j], array[i]]; // Hoán đổi hai phần tử
    }
    return array;
  };
  const bag: TetrominoType[] = ["I", "O", "T", "S", "Z", "J", "L"];
  while (n > 0) {
    shuffleArray(bag);
    block_list.push(...bag);
    n -= 7;
  }
  return block_list;
}
export function generateBlocks_classic(n: number): TetrominoType[] {
  const block_list: TetrominoType[] = [];
  const mapper: TetrominoType[] = ["I", "O", "T", "S", "Z", "J", "L"];
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
    .map(() => Array(BOARD_WIDTH).fill({ value: 0, type: "0" }));
}

export const hasCollision = (board: BoardGrid, block: Block, dCol: number, dRow: number) => {
  let isCollision = false;
  block.shape
      .filter((row) => row.some((isSet) => isSet))
      .forEach((row, rowIndex) => {
          row.forEach((cell, colIndex) => {
              if (cell) {
                  if (colIndex + dCol >= BOARD_WIDTH ||
                      colIndex + dCol < 0 ||
                      rowIndex + dRow >= BOARD_HEIGHT ||
                      board[rowIndex + dRow][colIndex + dCol].value == 1
                  ) {

                      isCollision = true;
                  }
              }

          })
      })
  return isCollision
}

export const findLandingPosition = (
  curBoard: BoardGrid,
  block: Block,
  dRow: number,
  dCol: number
): number => {
  const shapeHeight = block.shape.filter((row) =>
    row.some((isSet) => isSet)
  ).length;
  let landingRow: number = BOARD_HEIGHT - 1 - shapeHeight;
  for (let i = dRow + 1; i <= BOARD_HEIGHT; i++) {
    if (hasCollision(curBoard, block, dCol, i)) {
      landingRow = i - 1;
      break;
    }
  }
  return landingRow;
};
export const addBlockToBoard = (
  board: BoardGrid,
  dRow: number,
  dCol: number,
  block: Block
) => {
  const landingRow = findLandingPosition(board, block, dRow, dCol);
  
  //render block ghost
  if (landingRow !== dRow) {
    block.shape.filter((row) => row.some((isSet) => isSet)).forEach((row, rowIndex) => {
      row.forEach((cell, colIndex) => {
        if (cell) {
          const ghostRow = landingRow + rowIndex;
          const ghostCol = dCol + colIndex;
          if (ghostRow < BOARD_HEIGHT && ghostCol < BOARD_WIDTH) {
            board[ghostRow][ghostCol] = { value: 0, type: "ghost" };
          }
        }
      });
    });
  }

  block.shape.filter((row) => row.some((isSet) => isSet)).forEach((row, rowIndex) => {
    row.forEach((cell, colIndex) => {
      if (cell) {
        const currentRow = dRow + rowIndex;
        const currentCol = dCol + colIndex;
        if (currentRow < BOARD_HEIGHT && currentCol < BOARD_WIDTH) {
          board[currentRow][currentCol] = { value: 1, type: cell.toString() };
        }
      }
    });
  });
};
export const  getWallKickData = (
  tetrominoType: TetrominoType,
  fromState: RotationState,
  toState: RotationState
): [number, number][] =>{
  if (tetrominoType === "O") return [[0, 0]]; // O không cần wall kick
  const key = `${fromState}->${toState}`;
  
  console.log(`Getting wall kick for ${tetrominoType}: ${key}`); // Debug
  return tetrominoType === "I" ? WALL_KICK_I[key] : WALL_KICK_JLSTZ[key];
}

export const clearLines = (curBoard: BoardGrid, rows: number[]): BoardGrid => {
  const newBoard = createEmptyBoard();
  let j = BOARD_HEIGHT - 1; 

  for (let i = BOARD_HEIGHT - 1; i >= 0; i--) {
      if (rows.includes(i)) continue; 
      for (let colIndex = 0; colIndex < BOARD_WIDTH; colIndex++) {
          newBoard[j][colIndex] = curBoard[i][colIndex];
      }
      j--; 
  }

  return newBoard;
};