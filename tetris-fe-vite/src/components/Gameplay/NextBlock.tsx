"use client";
import { renderShapeMatrix } from "@/lib/utils";
import { Block } from "@/types/tetris";

export default function NextBlock({ upcoming }: { upcoming: Block }) {
  return (
    <div className="flex flex-col items-center gap-1">
      <p className="text-sm text-gray-400">NEXT</p>
      {<div
          className="bg-black p-1 border border-gray-500 w-[72px] h-[72px] flex items-center justify-center"
        >
          {renderShapeMatrix(upcoming.shape, upcoming.type, 14)}
        </div>}
    </div>
  );
}