"use client";

import { addBlockToBoard, clearLines, createEmptyBoard, findLandingPosition, getWallKickData, hasCollision, rotateRight } from "@/lib/gamelogic";

import {
  TetrominoType,
  Block,
  BoardGrid,
  BOARD_HEIGHT,
  BOARD_WIDTH,
  TETROMINO_SHAPES,
  RotationState,
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
  
};
type BoardAction = {
  type: "start" | "drop" | "commit" | "keyevent" | "end";
  payload?: {
    listBlock?: TetrominoType[];
    key?: string;
    
    committedBoard?: BoardGrid;
    rotatedBlock?: Block;
  };
};
const boardReducer = (state: BoardState, action: BoardAction): BoardState => {
  let boardState = { ...state };
  let blockType;
  switch (action.type) {
    case "start":
      blockType = action.payload?.listBlock?.[0] ?? undefined;
      boardState.activeBlock = blockType
        ? { type: blockType, shape: TETROMINO_SHAPES[blockType], rState: 0 as RotationState }
        : undefined;
      boardState.dRow = 0;
      boardState.dCol = 3;
      boardState.nextBlockIndex++;
      return boardState;
    case "drop":
      boardState.dRow++;
      break;
    case "commit":
      //check if any lines completed
      const committedBoard = action.payload?.committedBoard!;
      let clearedBoard = undefined;
      let completedLines: number[] = [];
      for (let i = 2; i < BOARD_HEIGHT; i++) {
        if (committedBoard[i].every((cell) => cell.value > 0)) {
          completedLines.push(i);
        }
      }

      //check if special clear ( t spin, perfect , back to back ...)
      
      //clear line
      if(completedLines.length>0){
         clearedBoard = clearLines(committedBoard, completedLines);
        
      }
      // continue with next block
      blockType = action.payload!.listBlock![boardState.nextBlockIndex];
      boardState.activeBlock = {
        type: blockType,
        shape: TETROMINO_SHAPES[blockType],
        rState: 0 as RotationState,
      };
      boardState.board = clearedBoard?clearedBoard:committedBoard;
      boardState.nextBlockIndex++;
      boardState.isHoldAvailable = true;
      boardState.dRow = 0;
      boardState.dCol = 3;
      return boardState;
    case "keyevent":
      if (action.payload?.key === "left") {
        let nCol = boardState.dCol - 1;
        if (!hasCollision(boardState.board, boardState.activeBlock!, nCol, boardState.dRow)) {
          boardState.dCol = nCol;
          return boardState;
        }
      } else if (action.payload?.key === "right") {
        let nCol = boardState.dCol + 1;
        if (!hasCollision(boardState.board, boardState.activeBlock!, nCol, boardState.dRow)) {
          boardState.dCol = nCol;
          return boardState;
        }
      } else if (action.payload?.key === "rotate") {
        const { rotatedBlock } = action.payload;
        const fromState = boardState.activeBlock?.rState;
        const toState = rotatedBlock?.rState;
        if (rotatedBlock) {
          if (!hasCollision(boardState.board, rotatedBlock, boardState.dCol, boardState.dRow)) {
            boardState.activeBlock = rotatedBlock;
            return boardState;
          }
          //try wallkick
          const wallKickOffsets = getWallKickData(rotatedBlock.type, fromState!, toState!);
          for (const [dx, dy] of wallKickOffsets) {
            const newDCol = boardState.dCol + dx;
            const newDRow = boardState.dRow + dy;
            if (!hasCollision(boardState.board, rotatedBlock, newDCol, newDRow)) {
              console.log(`Applying wall kick: dx=${dx}, dy=${dy}`); // Debug
              boardState.dCol=newDCol;
              boardState.dRow =newDRow;
              boardState.activeBlock = rotatedBlock;
              return boardState;
            }
          }
        }
        return boardState;
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
            rState: 0 as RotationState,
          };
          boardState.activeBlock = {
            type: blockType,
            shape: TETROMINO_SHAPES[blockType],
            rState: 0 as RotationState,
          };
          boardState.nextBlockIndex++;
          boardState.isHoldAvailable = false;
        } else if (boardState.isHoldAvailable) {
          const activeBlockType = boardState.activeBlock?.type;
          boardState.activeBlock = boardState.holdBlock;
          boardState.holdBlock = {
            type: activeBlockType!,
            shape: TETROMINO_SHAPES[activeBlockType!],
            rState: 0 as RotationState,
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

export const useBoard = (): [boardState: BoardState, dispatchBoard: React.Dispatch<BoardAction>] => {
  const [boardState, dispatchBoard] = useReducer(boardReducer, {
    board: createEmptyBoard(),
    activeBlock: undefined,
    dRow: 0,
    dCol: 0,
    isHoldAvailable: true,
    holdBlock: undefined,
    nextBlockIndex: 0,

  });

  return [boardState, dispatchBoard];
};
