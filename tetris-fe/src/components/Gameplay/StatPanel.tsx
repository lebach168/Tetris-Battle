"use client";

import { Block } from "@/types/tetris";
import NextBlock from "./NextBlock";

export default function StatPanel({ score, lines, combo, nextBlock }: { score: number; lines: number; combo: number,nextBlock:Block }) {
  return (
    <div className="flex flex-col items-end gap-1 text-sm text-gray-300 ">
        <NextBlock upcoming={nextBlock} />
      <p>Score: <span className="text-white">{score}</span></p>
      <p>Lines: <span className="text-white">{lines}</span></p>
      <p>Combo: <span className="text-white">{combo}</span></p>
    </div>
  );
}