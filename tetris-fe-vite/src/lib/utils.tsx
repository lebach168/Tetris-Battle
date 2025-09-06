import Cell from "@/components/Gameplay/Cell";
import { ReverseTetrominoMap, TetrominoMap, TetrominoType } from "@/types/tetris";

export function renderShapeMatrix(shape: number[][], type: string, size = 18) {
  const rows = shape.length;
  const cols = shape[0].length;

  return (
    <div
      className="grid gap-[1px]"
      style={{
        gridTemplateRows: `repeat(${rows}, ${size}px)`,
        gridTemplateColumns: `repeat(${cols}, ${size}px)`,
      }}
    >
      {shape.flatMap((row, y) => row.map((cell, x) => <Cell key={`${y}-${x}`} type={cell ? type : "0"} size={size} />))}
    </div>
  );
}


export function convertTetrominoToNumArray(arr: TetrominoType[]): number[] {
  return arr.map(t => TetrominoMap[t]);
}

export function convertNumToTetrominoArray(arr:number[]):TetrominoType[]{
  return arr.map(n=>ReverseTetrominoMap[n]);
}