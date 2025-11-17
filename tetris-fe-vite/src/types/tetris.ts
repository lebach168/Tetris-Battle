import { GravityCurve } from '@/utils/gamelogic.ts';
import type { InputBuffer } from '@/types/common.ts';

export const BOARD_WIDTH = 10;
export const BOARD_HEIGHT = 22;
export const VISIBLE_HEIGHT = 20;

export type Tetromino = 1 | 2 | 3 | 4 | 5 | 6 | 7;
export const TetrominoMap: Record<string, number> = {
  I: 1,
  O: 2,
  T: 3,
  Z: 4,
  L: 5,
  S: 6,
  J: 7,
};
export const ReverseTetrominoMap: Record<number, string> = {
  1: 'I',
  2: 'O',
  3: 'T',
  4: 'Z',
  5: 'L',
  6: 'S',
  7: 'J',
};

export type Block = {
  type: Tetromino;
  shape: number[][];
  form: number;
}; //shape change when rotate

export type Cell = {
  value: number; // 1->8... 8 for garbage.
  type?: string; //"ghost";
};
export type BoardGrid = Cell[][];
export const TETROMINO_SHAPES: Record<Tetromino, number[][]> = {
  1: [
    [0, 0, 0, 0],
    [1, 1, 1, 1],
    [0, 0, 0, 0],
    [0, 0, 0, 0],
  ], //I
  2: [
    [2, 2],
    [2, 2],
  ], //O
  3: [
    [0, 3, 0],
    [3, 3, 3],
    [0, 0, 0],
  ], //T
  4: [
    [4, 4, 0],
    [0, 4, 4],
    [0, 0, 0],
  ], //Z
  5: [
    [0, 0, 5],
    [5, 5, 5],
    [0, 0, 0],
  ], //L
  6: [
    [0, 6, 6],
    [6, 6, 0],
    [0, 0, 0],
  ], //S
  7: [
    [7, 0, 0],
    [7, 7, 7],
    [0, 0, 0],
  ], //J
};
export const DropSpeed = {
  DropLevel: (level: number) => {
    return GravityCurve(level);
  },
  SoftDrop: 100,
} as const;
export const LockDelay = 300;
// Wall Kick Data cho J, L, S, T, Z
export const WALL_KICK_JLSTZ: Record<string, [number, number][]> = {
  '0->1': [
    [0, 0],
    [-1, 0],
    [-1, 1],
    [0, -2],
    [-1, -2],
  ], // 0->R
  '1->2': [
    [0, 0],
    [1, 0],
    [1, -1],
    [0, 2],
    [1, 2],
  ], // R->2
  '2->3': [
    [0, 0],
    [1, 0],
    [1, 1],
    [0, -2],
    [1, -2],
  ], // 2->L
  '3->0': [
    [0, 0],
    [-1, 0],
    [-1, -1],
    [0, 2],
    [-1, 2],
  ], // L->0
  '1->0': [
    [0, 0],
    [1, 0],
    [1, -1],
    [0, 2],
    [1, 2],
  ], // R->0 (Counterclockwise)
  '2->1': [
    [0, 0],
    [-1, 0],
    [-1, 1],
    [0, -2],
    [-1, -2],
  ], // 2->R
  '3->2': [
    [0, 0],
    [-1, 0],
    [-1, -1],
    [0, 2],
    [-1, 2],
  ], // L->2
  '0->3': [
    [0, 0],
    [1, 0],
    [1, 1],
    [0, -2],
    [1, -2],
  ], // 0->L
};

// Wall Kick Data cho I
export const WALL_KICK_I: Record<string, [number, number][]> = {
  '0->1': [
    [0, 0],
    [-2, 0],
    [1, 0],
    [-2, -1],
    [1, 2],
  ], // 0->R
  '1->2': [
    [0, 0],
    [-1, 0],
    [2, 0],
    [-1, 2],
    [2, -1],
  ], // R->2
  '2->3': [
    [0, 0],
    [2, 0],
    [-1, 0],
    [2, 1],
    [-1, -2],
  ], // 2->L
  '3->0': [
    [0, 0],
    [1, 0],
    [-2, 0],
    [1, -2],
    [-2, 1],
  ], // L->0
  '1->0': [
    [0, 0],
    [2, 0],
    [-1, 0],
    [2, 1],
    [-1, -2],
  ], // R->0 (Counterclockwise)
  '2->1': [
    [0, 0],
    [1, 0],
    [-2, 0],
    [1, -2],
    [-2, 1],
  ], // 2->R
  '3->2': [
    [0, 0],
    [-2, 0],
    [1, 0],
    [-2, -1],
    [1, 2],
  ], // L->2
  '0->3': [
    [0, 0],
    [-1, 0],
    [2, 0],
    [-1, 2],
    [2, -1],
  ], // 0->L
};
export const WALL_KICK_I_AKIRA: Record<string, [number, number][]> = {
  '0->1': [
    [0, 0],
    [-2, 0],
    [1, 0],
    [-2, -1],
    [1, 2],
  ], // 0->R
  '1->2': [
    [0, 0],
    [-1, 0],
    [2, 0],
    [-1, 2],
    [2, -1],
  ], // R->2
  '2->3': [
    [0, 0],
    [2, 0],
    [-1, 0],
    [2, 1],
    [-1, -1],
  ], // 2->L
  '3->0': [
    [0, 0],
    [1, 0],
    [-2, 0],
    [1, -2],
    [-2, 1],
  ], // L->0
  '1->0': [
    [0, 0],
    [2, 0],
    [-1, 0],
    [2, 1],
    [-1, -2],
  ], // R->0 (Counterclockwise)
  '2->1': [
    [0, 0],
    [1, 0],
    [-2, 0],
    [1, -2],
    [1, -1],
  ], // 2->R
  '3->2': [
    [0, 0],
    [-2, 0],
    [1, 0],
    [-2, -1],
    [1, 2],
  ], // L->2
  '0->3': [
    [0, 0],
    [-1, 0],
    [2, 0],
    [-1, 2],
    [2, -1],
  ], // 0->L
};

export type BackToBackType = 'T-Spin' | 'Tetris' | 'None';
export const keys: Array<keyof InputBuffer> = [
  'left',
  'right',
  'rotate',
  'rrotate',
  'down',
  'space',
  'hold',
  'downOff',
];
