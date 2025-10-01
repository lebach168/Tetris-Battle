import { createFileRoute } from '@tanstack/react-router';
import { useTetrisBattle } from '@/hooks/useTetrisBattle.ts';
import Board from '@/components/Gameplay/Board.tsx';

export const Route = createFileRoute('/match/$id')({
  component: GameWindow,
});

function GameWindow() {
  const { board, startGame, isPlaying, isReady } = useTetrisBattle();

  return (
    <>
      {!isPlaying && (
        <button
          className="bg-gray-300 hover:bg-gray-400 text-gray-800 font-bold py-2 px-4 rounded h-min self-center"
          onClick={startGame}
          //disabled={!isReady}
        >
          Start
        </button>
      )}
      <div className="flex w-full h-full border-4">
        {/* Player Side */}
        <div className="flex-1 ">
          <Board board={board}></Board>
        </div>
      </div>
    </>
  );
}
