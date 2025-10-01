import PlayerSide from './PlayerSide';

import { useTetris } from '@/hooks/useTetrisBattle.ts';

import { Button } from '../ui/button';

export default function GameScreen() {
  const { board, startGame, isPlaying, isReady } = useTetris();

  return (
    <div>
      {!isPlaying && (
        <Button
          className="bg-gray-300 hover:bg-gray-400 text-gray-800 font-bold py-2 px-4 rounded h-min self-center"
          onClick={startGame}
          disabled={!isReady}
        >
          Start
        </Button>
      )}
      {/* Player Side */}
      <div className="flex w-full h-full border-4">
        <div className="flex-1 ">
          <PlayerSide board={board} />
        </div>
        {/* <div className="flex-1">
          <OpponentSide board={opponentBoard} />
        </div> */}
      </div>
    </div>
  );
}
