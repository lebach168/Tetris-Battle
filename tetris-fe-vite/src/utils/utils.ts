import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { type BoardGrid, type Tetromino, TetrominoMap } from '@/types/tetris.ts';

import type { InputBuffer } from '@/types/common.ts';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
export function convertTetrominoToNumArray(arr: Tetromino[]): number[] {
  return arr.map((t) => TetrominoMap[t]);
}

export function convertNumToTetrominoArray(arr: number[]): Tetromino[] {
  return arr.map((n) => n as Tetromino);
}
export function convertTo2DNumber(board: BoardGrid): number[][] {
  return board.map((row) => row.map((cell) => cell.value ?? 0));
}
export function convertToBoardCell(board: number[][]): BoardGrid {
  return board.map((row) => row.map((cell) => ({ value: cell })));
}
export function mapKeyToString(k: keyof InputBuffer): string {
  switch (k) {
    case 'left':
      return 'left';
    case 'right':
      return 'right';
    case 'rotate':
      return 'rotate';
    case 'rrotate':
      return 'rrotate';
    case 'down':
      return 'down';
    case 'space':
      return 'space';
    case 'hold':
      return 'hold';
    case 'downOff':
      return 'downOff';
    default:
      return '';
  }
}
