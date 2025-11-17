'use client';

import { type Block, type BoardGrid, type Tetromino, TETROMINO_SHAPES } from '@/types/tetris';
import { type RefObject, useRef } from 'react';
import {
  clearLines,
  createEmptyBoard,
  findLandingPosition,
  generateBlocks_7bag,
  getWallKickData,
  hasCollision,
  rotateLeft,
  rotateRight,
} from '@/utils/gamelogic.ts';

export type BoardState = {
  board: BoardGrid;
  activeBlock?: Block;
  cCol: number;
  cRow: number;
  canHold: boolean;
  holdBlock: number; //tetromino type  - 0 is empty value
  blockIndex: number;
  listBlock?: Tetromino[];
};

type BoardAction = {
  type: 'start' | 'drop' | 'commit' | 'key_event' | 'end';
  payload?: {
    key?: string;
    committedBoard?: BoardGrid;
    listBlock?: Tetromino[];
  };
};

export const useMovementPrediction = (): {
  boardStateRef: RefObject<BoardState>;
  applyAction: (action: BoardAction) => void;
} => {
  const boardStateRef = useRef<BoardState>({
    board: createEmptyBoard(),
    activeBlock: undefined,
    cCol: 4,
    cRow: 0,
    canHold: true,
    holdBlock: 0,
    blockIndex: 0,
    listBlock: [],
  });

  const handleKeyEvent = (action: BoardAction) => {
    const state = boardStateRef.current;

    switch (action.payload?.key) {
      case 'left':
        if (!hasCollision(state.board, state.activeBlock!, state.cRow, state.cCol - 1)) {
          state.cCol--;
        }
        break;
      case 'right':
        if (!hasCollision(state.board, state.activeBlock!, state.cRow, state.cCol + 1)) {
          state.cCol++;
        }
        break;
      case 'rotate_right':
      case 'rotate_left':
        if (state.activeBlock) {
          const rotateBlock = structuredClone(state.activeBlock);
          const isRight = action.payload?.key === 'rotate_right';

          rotateBlock.shape = isRight
            ? rotateRight(rotateBlock.shape)
            : rotateLeft(rotateBlock.shape);

          const delta = isRight ? 1 : 3;
          rotateBlock.form = (rotateBlock.form + delta) % 4;
          if (!hasCollision(state.board, rotateBlock, state.cRow, state.cCol)) {
            state.activeBlock = rotateBlock;
            break;
          }
          const fromState = state.activeBlock?.form;
          const toState = rotateBlock?.form;
          const wallKickOffsets = getWallKickData(rotateBlock.type, fromState!, toState!);
          for (const [dx, dy] of wallKickOffsets) {
            const newCol = state.cCol + dx;
            const newRow = state.cRow + dy;
            if (!hasCollision(state.board, rotateBlock, newRow, newCol)) {
              console.log(`Applying wall kick: dx=${dx}, dy=${dy}`);
              state.cCol = newCol;
              state.cRow = newRow;
              state.activeBlock = rotateBlock;
              break;
            }
          }
        }
        break;

      case 'space':
        if (state.activeBlock) {
          const landingRow = findLandingPosition(
            state.board,
            state.activeBlock,
            state.cRow,
            state.cCol,
          );
          state.cRow = landingRow;
        }
        break;
      case 'hold':
        if (state.canHold && state.activeBlock) {
          const hold = state.activeBlock.type;
          if (state.holdBlock != 0) {
            state.activeBlock = {
              type: state.holdBlock as Tetromino,
              shape: TETROMINO_SHAPES[state.holdBlock as Tetromino],
              form: 0,
            };
          } else {
            state.blockIndex++;
            const nextBlock = state.listBlock![state.blockIndex];
            state.activeBlock = { type: nextBlock, shape: TETROMINO_SHAPES[nextBlock], form: 0 };
          }
          state.holdBlock = hold;
          state.canHold = false;
          state.cRow = 0;
          state.cCol = 4;
        }
        break;
    }
  };

  const applyAction = (action: BoardAction) => {
    const state = boardStateRef.current;

    if (action.type === 'key_event') {
      handleKeyEvent(action);
    } else if (action.type === 'drop') {
      if (!hasCollision(state.board, state.activeBlock!, state.cRow + 1, state.cCol)) {
        state.cRow++;
      }
    } else if (action.type === 'commit') {
      if (action.payload?.committedBoard) {
        clearLines(action.payload.committedBoard);
        state.board = action.payload.committedBoard;
      }
      state.blockIndex++;
      const nextBlock = state.listBlock![state.blockIndex];
      state.activeBlock = {
        type: nextBlock,
        shape: TETROMINO_SHAPES[nextBlock],
        form: 0,
      };
      state.cCol = 4;
      state.cRow = 0;
      state.canHold = true;
    } else if (action.type === 'start') {
      console.log('list block : ', action.payload?.listBlock);
      state.listBlock = action.payload?.listBlock?.length
        ? action.payload.listBlock
        : generateBlocks_7bag(1000);

      state.blockIndex = 0;
      state.cRow = 0;
      state.cCol = 4;
      if (state.listBlock[0]) {
        state.activeBlock = {
          type: state.listBlock[0],
          shape: TETROMINO_SHAPES[state.listBlock[0]],
          form: 0,
        };
      }
    }
  };

  return {
    boardStateRef,
    applyAction,
  };
};
