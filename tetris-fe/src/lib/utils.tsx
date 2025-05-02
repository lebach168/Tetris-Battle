import Cell from "@/components/Gameplay/Cell";

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
        {shape.flatMap((row, y) =>
          row.map((cell, x) => (
            <Cell key={`${y}-${x}`} type={cell ? type : "0"} size={size} />
          ))
        )}
      </div>
    );
  }