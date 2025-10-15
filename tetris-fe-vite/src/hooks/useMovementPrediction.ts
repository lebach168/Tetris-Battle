'use client';

import {
  type Block,
  type BoardGrid,
  type RotationState,
  type Tetromino,
  TETROMINO_SHAPES,
} from '@/types/tetris';
import { type Dispatch, useReducer } from 'react';
import {
  applyBlockOnBoard,
  clearLines,
  createEmptyBoard,
  findLandingPosition,
  generateBlocks_7bag,
  getWallKickData,
  hasCollision,
  rotateLeft,
  rotateRight,
} from '@/utils/gamelogic.ts';
/* 
    Board size:
        width : 10;
        real height: 22;
        available height: 20;

*/
export type BoardState = {
  board: BoardGrid;
  activeBlock?: Block;
  cCol: number; //current col
  cRow: number; // current row
  isHoldAvailable: boolean;
  holdBlock?: Block; //original shape form
  nextBlockIndex: number;
  listBlock?: Tetromino[];
};
type BoardAction = {
  type: 'start' | 'drop' | 'commit' | 'key_event' | 'end';
  payload?: {
    key?: string;
    committedBoard?: BoardGrid;
    // rotatedBlock?: Block;
  };
};
function boardReducer(state: BoardState, action: BoardAction): BoardState {
  const newState = { ...state };
  if (action.type === 'key_event') {
    handleKeyEvent(newState, action);
  } else if (action.type === 'drop') {
    newState.cRow++;
  } else if (action.type === 'commit') {
    if (action.payload?.committedBoard) {
      clearLines(action.payload.committedBoard);
      console.log(action.payload.committedBoard);
      newState.board = action.payload.committedBoard;
    }
    const nextBlock = state.listBlock![state.nextBlockIndex];
    newState.nextBlockIndex++;
    newState.activeBlock = {
      type: nextBlock,
      shape: TETROMINO_SHAPES[nextBlock],
      rState: 0,
    };
    newState.cCol = 4;
    newState.cRow = 0;
    newState.isHoldAvailable = true;

    //TODO clear line

    //TODO check special move : t spin, back to back ...
  } else if (action.type === 'end') {
    //
  } else if (action.type === 'start') {
    newState.listBlock = generateBlocks_7bag(1000);
    newState.nextBlockIndex = 1;
    newState.cRow = 0;
    newState.cCol = 4;
    newState.activeBlock = {
      type: newState.listBlock[0],
      shape: TETROMINO_SHAPES[newState.listBlock[0]],
      rState: 0,
    };
  }

  return newState;
}

function handleKeyEvent(state: BoardState, action: BoardAction) {
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
        rotateBlock.rState = ((rotateBlock.rState + delta) % 4) as RotationState;
        if (!hasCollision(state.board, rotateBlock, state.cRow, state.cCol)) {
          state.activeBlock = rotateBlock;
          break;
        }
        const fromState = state.activeBlock?.rState;
        const toState = rotateBlock?.rState;
        const wallKickOffsets = getWallKickData(rotateBlock.type, fromState!, toState!);
        for (const [dx, dy] of wallKickOffsets) {
          const newDCol = state.cCol + dx;
          const newDRow = state.cRow + dy;
          if (!hasCollision(state.board, rotateBlock, newDRow, newDCol)) {
            console.log(`Applying wall kick: dx=${dx}, dy=${dy}`); // Debug
            state.cCol = newDCol;
            state.cRow = newDRow;
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
        const committed = structuredClone(state.board);
        applyBlockOnBoard(committed, landingRow, state.cCol, state.activeBlock);
        clearLines(committed);
        state.board = committed;
        const nextBlock = state.listBlock![state.nextBlockIndex];
        state.nextBlockIndex++;
        state.activeBlock = {
          type: nextBlock,
          shape: TETROMINO_SHAPES[nextBlock],
          rState: 0,
        };
        state.cCol = 4;
        state.cRow = 0;
        state.isHoldAvailable = true;
      }
      break;
    case 'hold':
      if (state.isHoldAvailable && state.activeBlock) {
        const hold = state.activeBlock.type;
        if (state.holdBlock) {
          state.activeBlock = state.holdBlock;
        } else {
          const nextBlock = state.listBlock![state.nextBlockIndex];
          state.activeBlock = { type: nextBlock, shape: TETROMINO_SHAPES[nextBlock], rState: 0 };
          state.nextBlockIndex++;
        }
        state.holdBlock = { type: hold, shape: TETROMINO_SHAPES[hold], rState: 0 };
        state.isHoldAvailable = false;
      }
      break;
  }
}
export const usePredictMovement = (): {
  boardState: BoardState;
  dispatchAction: Dispatch<BoardAction>;
} => {
  const [boardState, dispatch] = useReducer(boardReducer, {
    board: createEmptyBoard(),
    activeBlock: undefined,
    cCol: 0, //current col
    cRow: 0, // current row
    isHoldAvailable: true,
    holdBlock: undefined,
    nextBlockIndex: -1,
    listBlock: [],
  });
  return {
    boardState,
    dispatchAction: dispatch,
  };
};
