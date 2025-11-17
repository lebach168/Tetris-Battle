import {
  type BoardGrid,
  type Cell,
  DropSpeed,
  keys,
  LockDelay,
  VISIBLE_HEIGHT,
} from '@/types/tetris';
import {
  applyBlockOnBoard,
  createEmptyBoard,
  findLandingPosition,
  getTetrominoTypeFromShape,
  hasCollision,
} from '@/utils/gamelogic';
import { useCallback, useEffect, useRef, useState } from 'react';
import { type BoardState, useMovementPrediction } from '@/hooks/useMovementPrediction.ts';
import type { InputBuffer, WsMessage } from '@/types/common.ts';
import { convertNumToTetrominoArray, mapKeyToString } from '@/utils/utils.ts';
import type { MessageHandler } from '@/hooks/useWebsocket.ts';

export type FrameHistory = {
  frame: number;
  state: BoardState;
  inputs: string[];
  gravityTime: number;
  lockTime: number;
  onGround: boolean;
  dropSpeed: number;
  timestamp: number;
};
export type ServerState = {
  frame: number;
  boardState: BoardState;
  gravityTime: number;
  lockTime: number;
  onGround: boolean;
  confirmed: boolean;
};
export function useTetrisBattle(sendMsg: (msg: WsMessage) => void) {
  // Game state

  const [level, _setLevel] = useState(0);

  const [isReady, _setIsReady] = useState(false);
  const [isPlaying, setIsPlaying] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [visibleBoard, setVisibleBoard] = useState<BoardGrid>(
    createEmptyBoard().slice(2, 2 + VISIBLE_HEIGHT),
  );

  // Frame state -
  const lastFrameTime = useRef(0);
  const animationFrameId = useRef<number | null>(null);
  const gravityTimer = useRef(0);
  const lockTimer = useRef(0);
  const netAccumulatedTime = useRef(0);
  const frameCounter = useRef(0);
  const TICK_INTERVAL_MS = 1000 / 30;
  const NET_SEND_EVERY = 2;
  const pendingInputs = useRef<Array<{ keys: string[]; frame: number }>>([]);
  const serverStateRef = useRef<ServerState | null>(null);
  // Input buffers
  const inputBuffer = useRef<InputBuffer>({
    left: false,
    right: false,
    rotate: false,
    rrotate: false,
    down: false,
    downOff: false,
    space: false,
    hold: false,
  });
  // Separate input buffer used for replay/resimulation to avoid mutating
  // the live inputBuffer and causing race conditions with key handlers.
  const replayInputBuffer = useRef<InputBuffer>({
    left: false,
    right: false,
    rotate: false,
    rrotate: false,
    down: false,
    downOff: false,
    space: false,
    hold: false,
  });
  //store buffer

  // Use ref for all game state
  const { boardStateRef, applyAction } = useMovementPrediction();

  const history = useRef<FrameHistory[]>([]);
  const HISTORY_MAX_SIZE = 100;
  const inputDesync = useRef(false);
  // Reconcile / syncing refs

  // const replayQueueRef = useRef<FrameHistory[] | null>(null);
  const replayFromFrameRef = useRef(0);
  // index into history array where resimulation should start
  const replayIndexRef = useRef<number | null>(null);

  // Game state ref
  const gameStateRef = useRef({
    onGround: false,
    dropSpeed: DropSpeed.DropLevel(0),
    level: 0,
    isPlaying: false,
    isPaused: false,
  });

  // Helper to restore fields into the existing boardStateRef.current instead of
  // replacing the ref itself. Keeps other refs that hold the object stable.
  function restoreBoardStateFromSnapshot(snapshot: BoardState) {
    const cur = boardStateRef.current;
    // Deep-clone mutable nested structures to avoid accidental shared references
    cur.board = structuredClone(snapshot.board);
    cur.activeBlock = snapshot.activeBlock ? structuredClone(snapshot.activeBlock) : undefined;
    cur.cCol = snapshot.cCol;
    cur.cRow = snapshot.cRow;
    cur.blockIndex = snapshot.blockIndex;
    cur.canHold = snapshot.canHold;
    cur.holdBlock = snapshot.holdBlock;
    // debug: log restored active block and coordinates
    console.debug('[restoreBoardState] restored activeBlock:', {
      activeBlock: cur.activeBlock,
      cRow: cur.cRow,
      cCol: cur.cCol,
      blockIndex: cur.blockIndex,
      listBlockLength: cur.listBlock?.length,
    });
  }

  const updateVisibleBoard = useCallback(() => {
    const boardCopy = structuredClone(boardStateRef.current.board) as BoardGrid;
    if (boardStateRef.current.activeBlock) {
      applyBlockOnBoard(
        boardCopy,
        boardStateRef.current.cRow,
        boardStateRef.current.cCol,
        boardStateRef.current.activeBlock,
      );
    }
    setVisibleBoard(boardCopy.slice(2, VISIBLE_HEIGHT + 2));
  }, [boardStateRef]);

  const startGame = useCallback(() => {
    const readyMessage = {
      type: 'ready',
      timestamp: Date.now(),
    } as WsMessage;
    sendMsg(readyMessage);
  }, [sendMsg]);

  const pauseGame = useCallback(() => {
    setIsPaused(true);
    gameStateRef.current.isPaused = true;
    const pauseMessage = {
      type: 'pause',
    };
    sendMsg(pauseMessage);
  }, [sendMsg]);

  const unpauseGame = useCallback(() => {
    setIsPaused(false);
    gameStateRef.current.isPaused = false;
    const unpauseMessage = {
      type: 'unpause',
      payload: {},
    };
    sendMsg(unpauseMessage);
  }, [sendMsg]);

  const checkGameover = useCallback(() => {
    const boardState = boardStateRef.current;

    if (boardState.activeBlock && hasCollision(boardState.board, boardState.activeBlock, 0, 4)) {
      applyAction({ type: 'end' });
      setIsPlaying(false);
      gameStateRef.current.isPlaying = false;
      console.log('Game Over!');
    }
  }, [boardStateRef]); /*applyAction, */

  const handleCommitPhase = useCallback(() => {
    //check commit condition if no longer on ground  reset locktimer and continue
    //else check clear line logic -> spawn new piece (block) -> check game over
    const state = boardStateRef.current;
    const landingRow = findLandingPosition(state.board, state.activeBlock!, state.cRow, state.cCol);
    if (!gameStateRef.current.onGround) {
      return;
    }
    if (state.cRow < landingRow) {
      gameStateRef.current.onGround = false;
      lockTimer.current = 0;
      return;
    }
    if (lockTimer.current >= LockDelay) {
      const copyBoard = structuredClone(state.board);
      applyBlockOnBoard(copyBoard, landingRow, state.cCol, state.activeBlock!);
      //update board value and spawn new block inside apply action type commit
      applyAction({ type: 'commit', payload: { committedBoard: copyBoard } });
      //now check gameover with current state after commit
      checkGameover();
      gameStateRef.current.onGround = false;
      lockTimer.current = 0;
      gravityTimer.current = 0;
    }
  }, [applyAction, boardStateRef, checkGameover]);

  // handleInputEvent accepts an optional input source object so that
  // replay/resimulation uses a separate buffer and doesn't mutate the live
  // inputBuffer (which receives real-time key events from the user).
  const handleInputEvent = useCallback(
    (src?: InputBuffer) => {
      // processing order : horizontal move -> rotate -> soft drop -> hold -> hard drop
      const ib = src ?? inputBuffer.current;
      if (ib.left) {
        applyAction({ type: 'key_event', payload: { key: 'left' } });
        ib.left = false;
      }
      if (ib.right) {
        applyAction({ type: 'key_event', payload: { key: 'right' } });
        ib.right = false;
      }
      if (ib.rotate) {
        applyAction({ type: 'key_event', payload: { key: 'rotate_right' } });
        ib.rotate = false;
      }
      if (ib.rrotate) {
        applyAction({ type: 'key_event', payload: { key: 'rotate_left' } });
        ib.rrotate = false;
      }
      if (ib.down) {
        gameStateRef.current.dropSpeed = DropSpeed.SoftDrop;
        ib.down = false;
      }
      if (ib.downOff) {
        //arrow down key up
        gameStateRef.current.dropSpeed = DropSpeed.DropLevel(level);
        ib.downOff = false;
      }
      if (ib.hold) {
        applyAction({ type: 'key_event', payload: { key: 'hold' } });
        ib.hold = false;
      }
      if (ib.space) {
        applyAction({ type: 'key_event', payload: { key: 'space' } });
        gameStateRef.current.onGround = true;
        lockTimer.current = LockDelay;
        ib.space = false;
      }
    },
    [applyAction, level],
  );

  const gameloop = useCallback(
    (timestamp: number) => {
      if ((!isPlaying || isPaused) && animationFrameId.current) {
        cancelAnimationFrame(animationFrameId.current);
        animationFrameId.current = null;
        return;
      }
      let syncing = false; // flag
      const delta = timestamp - lastFrameTime.current;
      lastFrameTime.current = timestamp;
      netAccumulatedTime.current += delta;
      const replayFrame = replayFromFrameRef.current;
      if (inputDesync.current) {
        syncing = true;

        const idx = history.current.findIndex((h) => h.frame === replayFrame);

        if (idx >= 0) {
          // anchor = end-of-previous-frame if available, otherwise assume one tick before this entry
          const anchorTimestamp =
            idx > 0
              ? history.current[idx - 1].timestamp
              : history.current[idx].timestamp - TICK_INTERVAL_MS;

          // elapsed real ms since the end of previous frame — this includes all loop deltas
          netAccumulatedTime.current = Math.max(0, timestamp - anchorTimestamp);
          // restore board state and timers to the anchor frame so we can replay forward
          // Note: pass a fresh structured clone of the history state into the helper
          restoreBoardStateFromSnapshot(structuredClone(history.current[idx].state) as BoardState);
          gravityTimer.current = history.current[idx].gravityTime;

          lockTimer.current = history.current[idx].lockTime;
          gameStateRef.current.onGround = history.current[idx].onGround;
          gameStateRef.current.dropSpeed = history.current[idx].dropSpeed;
          // set replay pointer to this index in history array
          replayIndexRef.current = idx;
          // ensure replayFromFrameRef remains consistent with history entry
          replayFromFrameRef.current = history.current[idx].frame;
          console.log(
            'current frame: ' +
              frameCounter.current +
              ' mismatch frame: ' +
              replayFrame +
              ' accumulator: ' +
              netAccumulatedTime.current,
          );
        } else {
          // no matching history: avoid huge jumps — advance a small amount so loop can progress
          //case : missing history records
        }
      }
      while (netAccumulatedTime.current >= TICK_INTERVAL_MS) {
        netAccumulatedTime.current -= TICK_INTERVAL_MS;

        // choose current frame: if syncing, use the replay index frame (don't advance the global frame counter)
        let curFrame: number;
        let usingReplay = false;
        let replayIdxLocal: number | null = null;
        if (syncing && replayIndexRef.current != null) {
          usingReplay = true;
          replayIdxLocal = replayIndexRef.current;
          curFrame = history.current[replayIdxLocal].frame;
        } else {
          curFrame = ++frameCounter.current;
        }

        // Determine inputs for this tick. If we're syncing, consume inputs from
        // the replay queue (history) instead of live inputBuffer and don't send
        // network input messages.

        const currentInputs: Array<{ keys: string[]; frame: number }> = [];
        let keyArray: string[] = [];
        if (usingReplay && replayIdxLocal != null) {
          // take inputs from history entry (server-confirmed or previously patched)
          keyArray = history.current[replayIdxLocal].inputs.slice();
          // do not touch pendingInputs or send network messages while resimulating
        } else {
          // live input extraction (normal play)
          for (const k of keys) {
            if (inputBuffer.current[k]) {
              keyArray.push(mapKeyToString(k));
            }
          }
          currentInputs.push({ keys: keyArray, frame: curFrame });
          pendingInputs.current.push(...currentInputs);
        }

        // Send inputs periodically
        if (!syncing && frameCounter.current % NET_SEND_EVERY === 0) {
          sendMsg({
            type: 'inputs',
            payload: {
              inputs: pendingInputs.current.slice(),
              latestFrame: curFrame,
            },
          });
          pendingInputs.current = [];
        }

        // Apply live input buffer behavior
        if (usingReplay && replayIdxLocal != null) {
          // Use a separate replay input buffer so live inputBuffer is not mutated
          // (avoids race conditions with real-time key handlers).
          const rib = replayInputBuffer.current;
          Object.keys(rib).forEach((k) => {
            rib[k as keyof InputBuffer] = false;
          });
          for (const k of keyArray) {
            switch (k) {
              case 'left':
                rib.left = true;
                break;
              case 'right':
                rib.right = true;
                break;
              case 'rotate_right':
                rib.rotate = true;
                break;
              case 'rotate_left':
                rib.rrotate = true;
                break;
              case 'down':
                rib.down = true;
                break;
              case 'space':
                rib.space = true;
                break;
              case 'hold':
                rib.hold = true;
                break;
              case 'downOff':
                rib.downOff = true;
                break;
            }
          }
          // apply using the replay buffer (doesn't affect live inputBuffer)
          handleInputEvent(rib);
        } else {
          handleInputEvent(inputBuffer.current);
        }

        // Gravity drop
        const currentDropSpeed = gameStateRef.current.dropSpeed;
        const bs = boardStateRef.current;

        const landingRow = findLandingPosition(bs.board, bs.activeBlock!, bs.cRow, bs.cCol);
        if (!gameStateRef.current.onGround) {
          gravityTimer.current += TICK_INTERVAL_MS;
          if (gravityTimer.current >= currentDropSpeed) {
            if (bs.cRow < landingRow) {
              applyAction({ type: 'drop' });
              gravityTimer.current -= currentDropSpeed;
            }
            if (bs.cRow >= landingRow) {
              gameStateRef.current.onGround = true;
            }
          }
        } else {
          lockTimer.current += TICK_INTERVAL_MS;
        }

        handleCommitPhase();

        // If we are syncing, after commit/snapshots update the corresponding
        // history entry from the replay queue and advance the replay pointer.
        // Record frame state Update history
        if (history.current.length > HISTORY_MAX_SIZE) {
          // if we are currently replaying, shifting the history's head will change
          // indices; compensate so replayIndexRef still points to the correct entry
          if (syncing && replayIndexRef.current != null) {
            replayIndexRef.current = Math.max(0, replayIndexRef.current - 1);
          }
          history.current.shift();
        }
        const stateSnapshot = structuredClone(boardStateRef.current);
        if (usingReplay && replayIdxLocal != null) {
          // update the historical entry (preserve original inputs/timestamp/frame)
          const orig = history.current[replayIdxLocal];
          history.current[replayIdxLocal] = {
            frame: orig.frame,
            state: stateSnapshot,
            inputs: orig.inputs.slice(), // keep server-confirmed inputs
            gravityTime: gravityTimer.current,
            lockTime: lockTimer.current,
            onGround: gameStateRef.current.onGround,
            dropSpeed: currentDropSpeed,
            timestamp: orig.timestamp,
          };
          // advance replay index to next history frame
          replayIndexRef.current = replayIdxLocal + 1;
          // if we've consumed all known history entries, finish syncing
          if (replayIndexRef.current >= history.current.length) {
            // set global frame counter to last known frame + 1 so regular loop resumes
            const last = history.current[history.current.length - 1];
            frameCounter.current = last ? last.frame + 1 : frameCounter.current;
            syncing = false;
            inputDesync.current = false;
            replayFromFrameRef.current = 0;
            replayIndexRef.current = null;
          }
        } else {
          history.current.push({
            frame: curFrame,
            state: stateSnapshot,
            inputs: keyArray,
            gravityTime: gravityTimer.current,
            lockTime: lockTimer.current,
            onGround: gameStateRef.current.onGround,
            dropSpeed: currentDropSpeed,
            timestamp: timestamp - netAccumulatedTime.current, //real start time of frame with no surplus
          });
        }
        //// Update visible board for rendering every frame
        updateVisibleBoard();
      }

      animationFrameId.current = requestAnimationFrame(gameloop);
    },
    [
      isPlaying,
      isPaused,
      sendMsg,
      applyAction,
      TICK_INTERVAL_MS,
      boardStateRef,
      handleCommitPhase,
      handleInputEvent,
      updateVisibleBoard,
    ],
  );

  useEffect(() => {
    if (isPlaying && !isPaused && !animationFrameId.current) {
      lastFrameTime.current = performance.now();
      gravityTimer.current = 0;
      gameStateRef.current.isPlaying = true;
      gameStateRef.current.isPaused = false;
      animationFrameId.current = requestAnimationFrame(gameloop);
    }
    return () => {
      if (animationFrameId.current) {
        cancelAnimationFrame(animationFrameId.current);
        animationFrameId.current = null;
      }
    };
  }, [isPlaying, isPaused]);

  // Keyboard event listeners
  const inputCounter = useRef(0);
  useEffect(() => {
    if (!isPlaying || isPaused) return;

    const ib = inputBuffer.current; // capture stable ref for cleanup

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.repeat) return;
      if (event.key === ' ') {
        event.preventDefault();
        ib.space = true;
      }
      if (event.key === 'ArrowDown') {
        ib.down = true;

        inputCounter.current++;
      }
      if (event.key === 'ArrowLeft') {
        ib.left = true;

        inputCounter.current++;
      }
      if (event.key === 'ArrowRight') {
        ib.right = true;

        inputCounter.current++;
      }
      if (event.key === 'ArrowUp') {
        ib.rotate = true;

        inputCounter.current++;
      }
      if (event.key === 'z') {
        ib.rrotate = true;

        inputCounter.current++;
      }
      if (event.key === 'c') {
        ib.hold = true;

        inputCounter.current++;
      }
    };

    const handleKeyUp = (event: KeyboardEvent) => {
      if (event.key === 'ArrowDown') {
        ib.downOff = true;

        inputCounter.current++;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    document.addEventListener('keyup', handleKeyUp);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.removeEventListener('keyup', handleKeyUp);
      // Reset input buffers
      Object.keys(ib).forEach((key) => {
        ib[key as keyof InputBuffer] = false;
      });
    };
  }, [isPlaying, isPaused, level]);

  const getServerState = useCallback(
    (msg: WsMessage) => {
      const { state, latestFrame } = msg.payload as {
        state: {
          board: number[][];
          block: number[][];
          bForm: number;
          cRow: number;
          cCol: number;
          canHold: boolean;
          holdBlock: number;
          blockIndex: number;
          accumulator: number;
          lockTime: number;
          dropSpeed: number;
          onGround: boolean;
        };
        latestFrame: number;
      };

      if (serverStateRef.current == null) {
        serverStateRef.current = {
          frame: 0,
          // initialize with a shallow copy; we'll replace fields below
          boardState: structuredClone(boardStateRef.current),
          gravityTime: 0,
          lockTime: 0,
          onGround: false,
          confirmed: true,
        };
      }
      serverStateRef.current.confirmed = false;
      serverStateRef.current.frame = latestFrame;
      serverStateRef.current.gravityTime = state.accumulator;
      serverStateRef.current.onGround = state.onGround;
      serverStateRef.current.lockTime = state.lockTime;
      //serverStateRef.current.accumulatedTime =
      // build a fresh BoardState object from payload to avoid mutating live state
      serverStateRef.current.boardState = {
        board: state.board.map((row) =>
          row.map((cell) => ({ value: cell, type: cell.toString() }) as Cell),
        ),
        activeBlock: {
          shape: state.block,
          type: getTetrominoTypeFromShape(state.block),
          form: state.bForm,
        },
        cCol: state.cCol,
        cRow: state.cRow,
        blockIndex: state.blockIndex,
        canHold: state.canHold,
        holdBlock: state.holdBlock,
        dropSpeed: state.dropSpeed,
        listBlock: boardStateRef.current.listBlock,
      } as BoardState;
    },
    [boardStateRef],
  );
  const checkInputsSync = useCallback((msg: WsMessage) => {
    // assume server sends msg.payload.inputs: Array<{ frame:number; keys:string[] }>
    const serverRecords = (msg.payload?.inputs ?? []) as Array<{ frame: number; keys: string[] }>;
    if (!Array.isArray(serverRecords)) return;

    for (const entry of serverRecords) {
      const { frame, keys: serverKeys } = entry;
      if (!Array.isArray(serverKeys)) {
        continue;
      }

      const idx = history.current.findIndex((h) => h.frame === frame);
      if (idx === -1) {
        continue;
      }

      const hist = history.current[idx];
      const equal =
        hist.inputs.length === serverKeys.length &&
        hist.inputs.every((k) => serverKeys.includes(k));

      if (!equal) {
        // overwrite history inputs with server-confirmed keys
        hist.inputs = serverKeys.slice();
        // only record the first mismatch as the resimulation anchor. Later mismatches
        // will be handled by the same replay started from the earliest bad frame.
        if (!inputDesync.current) {
          inputDesync.current = true;
          replayFromFrameRef.current = frame;
        }
        console.log(frame, ': mismatch');
      }

      // overwrite history inputs with server-confirmed keys
    }
  }, []);
  const checkGarbageSync = useCallback((msg: WsMessage) => {
    const payload = msg.payload as {
      state?: {
        board?: number[][];
        block?: number[][];
        cRow?: number;
        cCol?: number;
      };
      latestFrame: number;
    };

    if (!payload || !payload.state || !payload.state.board) {
      console.log('No garbage state data in message payload');
      return;
    }

    const targetFrame = payload.latestFrame;
    const idx = history.current.findIndex((h) => h.frame === targetFrame);

    if (idx === -1) {
      console.warn(`Garbage sync: frame ${targetFrame} not found in history`);
      return;
    }

    // Convert server board to BoardGrid
    const board = payload.state.board.map((row) => row.map((cell) => ({ value: cell }) as Cell));

    // Build new BoardState from payload
    const newState: BoardState = {
      board: board,
      cCol: payload.state.cCol ?? history.current[idx].state.cCol,
      cRow: payload.state.cRow ?? history.current[idx].state.cRow,
      canHold: history.current[idx].state.canHold,
      holdBlock: history.current[idx].state.holdBlock,
      blockIndex: history.current[idx].state.blockIndex,
      listBlock: history.current[idx].state.listBlock,
    };

    // Update history entry at latestFrame with new state
    const histEntry = history.current[idx];
    history.current[idx] = {
      ...histEntry,
      state: newState,
    };

    // Trigger replay from latestFrame + 1
    // The gameloop will detect inputDesync and start replaying from that frame
    inputDesync.current = true; //xem garbage như input của server (khỏi đổi tên state)
    replayFromFrameRef.current = targetFrame + 1;
    replayIndexRef.current = null; // Will be set by gameloop when it finds the frame

    console.log(
      `Garbage sync: updated frame ${targetFrame}, replaying from frame ${targetFrame + 1}`,
    );
  }, []);
  const messageHandler: MessageHandler = useCallback(
    (msg: WsMessage) => {
      if (msg.type === 'start') {
        if (msg.error !== undefined) {
          console.log(msg.error);
          return;
        }
        if (!gameStateRef.current.isPlaying) {
          const startAt = msg.payload!.startAt;
          const listBlockNums = msg.payload!.listBlock ?? [];
          const listBlock = convertNumToTetrominoArray(listBlockNums);
          const currentTime = Date.now();
          const delay = startAt ? Math.max(0, startAt - currentTime) : 0;

          setTimeout(() => {
            applyAction({ type: 'start', payload: { listBlock } });
            setIsPlaying(true);
            setIsPaused(false);
            gameStateRef.current.isPlaying = true;
            gameStateRef.current.isPaused = false;
            updateVisibleBoard();
          }, delay);

          const startMessage = {
            type: 'start',
            payload: { lastestFrame: 0 },
            timestamp: Date.now(),
          } as WsMessage;
          sendMsg(startMessage);
        }
      } else if (msg.type === 'unpause') {
        setIsPaused(false);
        gameStateRef.current.isPaused = false;
      } else if (msg.type === 'pause') {
        setIsPaused(true);
        gameStateRef.current.isPaused = true;
      } else if (msg.type === 'input-server') {
        checkInputsSync(msg);
      } else if (msg.type === 'server-state') {
        // populate serverStateRef for later reconcile
        getServerState(msg);
      } else if (msg.type === 'garbage-sync') {
        checkGarbageSync(msg);
      }
    },
    [applyAction, sendMsg, updateVisibleBoard, checkInputsSync, getServerState, checkGarbageSync],
  );

  return {
    board: visibleBoard,
    startGame,
    pauseGame,
    unpauseGame,
    isPlaying,
    isPaused,
    isReady,
    messageHandler,
  };
}
