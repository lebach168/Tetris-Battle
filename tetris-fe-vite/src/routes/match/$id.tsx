import { createFileRoute } from '@tanstack/react-router';
import { useTetrisBattle } from '@/hooks/useTetrisBattle.ts';
import Board from '@/components/Gameplay/Board.tsx';
import { useWebsocket } from '@/hooks/useWebsocket.ts';
import { useEffect, useMemo } from 'react';
import { useOpponent } from '@/hooks/useOpponent.ts';

export const Route = createFileRoute('/match/$id')({
  component: GameWindow,
});

function GameWindow() {
  const { id } = Route.useParams();
  const ws_url = useMemo(() => `ws://localhost:8080/ws/match?roomid=ABC12&playerid=${id}`, [id]);
  const { sendMsg, msgHandlerRegister } = useWebsocket({ ws_url });

  const { board, startGame, pauseGame, unpauseGame, isPlaying, isPaused, messageHandler } =
    useTetrisBattle(sendMsg);
  const { board: opponentBoard } = useOpponent(msgHandlerRegister);

  useEffect(() => {
    const unregisterStart = msgHandlerRegister('start', messageHandler);
    const unregisterPause = msgHandlerRegister('pause', messageHandler);
    const unregisterUnpause = msgHandlerRegister('unpause', messageHandler);
    const unregisterInputConfirmation = msgHandlerRegister('input-server', messageHandler);
    const unregisterGarbageConfirmation = msgHandlerRegister('garbage-sync', messageHandler);
    return () => {
      unregisterStart();
      unregisterPause();
      unregisterUnpause();
      unregisterInputConfirmation();
      unregisterGarbageConfirmation();
    };
  }, [msgHandlerRegister, messageHandler]);

  // Step 0: Gửi info message sau khi tạo WS connection
  useEffect(() => {
    const infoMessage = {
      type: 'info',
      playerid: id,
      payload: {},
      timestamp: Date.now(),
    };
    sendMsg(infoMessage);
    console.log('Sent info message:', infoMessage);
  }, [sendMsg, id]);

  // Step 1: Gửi sync_clock messages với interval
  // useEffect(() => {
  //   const syncInterval = setInterval(() => {
  //     const syncMessage = {
  //       type: 'sync_clock',
  //       playerid: id,
  //       payload: {},
  //       timestamp: Date.now(),
  //     };
  //     sendMsg(syncMessage);
  //     console.log('Sent sync_clock message:', syncMessage);
  //   }, 15); // 15ms interval
  //
  //   // Cleanup interval
  //   return () => clearInterval(syncInterval);
  // }, [sendMsg, id]);
  return (
    <>
      <div className="flex w-full h-full items-center justify-between border-4 relative">
        {/* Player 1 (sát lề trái) */}
        <div className="flex justify-center">
          <Board board={board} isOpponent={false} />
        </div>

        {/* Control Buttons ở giữa */}
        <div className="absolute left-1/2 -translate-x-1/2">
          {/* Start Button - chỉ hiện khi chưa bắt đầu game */}
          <button
            className={`bg-gray-300 hover:bg-gray-400 text-gray-800 font-bold py-2 px-4 rounded 
        transition-opacity duration-700 
        ${isPlaying ? 'opacity-0 pointer-events-none' : 'opacity-100'}`}
            onClick={startGame}
          >
            Start
          </button>

          {/* Pause/Unpause Button - chỉ hiện khi đang chơi */}
          {isPlaying && (
            <button
              className="bg-gray-300 hover:bg-gray-400 text-gray-800 font-bold py-2 px-4 rounded"
              onClick={isPaused ? unpauseGame : pauseGame}
            >
              {isPaused ? 'Resume' : 'Pause'}
            </button>
          )}
        </div>

        {/* Player 2 (sát lề phải) */}
        <div className="flex justify-center">
          <Board board={opponentBoard} isOpponent={true} />
        </div>
      </div>
    </>
  );
}
