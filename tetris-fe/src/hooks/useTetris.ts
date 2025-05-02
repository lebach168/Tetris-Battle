"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useBoard } from "./useBoard";
import { Block, BOARD_HEIGHT, BoardGrid, RotationState, TETROMINO_SHAPES, TetrominoType, TickSpeed, VISIBLE_HEIGHT } from "@/types/tetris";
import {
  hasCollision,
  addBlockToBoard,
  generateBlocks_7bag,
  GravityCurve,
  rotateLeft,
  rotateRight,
  createEmptyBoard,
} from "@/lib/gamelogic";

const TickSpeedType = {
  DropLevel: (level: number) => {
    return GravityCurve(level);
  },
  SoftDrop: TickSpeed.Fast,
  LockDelay: TickSpeed.LockDelay,
};
export const useTetris = (): {
  board: BoardGrid;
  startGame: () => void;
  isPlaying: boolean;
} => {
  const level =1;//temp
  const [{ board, activeBlock, dRow, dCol, isHoldAvailable, holdBlock, nextBlockIndex}, dispatchBoard] =
    useBoard();
  const [isPlaying, setIsPlaying] = useState(false);
  const [tickSpeed, setTickSpeed] = useState<TickSpeed | null>(TickSpeedType.DropLevel(level));
  const [isCommitting, setIsCommitting] = useState(false);

  const inputBuffer = useRef<InputBuffer>({
    left: false,
    right: false,
    up: false,
    rotateCounterClockwise:false,
    down: false,
    space: false,
    hold: false,
    // botActions: [],
  });
  const lastFrameTime = useRef<number>(0);
  const accumulatedTime = useRef<number>(0);
  const animationFrameId = useRef<number | null>(null);
  const listBlock = useRef<TetrominoType[]>([]);
  //start game

  const handleInputEvent = useCallback(() => {
    const buffer = inputBuffer.current;
    if (buffer.down) {
      setTickSpeed(TickSpeed.Fast);
    }
    if (buffer.left) {
      dispatchBoard({ type: "keyevent", payload: { key: "left" } });
      buffer.left = false;
    }
    if (buffer.right) {
      dispatchBoard({ type: "keyevent", payload: { key: "right" } });
      buffer.right = false;
    }
    if (buffer.up) {
      const rotatedShape = rotateRight(activeBlock?.shape!);
      const rotatedBlock:Block = {type:activeBlock!.type,shape:rotatedShape,rState:(activeBlock?.rState!+1)%4 as RotationState}
      dispatchBoard({ type: "keyevent", payload: { key: "rotate", rotatedBlock} });
      buffer.up = false;
    }
    if(buffer.rotateCounterClockwise){
      const rotatedShape = rotateLeft(activeBlock?.shape!);
      const rotatedBlock:Block = {type:activeBlock!.type,shape:rotatedShape,rState:(activeBlock?.rState!-1)%4 as RotationState}
      dispatchBoard({ type: "keyevent", payload: { key: "rotate", rotatedBlock } });
      buffer.rotateCounterClockwise = false;
    }
    if (buffer.space) {
      dispatchBoard({ type: "keyevent", payload: { key: "space" } });
      buffer.space = false;
    }
    if (buffer.hold) {
      dispatchBoard({ type: "keyevent", payload: { key: "hold", listBlock: listBlock.current } });
      buffer.hold = false;
    }
  }, [activeBlock, dispatchBoard]);

  const renderBoard = useCallback(() => {
    const renderedBoard = structuredClone(board) as BoardGrid;
    if (isPlaying && activeBlock) {
      addBlockToBoard(renderedBoard, dRow, dCol, activeBlock);
    }
    return renderedBoard.slice(2, 2 + VISIBLE_HEIGHT);
  }, [board, isPlaying, activeBlock, dRow, dCol]);

  /*
    Game flow

  */
  const startGame = useCallback(() => {
    setIsPlaying(true);
    const generated = generateBlocks_7bag(1000);
    listBlock.current = generated;
    dispatchBoard({
      type: "start",
      payload: { listBlock: generated },
    });
  }, []);

  //commit a block
  const commitPosition = useCallback(() => {
    if (!activeBlock || !hasCollision(board, activeBlock, dCol, dRow + 1)) {
      setIsCommitting(false);
      setTickSpeed(TickSpeedType.DropLevel(0));
      return;
    }
    let cloneBoard = structuredClone(board) as BoardGrid;
    addBlockToBoard(cloneBoard, dRow, dCol, activeBlock);

    setTickSpeed(TickSpeedType.DropLevel(0));
    
    dispatchBoard({
      type: "commit",
      payload: { listBlock: listBlock.current, committedBoard: cloneBoard },
    });

    checkGameover(cloneBoard);
    setIsCommitting(false);
  }, [board, activeBlock, dRow, dCol, nextBlockIndex, holdBlock, dispatchBoard]);

  const checkGameover = (board: BoardGrid) => {
    const nextBlockType: TetrominoType = listBlock.current[nextBlockIndex];
    const nextBlock = nextBlockType ? { type: nextBlockType, shape: TETROMINO_SHAPES[nextBlockType] , rState:0 as RotationState} : undefined;
    if (nextBlock && hasCollision(board, nextBlock, 4, 0)) {
      dispatchBoard({ type: "end" });
      setIsPlaying(false);
      setTickSpeed(null);
      // Thông báo game over (tùy chọn)
      console.log("Game Over!");
    }
  };
  const gameTick = useCallback(() => {
    if (isCommitting) {
      commitPosition();
    } else if (activeBlock && hasCollision(board, activeBlock, dCol, dRow + 1)) {
      setTickSpeed(TickSpeedType.LockDelay);
      setIsCommitting(true);
    } else {
      dispatchBoard({ type: "drop" });
      //dispatchGameAction({ type: "soft_drop", cells: 1 });
    }
  }, [board, activeBlock, dCol, dRow, isCommitting, commitPosition, dispatchBoard]);

  //game loop
  const gameLoop = useCallback(
    (timestamp: number) => {
      if (!isPlaying) {
        if (animationFrameId.current) {
          cancelAnimationFrame(animationFrameId.current);
        }
        animationFrameId.current = null;
        return;
      }

      const deltaTime = timestamp - lastFrameTime.current;
      lastFrameTime.current = timestamp;
      accumulatedTime.current += deltaTime;

      const currentTickSpeed = tickSpeed ? tickSpeed : TickSpeedType.DropLevel(level);
      if (accumulatedTime.current >= currentTickSpeed) {
        gameTick();
        accumulatedTime.current -= currentTickSpeed;
      }

      handleInputEvent(); //1 frame ~16ms, thời gian để bấm giữa 2 phím ~40-100ms (1 giây bấm được 10-15 phím) -> mỗi frame chỉ cần kiểm tra input 1 lần

      animationFrameId.current = requestAnimationFrame(gameLoop);
    },
    [isPlaying, tickSpeed, gameTick, handleInputEvent]
  );

  useEffect(() => {
    if (isPlaying && !animationFrameId.current) {
      lastFrameTime.current = performance.now();
      accumulatedTime.current = 0;
      animationFrameId.current = requestAnimationFrame(gameLoop);
    }
    return () => {
      if (animationFrameId.current) {
        cancelAnimationFrame(animationFrameId.current);
        animationFrameId.current = null;
      }
    };
  }, [isPlaying, gameLoop]);
  //hande event listener
  useEffect(() => {
    if (!isPlaying) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.repeat) return;

      if (event.key === " ") {
        event.preventDefault();
        inputBuffer.current.space = true;
      }
      if (event.key === "ArrowDown") inputBuffer.current.down = true;
      if (event.key === "ArrowLeft") inputBuffer.current.left = true;
      if (event.key === "ArrowRight") inputBuffer.current.right = true;
      if (event.key === "ArrowUp"||event.key === "x") inputBuffer.current.up = true;
      if (event.key === "z") inputBuffer.current.rotateCounterClockwise = true;
      if (event.key === "c") inputBuffer.current.hold = true;
    };

    const handleKeyUp = (event: KeyboardEvent) => {
      if (event.key === "ArrowDown") {
        inputBuffer.current.down = false;
        setTickSpeed(TickSpeedType.DropLevel(level));
      }
      if (event.key === "ArrowLeft") inputBuffer.current.left = false;
      if (event.key === "ArrowRight") inputBuffer.current.right = false;
    };

    document.addEventListener("keydown", handleKeyDown);
    document.addEventListener("keyup", handleKeyUp);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.removeEventListener("keyup", handleKeyUp);
      inputBuffer.current = {
        left: false,
        right: false,
        up: false,
        rotateCounterClockwise:false,
        down: false,
        space: false,
        hold: false,
      };
    };
  }, [isPlaying]);

  return {
    board:isPlaying ? renderBoard() : createEmptyBoard().slice(2, 2 + VISIBLE_HEIGHT),
    startGame,
    isPlaying,
  };
};
