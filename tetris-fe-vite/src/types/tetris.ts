export const BOARD_WIDTH = 10;
export const BOARD_HEIGHT = 22;
export const VISIBLE_HEIGHT = 20;

export type TetrominoType = "I" | "O" | "T" | "S" | "Z" | "J" | "L";
export const TetrominoMap: Record<TetrominoType, number> = {
  I: 1,
  O: 2,
  T: 3,
  S: 4,
  Z: 5,
  J: 6,
  L: 7,
};
export const ReverseTetrominoMap: Record<number, TetrominoType> = {
  1: "I",
  2: "O",
  3: "T",
  4: "S",
  5: "Z",
  6: "J",
  7: "L",
};
export type RotationState = 0 | 1 | 2 | 3; 
export type Block = {
  type: TetrominoType;
  shape: number[][];
  rState:RotationState;
}; //shape change when rotate

export type Cell = {
  value: number; // 1 2 3   4 for garbage.
  type: string; // "0" | "1" | "2" | "3" | "4" | "ghost";
};
export type BoardGrid = Cell[][];
export const TETROMINO_SHAPES: Record<TetrominoType, number[][]> = {
  I: [
    [0, 0, 0, 0],
    [1, 1, 1, 1],
    [0, 0, 0, 0],
    [0, 0, 0, 0],
  ], //I
  O: [
    [1, 1],
    [1, 1],
  ], //O
  T: [
    [0, 1, 0],
    [1, 1, 1],
    [0, 0, 0],
  ], //T
  S: [
    [0, 2, 2],
    [2, 2, 0],
    [0, 0, 0],
  ], //S
  Z: [
    [3, 3, 0],
    [0, 3, 3],
    [0, 0, 0],
  ], //Z
  J: [
    [2, 0, 0],
    [2, 2, 2],
    [0, 0, 0],
  ], //J
  L: [
    [0, 0, 3],
    [3, 3, 3],
    [0, 0, 0],
  ], //L
};
export const TickSpeed = {
  Fast: 50,
  LockDelay: 300,
} as const;
export type TickSpeed = typeof TickSpeed[keyof typeof TickSpeed];


// Wall Kick Data cho J, L, S, T, Z
export const WALL_KICK_JLSTZ: Record<string, [number, number][]> = {
  "0->1": [[0, 0], [-1, 0], [-1, 1], [0, -2], [-1, -2]], // 0->R
  "1->2": [[0, 0], [1, 0], [1, -1], [0, 2], [1, 2]], // R->2
  "2->3": [[0, 0], [1, 0], [1, 1], [0, -2], [1, -2]], // 2->L
  "3->0": [[0, 0], [-1, 0], [-1, -1], [0, 2], [-1, 2]], // L->0
  "1->0": [[0, 0], [1, 0], [1, -1], [0, 2], [1, 2]], // R->0 (Counterclockwise)
  "2->1": [[0, 0], [-1, 0], [-1, 1], [0, -2], [-1, -2]], // 2->R
  "3->2": [[0, 0], [-1, 0], [-1, -1], [0, 2], [-1, 2]], // L->2
  "0->3": [[0, 0], [1, 0], [1, 1], [0, -2], [1, -2]], // 0->L
};

// Wall Kick Data cho I
export const WALL_KICK_I: Record<string, [number, number][]> = {
  "0->1": [[0, 0], [-2, 0], [1, 0], [-2, -1], [1, 2]], // 0->R
  "1->2": [[0, 0], [-1, 0], [2, 0], [-1, 2], [2, -1]], // R->2
  "2->3": [[0, 0], [2, 0], [-1, 0], [2, 1], [-1, -2]], // 2->L
  "3->0": [[0, 0], [1, 0], [-2, 0], [1, -2], [-2, 1]], // L->0
  "1->0": [[0, 0], [2, 0], [-1, 0], [2, 1], [-1, -2]], // R->0 (Counterclockwise)
  "2->1": [[0, 0], [1, 0], [-2, 0], [1, -2], [-2, 1]], // 2->R
  "3->2": [[0, 0], [-2, 0], [1, 0], [-2, -1], [1, 2]], // L->2
  "0->3": [[0, 0], [-1, 0], [2, 0], [-1, 2], [2, -1]], // 0->L
};
export const WALL_KICK_I_AKIRA: Record<string, [number, number][]> = {
  "0->1": [[0, 0], [-2, 0], [1, 0], [-2, -1], [1, 2]], // 0->R
  "1->2": [[0, 0], [-1, 0], [2, 0], [-1, 2], [2, -1]], // R->2
  "2->3": [[0, 0], [2, 0], [-1, 0], [2, 1], [-1, -1]], // 2->L
  "3->0": [[0, 0], [1, 0], [-2, 0], [1, -2], [-2, 1]], // L->0
  "1->0": [[0, 0], [2, 0], [-1, 0], [2, 1], [-1, -2]], // R->0 (Counterclockwise)
  "2->1": [[0, 0], [1, 0], [-2, 0], [1, -2], [1, -1]], // 2->R
  "3->2": [[0, 0], [-2, 0], [1, 0], [-2, -1], [1, 2]], // L->2
  "0->3": [[0, 0], [-1, 0], [2, 0], [-1, 2], [2, -1]], // 0->L
};

export type BackToBackType = "T-Spin" | "Tetris" | "None";