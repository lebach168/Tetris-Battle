"use client";
import { RenderShapeMatrix } from "@/components/Unit/RenderedMatrix.tsx";
import { Block } from "@/types/tetris"; 


export default function Hold({ holdBlock }: { holdBlock?: Block }) {
  return (
    <div className="flex flex-col items-center">
      <p className="text-sm text-gray-400">HOLD</p>
      <div className="bg-black p-1 border-2 border-gray-500 w-[80px] h-[80px] flex items-center justify-center">
        {holdBlock ? (
          RenderShapeMatrix(holdBlock.shape, holdBlock.type)
        ) : (
          <div className="text-gray-600 text-xs">None</div>
        )}
      </div>
    </div>
  );
}