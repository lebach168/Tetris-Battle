import { createFileRoute } from '@tanstack/react-router';
import { useTetrisBattle } from '@/hooks/useTetrisBattle.ts';
import Board from '@/components/Gameplay/Board.tsx';
import { createEmptyBoard } from '@/utils/gamelogic.ts';

export const Route = createFileRoute('/match/$id')({
  component: GameWindow,
});

function GameWindow() {

  const { board, startGame, isPlaying } = useTetrisBattle();

  return (
    <>
      <div className="flex w-full h-full items-center justify-between border-4 relative">
        {/* Player 1 (sát lề trái) */}
        <div className="flex justify-center">
          <Board board={board} />
        </div>

        {/* Start Button ở giữa */}
        <div className="absolute left-1/2 -translate-x-1/2">
          <button
            className={`bg-gray-300 hover:bg-gray-400 text-gray-800 font-bold py-2 px-4 rounded 
        transition-opacity duration-700 
        ${isPlaying ? 'opacity-0 pointer-events-none' : 'opacity-100'}`}
            onClick={startGame}
          >
            Start
          </button>
        </div>

        {/* Player 2 (sát lề phải) */}
        <div className="flex justify-center">
          <Board board={createEmptyBoard()} />
        </div>
      </div>
    </>
  );
}
