"use client"

import { useTetris } from "@/hooks/useTetris";
import Board from "./Board";

export default function GameplayWindow(){
    const {board,startGame,isPlaying} = useTetris();
  
    return(
        <div className="flex justify-center items-center min-h-screen bg-gray-900">
      <div className="flex gap-8">
        <Board board={board}></Board>
        
      </div>
      <div className="start">{isPlaying ? null :
                (<button className="bg-gray-300 hover:bg-gray-400 text-gray-800 font-bold py-2 px-4 rounded inline-flex items-center" onClick={startGame}>Start</button>)}</div>
    </div>
        
    )
}