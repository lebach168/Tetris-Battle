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
  addBlockToBoard,
  generateBlocks_7bag,
  GravityCurve,
  rotateLeft,
  rotateRight,
  createEmptyBoard,
} from "@/lib/gamelogic";
import { useWebSocket } from "@/components/WebSocketContext";
import { convertTetrominoToNumArray } from "@/lib/utils";
import { InputBuffer, WSMessage } from "@/types/common";

const TickSpeedType = {
  DropLevel: (level: number) => GravityCurve(level),
  SoftDrop: TickSpeed.Fast,
  LockDelay: TickSpeed.LockDelay,
};

export const useTetrisEventSource = (): {
  board: BoardGrid;
  startGame: () => void;
  isPlaying: boolean;
  init: () => void;
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

  // Khởi tạo WebSocket message handlers
  useEffect(() => {
    const unsubscribe = subscribe((msg: WSMessage) => {
      switch (msg.type) {
        case "KEY_DOWN": //
          handleRemoteKeyDown(msg.payload.key!);
          break;
        case "KEY_UP": // arrow down có hành vì khác so với các phím còn lại. case này dành riêng cho arrow down
          handleRemoteKeyUp(msg.payload.key!);
          break;
        case "start":
          handleGameStart(msg);
          break;
      }
    });

    return () => unsubscribe();
  }, [subscribe]);

  const handleRemoteKeyDown = (key: string) => {
    const keyMap: Record<string, keyof InputBuffer> = {
      ArrowLeft: "left",
      ArrowRight: "right",
      ArrowUp: "up",
      ArrowDown: "down",
      " ": "space",
      z: "rotateCounterClockwise",
      c: "hold",
    };

    const bufferKey = keyMap[key];
    if (bufferKey) {
      inputBuffer.current[bufferKey] = true;
      
      // Xử lý soft drop
      if (key === "ArrowDown") {
        setTickSpeed(TickSpeedType.SoftDrop);
      }
    }
  };

  const handleRemoteKeyUp = (key: string) => {
    const keyMap: Record<string, keyof InputBuffer> = {
      ArrowLeft: "left",
      ArrowRight: "right",
      ArrowDown: "down",
    };

    const bufferKey = keyMap[key];
    if (bufferKey) {
      inputBuffer.current[bufferKey] = false;
      
      // Reset tốc độ khi nhả phím xuống
      if (key === "ArrowDown") {
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

  // Phần còn lại giữ nguyên như useTetris gốc
  const renderBoard = useCallback(() => {
    const renderedBoard = structuredClone(board) as BoardGrid;
    if (isPlaying && activeBlock) {
      addBlockToBoard(renderedBoard, dRow, dCol, activeBlock);
    }
    return renderedBoard.slice(2, 2 + VISIBLE_HEIGHT);
  }, [board, isPlaying, activeBlock, dRow, dCol]);

  const init = useCallback(() => {
    const generated = generateBlocks_7bag(1000);
    listBlock.current = generated;
    sendMessage({
      type: "init",
      payload: {
        listBlock: convertTetrominoToNumArray(generated),
      },
    });
  }, [sendMessage]);

  const startGame = useCallback(() => {
    setIsPlaying(true);
    const generated = generateBlocks_7bag(1000);
    listBlock.current = generated;
    dispatchBoard({
      type: "start",
      payload: { listBlock: generated },
    });
  }, [dispatchBoard]);

  const commitPosition = useCallback(() => {
    /* Giữ nguyên như bản gốc */
  }, [/* Các dependencies gốc */]);

  const checkGameover = (board: BoardGrid) => {
    /* Giữ nguyên như bản gốc */
  };

  const gameTick = useCallback(() => {
    /* Giữ nguyên như bản gốc */
  }, [/* Các dependencies gốc */]);

  const gameLoop = useCallback(
    (timestamp: number) => {
      /* Giữ nguyên như bản gốc */
    },
    [isPlaying, tickSpeed, gameTick, handleInputEvent]
  );

  useEffect(() => {
    /* Giữ nguyên logic game loop */
  }, [isPlaying, gameLoop]);

  const handleGameStart = (msg: WSMessage) => {
    const delay = msg.payload!.startAt || 0;
    setTimeout(() => {
      startGame();
    }, delay);
  };

  return {
    board: isPlaying ? renderBoard() : createEmptyBoard().slice(2, 2 + VISIBLE_HEIGHT),
    startGame,
    isPlaying,
    init,
  };
};