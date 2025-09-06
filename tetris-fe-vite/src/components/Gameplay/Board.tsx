
import { BOARD_WIDTH, type BoardGrid, VISIBLE_HEIGHT } from "../../types/tetris";
import Cell from "./Cell";

const ROWS = VISIBLE_HEIGHT;
const COLS = BOARD_WIDTH;

export default function Board({ board }: { board: BoardGrid }) {
  
  return (
    <div
      className={`
        bg-black 
        border-b-4 border-l-4 border-r-4 border-gray-500
        pl-1 pr-1.5 pb-1
        grid
        gap-1
      `}
      style={{
        gridTemplateRows: `repeat(${ROWS}, minmax(0, 1fr))`,
        gridTemplateColumns: `repeat(${COLS}, minmax(0, 1fr))`,
        width: `${COLS * 20}px`, // 10 cột, mỗi cột 24px + gap 1px + padding 1px
        height: `${ROWS * 20}px`, // 20 hàng, mỗi hàng 24px + gap 1px + padding 1px
      }}
    >
      {board.flatMap((row, rowIndex) =>
        row.map((cell, colIndex) => (
          <Cell
            key={`${rowIndex}-${colIndex}`}
            type={cell.type || "empty"}
            size={20}
          />
        ))
      )}
    </div>
  );
}