"use client";
import { useTetris } from "@/hooks/useTetris";
import Board from "./Board";
import Hold from "./Hold";
import StatPanel from "./StatPanel";
import { TETROMINO_SHAPES, TetrominoType } from "@/types/tetris";

export default function GameplayWindow() {
  const { board, startGame, isPlaying } = useTetris();
    const username = "You"
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-900 text-white font-mono">
    
      <div className="flex flex-col items-center gap-3">
        {/* Button start - chỉ hiện khi chưa bắt đầu */}
        {!isPlaying && (
          <button
            className="bg-gray-300 hover:bg-gray-400 text-gray-800 font-bold py-2 px-4 rounded"
            onClick={startGame}
          >
            Start
          </button>
        )}

        {/* Gameplay block: Hold + Board + StatPanel */}
        {isPlaying && (
          <div className="flex gap-3 items-start">
            {/* Panel trái: Hold + UserInfo */}
            {/* <div className="flex flex-col items-end gap-4">
              
              <Hold />
            </div> */}

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
        )}
      </div>
    </div>
  );
}
