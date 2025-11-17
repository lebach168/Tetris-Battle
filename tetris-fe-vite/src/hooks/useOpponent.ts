import type { WsMessage } from '@/types/common.ts';
import { useCallback, useEffect, useRef, useState } from 'react';
import type { Block, BoardGrid, Tetromino } from '@/types/tetris.ts';
import { VISIBLE_HEIGHT } from '@/types/tetris.ts';
import { createEmptyBoard } from '@/utils/gamelogic.ts';
import { convertToBoardCell } from '@/utils/utils.ts';
import type { MessageHandler } from '@/hooks/useWebsocket.ts';

export function useOpponent(
  msgHandlerRegister: (type: string, handler: MessageHandler) => () => void,
) {
  // Only store the visible board in state to minimize updates
  const [visibleBoard, setVisibleBoard] = useState<BoardGrid>(() =>
    createEmptyBoard().slice(2, 2 + VISIBLE_HEIGHT),
  );

  // prevValuesRef lưu lại giá trị số (value) để so sánh, không lưu object cell
  const prevValuesRef = useRef<number[][]>(
    visibleBoard.map((row) => row.map((cell) => cell.value)),
  );

  // So sánh theo giá trị number primitive thay vì object reference
  const hasBoardValueDiff = (a: number[][], b: number[][]): boolean => {
    if (a.length !== b.length) return true;
    for (let r = 0; r < a.length; r++) {
      const rowA = a[r];
      const rowB = b[r];
      if (rowA.length !== rowB.length) return true;
      for (let c = 0; c < rowA.length; c++) {
        if (rowA[c] !== rowB[c]) return true;
      }
    }
    return false;
  };

  const getOpponentBoard = useCallback((msg: WsMessage) => {
    const payload = msg.payload as {
      state?: {
        board?: number[][];
        block?: number[][];
        cRow?: number;
        cCol?: number;
        isHoldAvailable: boolean;
        holdBlock?: Block;
        nextBlockIndex: number;
        listBlock?: Tetromino[];
      };
    };
    if (!payload || !payload.state || !payload.state.board) {
      console.log('No board data in message payload');
      return;
    }

    // Build full board cells
    const full = convertToBoardCell(payload.state.board);

    // Overlay active block if present
    if (
      payload.state.block &&
      typeof payload.state.cRow === 'number' &&
      typeof payload.state.cCol === 'number'
    ) {
      const shape = payload.state.block;
      const r0 = payload.state.cRow;
      const c0 = payload.state.cCol;
      for (let r = 0; r < shape.length; r++) {
        for (let c = 0; c < shape[r].length; c++) {
          if (shape[r][c] !== 0) {
            const rr = r0 + r;
            const cc = c0 + c;
            full[rr][cc] = { value: shape[r][c], type: String(shape[r][c]) };
          }
        }
      }
    }

    // Only compute and store the visible slice
    const nextVisible = full.slice(2, 2 + VISIBLE_HEIGHT) as BoardGrid;
    const nextValues = nextVisible.map((row) => row.map((cell) => cell.value));

    // Skip update if nothing changed
    if (!hasBoardValueDiff(prevValuesRef.current, nextValues)) return;

    // Update refs + state only when truly different
    prevValuesRef.current = nextValues;
    setVisibleBoard(nextVisible.map((row) => row.slice()) as BoardGrid);
  }, []);

  useEffect(() => {
    const unregister = msgHandlerRegister('opponent', getOpponentBoard);
    return unregister;
  }, [getOpponentBoard]);

  return { board: visibleBoard };
}
