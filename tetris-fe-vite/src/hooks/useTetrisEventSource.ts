"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useBoard } from "./useBoard";
import {
  Block,
  BOARD_HEIGHT,
  BoardGrid,
  RotationState,
  TETROMINO_SHAPES,
  TetrominoType,
  TickSpeed,
  VISIBLE_HEIGHT,
} from "@/types/tetris";
import {
  hasCollision,
  renderBlockOnBoard,
  generateBlocks_7bag,
  GravityCurve,
  rotateLeft,
  rotateRight,
  createEmptyBoard,
} from "@/lib/gamelogic";
import { useWebSocket } from "@/components/WebSocketContext";
import { convertNumToTetrominoArray, convertTetrominoToNumArray } from "@/lib/utils";
import { InputBuffer, WSMessage } from "@/types/common";

const TickSpeedType = {
  DropLevel: (level: number) => GravityCurve(level),
  SoftDrop: TickSpeed.Fast,
  LockDelay: TickSpeed.LockDelay,
};

export const useTetrisEventSource = (): {
  board: BoardGrid;
  isPlaying: boolean;
} => {
  const level = 1;
  const [{ board, activeBlock, dRow, dCol, isHoldAvailable, holdBlock, nextBlockIndex }, dispatchBoard] = useBoard();
  const [isPlaying, setIsPlaying] = useState(false);
  const [tickSpeed, setTickSpeed] = useState<TickSpeed | null>(TickSpeedType.DropLevel(level));
  const [isCommitting, setIsCommitting] = useState(false);

  const inputBuffer = useRef<InputBuffer>({
    left: false,
    right: false,
    up: false,
    rotateCounterClockwise: false,
    down: false,
    space: false,
    hold: false,
  });

  const { sendMessage, subscribe } = useWebSocket();
  const lastFrameTime = useRef<number>(0);
  const accumulatedTime = useRef<number>(0);
  const animationFrameId = useRef<number | null>(null);
  const listBlock = useRef<TetrominoType[]>([]);

  const handleRemoteKeyDown = (key: string) => {
    if (key in inputBuffer.current) {
      inputBuffer.current[key as keyof InputBuffer] = true;
      console.log(inputBuffer.current[key as keyof InputBuffer])
      if (key === "down") {
        setTickSpeed(TickSpeedType.SoftDrop);
      }
    }
    
  };

  const handleRemoteKeyUp = (key: string) => {
    if (key in inputBuffer.current) {
      inputBuffer.current[key as keyof InputBuffer] = false;
      if (key === "down") {
        setTickSpeed(TickSpeedType.DropLevel(level));
      }
    }
  };

  const handleInputEvent = useCallback(() => {
    const buffer = inputBuffer.current;

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
      const rotatedBlock: Block = {
        type: activeBlock!.type,
        shape: rotatedShape,
        rState: ((activeBlock?.rState! + 1) % 4) as RotationState,
      };
      dispatchBoard({ type: "keyevent", payload: { key: "rotate", rotatedBlock } });
      buffer.up = false;
    }

    if (buffer.rotateCounterClockwise) {
      const rotatedShape = rotateLeft(activeBlock?.shape!);
      const rotatedBlock: Block = {
        type: activeBlock!.type,
        shape: rotatedShape,
        rState: ((activeBlock?.rState! - 1) % 4) as RotationState,
      };
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
      renderBlockOnBoard(renderedBoard, dRow, dCol, activeBlock);
    }
    return renderedBoard.slice(2, 2 + VISIBLE_HEIGHT);
  }, [board, isPlaying, activeBlock, dRow, dCol]);

  /*
    HOOKS FLOW:  

  */

  const init = useCallback(() => {
    dispatchBoard({
      type: "start",
      payload: { listBlock: listBlock.current },
    });
  }, [sendMessage]);

  // const startGame = useCallback(() => {
  //   setIsPlaying(true);
  //   dispatchBoard({
  //     type: "start",
  //     payload: { listBlock: listBlock.current },
  //   });
  // }, [dispatchBoard]);

  // commit state
  const commitPosition = useCallback(() => {
    if (!activeBlock || !hasCollision(board, activeBlock, dCol, dRow + 1)) {
      setIsCommitting(false);
      setTickSpeed(TickSpeedType.DropLevel(0));
      return;
    }
    let cloneBoard = structuredClone(board) as BoardGrid;
    renderBlockOnBoard(cloneBoard, dRow, dCol, activeBlock);

    setTickSpeed(TickSpeedType.DropLevel(0));

    dispatchBoard({
      type: "commit",
      payload: { listBlock: listBlock.current, committedBoard: cloneBoard },
    });

    checkGameover(cloneBoard);
    setIsCommitting(false);
  }, [board, activeBlock, dRow, dCol, nextBlockIndex, holdBlock, dispatchBoard]);
  //this func catch on message
  const checkGameover = (board: BoardGrid) => {
    /* Giữ nguyên như bản gốc */
  };
  //handle tick speed
  const handleTickSpeed = useCallback(() => {
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

  const gameLoop = useCallback(
    (timestamp: number) => {
      if (!isPlaying) {
        animationFrameId.current && cancelAnimationFrame(animationFrameId.current);
        animationFrameId.current = null;
        return;
      }

      const deltaTime = timestamp - lastFrameTime.current;
      lastFrameTime.current = timestamp;
      accumulatedTime.current += deltaTime;

      const currentTickSpeed = tickSpeed ?? TickSpeedType.DropLevel(level);
      if (accumulatedTime.current >= currentTickSpeed) {
        handleTickSpeed();
        accumulatedTime.current -= currentTickSpeed;
      }

      handleInputEvent(); //1 frame ~16ms, thời gian để bấm giữa 2 phím ~40-100ms (1 giây bấm được 10-15 phím) -> mỗi frame chỉ cần kiểm tra input 1 lần

      animationFrameId.current = requestAnimationFrame(gameLoop);
    },
    [isPlaying, tickSpeed, handleTickSpeed, handleInputEvent]
  );

  //trigger loop khi isPlaying = true
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
  //catch ws message <--> main event listener (thay vì key event như use tetris)
  useEffect(() => {
    const unsubscribe = subscribe((msg: WSMessage) => {
      switch (msg.type) {
        case "init":
          if (msg.payload && msg.payload.listBlock) {
            listBlock.current = convertNumToTetrominoArray(msg.payload.listBlock);
            init();
          } else {
            console.log("init data is missing");
          }
          break;
        case "key_down":
          
          handleRemoteKeyDown(msg.payload.key!);
          break;
        case "key_up": // arrow down có hành vì khác so với các phím còn lại. case này dành riêng cho arrow down
          handleRemoteKeyUp(msg.payload.key!);
          break;
        case "start":
          const delay = (msg.payload!.startAt || Date.now()) - msg.payload.timestamp!;

          setTimeout(() => {
            setIsPlaying(true);
          }, delay);
          break;
      }
    });

    return () => unsubscribe();
  }, [subscribe]);

  return {
    board: isPlaying ? renderBoard() : createEmptyBoard().slice(2, 2 + VISIBLE_HEIGHT),
    isPlaying,
  };
};
