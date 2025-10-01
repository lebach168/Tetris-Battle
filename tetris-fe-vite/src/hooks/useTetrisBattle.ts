import {
  type BoardGrid,
  type RotationState,
  type Tetromino,
  TETROMINO_SHAPES,
  TickSpeed,
  VISIBLE_HEIGHT,
} from '@/types/tetris';
import { applyBlockOnBoard, createEmptyBoard, hasCollision } from '@/utils/gamelogic';
import { useCallback, useEffect, useRef, useState } from 'react';
import { type BoardState, useGameBoard } from '@/hooks/useGameBoard.ts';
import type { InputBuffer } from '@/types/common.ts';

export function useTetrisBattle(): {
  board: BoardGrid;
  startGame: () => void;
  isPlaying: boolean;
  isReady: boolean;
} {
  //game state
  const [level, setLevel] = useState(0);
  const [isReady, setIsReady] = useState(false);
  const [isPlaying, setIsPlaying] = useState(false);
  const [tickSpeed, setTickSpeed] = useState<number | null>(TickSpeed.DropLevel(level));
  const [isCommitting, setIsCommitting] = useState(false);

  //frame state

  const lastFrameTime = useRef<number>(0);

  const animationFrameId = useRef<number | null>(null);
  const accumulatedTime = useRef<number>(0);

  //const {sendMsg, receiveMsg} = useWebsocket();

  //keyboard state
  const inputBuffer = useRef<InputBuffer>({
    left: false,
    right: false,
    rotate: false, //up
    rotateCounterClockwise: false,
    down: false,
    space: false,
    hold: false,
    // botActions: [],
  });

  const { boardState, dispatchAction } = useGameBoard();

  //TODO init game +
  //TODO start game
  const startGame = useCallback(() => {
    dispatchAction({ type: 'start' });
    setIsPlaying(true);
  }, []);
  const checkGameover = (boardState: BoardState) => {
    const nextBlockType: Tetromino = boardState.listBlock![boardState.nextBlockIndex];
    const nextBlock = nextBlockType
      ? { type: nextBlockType, shape: TETROMINO_SHAPES[nextBlockType], rState: 0 as RotationState }
      : undefined;
    if (nextBlock && hasCollision(boardState.board, nextBlock, 4, 0)) {
      dispatchAction({ type: 'end' });
      setIsPlaying(false);
      setTickSpeed(null);
      // Thông báo game over (tùy chọn)
      console.log('Game Over!');
    }
  };
  const gameStateRef = useRef({ boardState, tickSpeed, level, isCommitting });
  useEffect(() => {
    gameStateRef.current = { boardState, tickSpeed, level, isCommitting };
  }, [boardState, tickSpeed, level, isCommitting]);
  const gameloop = useCallback(
    (timestamp: number) => {
      if (!isPlaying && animationFrameId.current) {
        cancelAnimationFrame(animationFrameId.current);
        animationFrameId.current = null;
        return;
      }
      const delta = timestamp - lastFrameTime.current;
      lastFrameTime.current = timestamp;
      accumulatedTime.current += delta;
      const currentTickSpeed =
        gameStateRef.current.tickSpeed ?? TickSpeed.DropLevel(gameStateRef.current.level);
      if (accumulatedTime.current >= currentTickSpeed) {
        handleCommitPhase();
        accumulatedTime.current -= currentTickSpeed;
      }
      handleInputEvent();
      //request animation frame sẽ giữ nguyên callback qua các lần gọi -> sử dụng ref state để lấy giá trị latest của
      // state kể các trong các function được gọi bên trong này.
      animationFrameId.current = requestAnimationFrame(gameloop);
    },
    [isPlaying],
  );
  //TODO trigger loop or unpause game
  useEffect(() => {
    if (isPlaying && !animationFrameId.current) {
      lastFrameTime.current = performance.now();
      accumulatedTime.current = 0;
      animationFrameId.current = requestAnimationFrame(gameloop);
    }
    return () => {
      if (animationFrameId.current) {
        cancelAnimationFrame(animationFrameId.current);
        animationFrameId.current = null;
      }
    };
  }, [isPlaying]);

  //TODO commits block
  function commitPosition() {
    //Check if the block is no longer has collision
    const curBoard = gameStateRef.current.boardState;
    if (
      curBoard.activeBlock &&
      !hasCollision(curBoard.board, curBoard.activeBlock, curBoard.cRow + 1, curBoard.cCol)
    ) {
      setIsCommitting(false);
      return;
    }

    const copyBoard = structuredClone(curBoard.board);
    applyBlockOnBoard(copyBoard, curBoard.cRow, curBoard.cCol, curBoard.activeBlock!);

    dispatchAction({ type: 'commit', payload: { committedBoard: copyBoard } });

    setTickSpeed(TickSpeed.DropLevel(gameStateRef.current.level));
    setIsCommitting(false);
  }

  // tick speed
  const handleCommitPhase = useCallback(() => {
    const curBoard = gameStateRef.current.boardState;
    if (gameStateRef.current.isCommitting) {
      commitPosition();
    } else if (
      curBoard.activeBlock &&
      hasCollision(curBoard.board, curBoard.activeBlock, curBoard.cRow + 1, curBoard.cCol)
    ) {
      setTickSpeed(TickSpeed.LockDelay);
      setIsCommitting(true);
    } else {
      dispatchAction({ type: 'drop' });
      //dispatchAction({ type: "soft_drop", cells: 1 });
    }
  }, []);

  const handleInputEvent = () => {
    if (inputBuffer.current.down) {
      setTickSpeed(TickSpeed.SoftDrop);
      inputBuffer.current.down = false;
    }
    if (inputBuffer.current.left) {
      dispatchAction({ type: 'key_event', payload: { key: 'left' } });
      inputBuffer.current.left = false;
    }
    if (inputBuffer.current.right) {
      dispatchAction({ type: 'key_event', payload: { key: 'right' } });
      inputBuffer.current.right = false;
    }
    if (inputBuffer.current.rotate) {
      dispatchAction({ type: 'key_event', payload: { key: 'rotate_right' } });
      inputBuffer.current.rotate = false;
    }
    if (inputBuffer.current.rotateCounterClockwise) {
      dispatchAction({ type: 'key_event', payload: { key: 'rotate_left' } });
      inputBuffer.current.rotateCounterClockwise = false;
    }
    if (inputBuffer.current.space) {
      dispatchAction({ type: 'key_event', payload: { key: 'space' } });
      inputBuffer.current.space = false;
    }
    if (inputBuffer.current.hold) {
      dispatchAction({ type: 'key_event', payload: { key: 'hold' } });
      inputBuffer.current.hold = false;
    }
  };

  const getVisibleBoard = (boardState: BoardState): BoardGrid => {
    const boardCopy = structuredClone(boardState.board) as BoardGrid; // tạo copy để tạo hiệu ứng không ảnh hưởng tới
    // board thật và react lấy giá trị được thay đổi
    applyBlockOnBoard(boardCopy, boardState.cRow, boardState.cCol, boardState.activeBlock!);
    return boardCopy.slice(2, VISIBLE_HEIGHT + 2);
  };
  //key event listener
  useEffect(() => {
    if (!isPlaying) return;
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.repeat) return;
      if (event.key === ' ') {
        // prevent scrolling down when press space bar
        event.preventDefault();
        inputBuffer.current.space = true;
      }
      if (event.key === 'ArrowDown') inputBuffer.current.down = true;
      if (event.key === 'ArrowLeft') inputBuffer.current.left = true;
      if (event.key === 'ArrowRight') inputBuffer.current.right = true;
      if (event.key === 'ArrowUp') inputBuffer.current.rotate = true;
      if (event.key === 'z') inputBuffer.current.rotateCounterClockwise = true;
      if (event.key === 'c') inputBuffer.current.hold = true;
    };

    //this event for sliding right and left or soft drop
    const handleKeyUp = (event: KeyboardEvent) => {
      if (event.key === 'ArrowDown') {
        setTickSpeed(TickSpeed.DropLevel(level));
        inputBuffer.current.down = false;
      }
      if (event.key === 'ArrowLeft') inputBuffer.current.left = false;
      if (event.key === 'ArrowRight') inputBuffer.current.right = false;
    };
    document.addEventListener('keydown', handleKeyDown);
    document.addEventListener('keyup', handleKeyUp);
    //cleanup
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.removeEventListener('keyup', handleKeyUp);
      inputBuffer.current = {
        left: false,
        right: false,
        rotate: false,
        rotateCounterClockwise: false,
        down: false,
        space: false,
        hold: false,
      };
    };
  }, [isPlaying, level]);

  return {
    board: isPlaying
      ? getVisibleBoard(boardState)
      : createEmptyBoard().slice(2, 2 + VISIBLE_HEIGHT),
    startGame,
    isPlaying,
    isReady,
  };
}
