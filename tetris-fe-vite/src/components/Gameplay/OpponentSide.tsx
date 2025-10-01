'use client';
import Board from './Board';
import { BoardGrid } from '@/types/tetris';

export default function OpponentSide({ board }: { board: BoardGrid }) {
  const username = 'Opponent';
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-900 text-white font-mono">
      <div className="flex flex-col items-center gap-3">
        {/* Gameplay block: Board + StatPanel */}
        <div className="flex gap-3 items-start">
          {/* Board ở giữa */}
          <div className="flex flex-col gap 2">
            <Board board={board} />
            <div className="mb-2">
              <p className="text-lg font-bold text-white">{username}</p>
            </div>
          </div>

          {/* Panel phải: StatPanel (score + next) */}
          {/* <div className="flex flex-col gap-4">
              <StatPanel
                score={1234}
                lines={10}
                combo={2}
                nextBlock={{
                  type: "Z" as TetrominoType,
                  shape: TETROMINO_SHAPES["Z"],
                  rState: 0,
                }}
              />
            </div> */}
        </div>
      </div>
    </div>
  );
}
