import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { ReverseTetrominoMap, type Tetromino, TetrominoMap } from '@/types/tetris.ts';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
export function convertTetrominoToNumArray(arr: Tetromino[]): number[] {
  return arr.map((t) => TetrominoMap[t]);
}

export function convertNumToTetrominoArray(arr: number[]): Tetromino[] {
  return arr.map((n) => ReverseTetrominoMap[n]);
}
