// import { BackToBackType } from "@/types/tetris";
// import React, { useReducer } from "react"

// type GameStat ={
//     score?:number;//classic mode only
//     combo:number;
//     backToBack:BackToBackType;//back to back type 
//     garbageLines:number;// for canceling garbage
//     pendingGarbage:number;
//     level: number;
//     linesCleared:number;
// }
// const initialState: GameStat = {
//     score: 0,
//     combo: 0,
//     backToBack: "None",
//     garbageLines: 0,
//     pendingGarbage: 0,
//     level: 1,
//     linesCleared: 0,
//   };
// type StatAction ={
//     type:'clear'|'',
//     payload:{

//     }
// }
// const statReducer = (action:StatAction):GameStat{
//     switch(action.type){
        
//     }
// }
// export const useGameStat = ():[stat:GameStat, dispatchStat:React.Dispatch<StatAction>] =>{
//     const [gameStat, dispatchStat] = useReducer(statReducer,initialState);

//     return [gameStat, dispatchStat];
// }