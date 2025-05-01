export const BOARD_WIDTH = 10;
export const BOARD_HEIGHT = 22;
export const VISIBLE_HEIGHT = 20;

export type TetrominoType = "I" | "O" | "T" | "S" | "Z" | "J" | "L";
export type Block = {
  type: TetrominoType;
  shape: number[][];
}; //shape change when rotate

export type Cell = {
  value: 1 | 0; // 1 2 3   4 for garbage.
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
export enum TickSpeed {
  Fast = 50,
  LockDelay = 300,
}
// export const TETROMINO_SHAPES = {
//   I: [[0, 0, 0, 0], [1, 1, 1, 1], [0, 0, 0, 0], [0, 0, 0, 0]],
//   O: [[1, 1], [1, 1]],
//   T: [[0, 1, 0], [1, 1, 1], [0, 0, 0]],
//   S: [[0, 2,2], [0, 2, 2], [2, 2, 0]],
//   J: [[2, 0, 0], [2, 2, 0], [0, 0, 0]],
//   Z: [[3, 3, 0], [0, 3, 3], [0, 0, 0]],
//   L: [[0, 0, 3], [3, 3, 3], [0, 0, 0]],
// };
