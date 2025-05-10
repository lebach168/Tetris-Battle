"use client";

import React from "react";

import PlayerSide from "./PlayerSide";
import OpponentSide from "./OpponentSide";
import { useTetris } from "@/hooks/useTetris";
import { useTetrisEventSource } from "@/hooks/useTetrisEventSource";

export default function GameScreen() {
  const { board, startGame, isPlaying, isReady } = useTetris();
  const { board: opponentBoard } = useTetrisEventSource();
  return (
    <div>
      {/* Button start - chỉ hiện khi chưa bắt đầu */}
      {!isPlaying && (
        <button
          className="bg-gray-300 hover:bg-gray-400 text-gray-800 font-bold py-2 px-4 rounded h-min self-center"
          onClick={startGame}
          disabled={!isReady}
        >
          Start
        </button>
      )}
      <div className="flex w-full h-full border-4">
        {/* Player Side */}
        <div className="flex-1 ">
          <PlayerSide board={board} />
        </div>

        {/* Opponent Side */}
        <div className="flex-1">
          <OpponentSide board={opponentBoard} />
        </div>
      </div>
    </div>
  );
}

GameScreen;
