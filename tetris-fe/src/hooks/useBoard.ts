"use client";

import {
  createEmptyBoard,
  findLandingPosition,
  hasCollision,
  generateBlocks_7bag,
} from "@/lib/gamelogic";

import {
  TetrominoType,
  Block,
  BoardGrid,
  BOARD_HEIGHT,
  BOARD_WIDTH,
  TETROMINO_SHAPES,
} from "@/types/tetris";
import { useReducer } from "react";

/* 
    Board size:
        width : 10;
        real height: 22;
        available height: 20;

*/

type BoardState = {
  board: BoardGrid;
  activeBlock?: Block;
  dCol: number; //dropping col
  dRow: number; // dropping row
  isHoldAvailable: boolean;
  holdBlock?: Block; //original shape form
  nextBlockIndex: number;
  level: number;
};
type BoardAction = {
  type: "start" | "drop" | "commit" | "keyevent" | "end";
  payload?: {
    listBlock?: TetrominoType[];
    key?: string;
    linesToClear?: number[];
  };
};
const boardReducer = (state: BoardState, action: BoardAction): BoardState => {
  let boardState = { ...state };
  let blockType;
  switch (action.type) {
    case "start":
      blockType = action.payload?.listBlock?.[0] ?? undefined;
      console.log(blockType);
      return {
        board: createEmptyBoard(),
        activeBlock: blockType
          ? { type: blockType, shape: TETROMINO_SHAPES[blockType] }
          : undefined,
        dRow: 0,
        dCol: 4,
        isHoldAvailable: true,
        holdBlock: undefined,
        nextBlockIndex: 1,
        level: 0,
      };
    case "drop":
      boardState.dRow++;
      break;
    case "commit":
      //clear line
      blockType = action.payload!.listBlock![boardState.nextBlockIndex];
      boardState.activeBlock = {
        type: blockType,
        shape: TETROMINO_SHAPES[blockType],
      };
      boardState.nextBlockIndex++;
      boardState.isHoldAvailable = true;
      boardState.dRow = 0;
      boardState.dCol = 4;
      return boardState;
    case "keyevent":
      if (action.payload?.key === "left") {
        let nCol = boardState.dCol - 1;
        if (
          !hasCollision(
            boardState.board,
            boardState.activeBlock!,
            nCol,
            boardState.dRow
          )
        ) {
          boardState.dCol = nCol;
          return boardState;
        }
      } else if (action.payload?.key === "right") {
        let nCol = boardState.dCol + 1;
        if (
          !hasCollision(
            boardState.board,
            boardState.activeBlock!,
            nCol,
            boardState.dRow
          )
        ) {
          boardState.dCol = nCol;
          return boardState;
        }
      } else if (action.payload?.key === "up") {
        // Logic for "up" key (if needed)
      } else if (action.payload?.key === "space") {
        boardState.dRow = findLandingPosition(
          boardState.board,
          boardState.activeBlock!,
          boardState.dRow,
          boardState.dCol
        );
        return boardState;
      } else if (action.payload?.key === "hold") {
        if (!boardState.holdBlock) {
          const currentBlockType = boardState.activeBlock?.type;
          blockType = action.payload!.listBlock![boardState.nextBlockIndex];
          boardState.holdBlock = {
            type: currentBlockType!,
            shape: TETROMINO_SHAPES[currentBlockType!],
          };
          boardState.activeBlock = {
            type: blockType,
            shape: TETROMINO_SHAPES[blockType],
          };
          boardState.nextBlockIndex++;
          boardState.isHoldAvailable = false;
        } else if (boardState.isHoldAvailable) {
          const activeBlockType = boardState.activeBlock?.type;
          boardState.activeBlock = boardState.holdBlock;
          boardState.holdBlock = {
            type: activeBlockType!,
            shape: TETROMINO_SHAPES[activeBlockType!],
          };
          boardState.isHoldAvailable = false;
        }
      } else {
        // Default logic (if needed)
      }

    // case "end":
  }

  return boardState;
};

export const useBoard = (): [
  boardState: BoardState,
  dispatchBoard: React.Dispatch<BoardAction>
] => {
  const [boardState, dispatchBoard] = useReducer(boardReducer, {
    board: [],
    activeBlock: undefined,
    dRow: 0,
    dCol: 0,
    isHoldAvailable: true,
    holdBlock: undefined,
    nextBlockIndex: 0,
    level: 1,
  });

  return [boardState, dispatchBoard];
};
