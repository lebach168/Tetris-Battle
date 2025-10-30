package game

import (
	"fmt"
	"math"
	"math/rand"
)

func ComputeDropSpeed(level int) int {
	if level == 0 {
		return 800
	}
	dropspeed := int(math.Max(1, float64((53-level)/2+1)))
	return dropspeed * (1000 / 60) // ms per cell
}

// GenerateList_7bag this func append N block to listblock
func GenerateList_7bag(list []int, n int) []int {
	blocks := []int{1, 2, 3, 4, 5, 6, 7}
	var shuffle func(list []int)
	shuffle = func(list []int) {
		for i := len(list) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			list[i], list[j] = list[j], list[i]
		}
	}
	for n > 0 {
		shuffle(blocks)
		list = append(list, blocks...)
		n -= 7
	}
	return list
}
func GenerateList_Classic(list []int, n int) []int {
	extend := make([]int, n)
	for i, _ := range extend {
		extend[i] = rand.Intn(7) + 1
	}
	list = append(list, extend...)
	return list
}
func CreateEmptyBoard() [][]int {
	board := make([][]int, BOARD_HEIGHT)
	for i := range board {
		board[i] = make([]int, BOARD_WIDTH)
	}
	return board
}

// RotateRight rotate SHAPE
func RotateRight(shape [][]int) [][]int {
	size := len(shape)
	rotated := make([][]int, size)
	for i := range rotated {
		rotated[i] = make([]int, size)
	}
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			rotated[i][j] = shape[size-1-j][i]
		}
	}
	return rotated
}

// RotateLeft for rrotate SHAPE
func RotateLeft(shape [][]int) [][]int {
	size := len(shape)
	rotated := make([][]int, size)
	for i := range rotated {
		rotated[i] = make([]int, size)
	}
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			rotated[i][j] = shape[j][size-1-i]
		}
	}
	return rotated
}

func copySlice(src [][]int) [][]int {
	dst := make([][]int, len(src))
	for i := range src {
		dst[i] = make([]int, len(src[i]))
		copy(dst[i], src[i])
	}
	return dst
}

func hasCollision(board [][]int, shape [][]int, row, col int) bool {
	for y := 0; y < len(shape); y++ {
		for x := 0; x < len(shape[y]); x++ {
			if shape[y][x] == 0 {
				continue
			}
			boardY := row + y
			boardX := col + x

			if boardY < 0 || boardY >= len(board) || boardX < 0 || boardX >= len(board[0]) {
				return true
			}
			if board[boardY][boardX] != 0 {
				return true
			}
		}
	}
	return false
}

// FindLandingPos return row index after hard drop (space inputBuffer)
func FindLandingPosition(board [][]int, block [][]int, row, col int) int {
	shapeHeight := 0
	for i := range block {
		for _, val := range block[i] {
			if val != 0 {
				shapeHeight++
				break
			}
		}
	}
	landingRow := row
	for testRow := row + 1; testRow <= BOARD_HEIGHT; testRow++ {
		if hasCollision(board, block, testRow, col) {
			landingRow = testRow - 1 // Dừng ở row trước khi collision
			break
		}
		if testRow == BOARD_HEIGHT {
			landingRow = BOARD_HEIGHT - shapeHeight
		}
	}

	return landingRow
}

func CheckGameOver(board [][]int, nextBlock [][]int) bool {
	if hasCollision(board, nextBlock, 0, 4) {
		return true
	}
	return false
}

type SpecialMove struct {
}

// clearLines return number of rows has cleared and update board after applied board
func clearLines(board [][]int) int {
	var newBoard [][]int
	cleared := 0

	for _, row := range board {
		full := true
		for _, cell := range row {
			if cell == 0 {
				full = false
				break
			}
		}
		if !full {
			newBoard = append(newBoard, row)
		} else {
			cleared++
		}
	}
	ib := BOARD_HEIGHT - 1
	if cleared > 0 {
		for i := len(newBoard) - 1; i >= 0; i, ib = i-1, ib-1 {
			for j := 0; j < BOARD_WIDTH; j++ {
				board[ib][j] = newBoard[i][j]
			}
		}
		for ; ib >= 0; ib-- {
			for j := 0; j < BOARD_WIDTH; j++ {
				board[ib][j] = 0
			}
		}
	}

	return cleared
}
func placeBlock(gs *GameState, ps *PlayerState, block [][]int, row, col int) {
	//fmt.Printf("place block row:%v col:%v\n block:%v\n", row, col, block)
	for i := range block {
		for j := range block[i] {
			if block[i][j] != 0 {
				ps.board[row+i][col+j] = block[i][j]
			}
		}
	}

}
func propagateState(previous *PlayerState, ps *PlayerState) {
	ps.board = copySlice(previous.board)
	ps.dropSpeed = previous.dropSpeed
	ps.cRow = previous.cRow
	ps.cCol = previous.cCol
	ps.gravityTimer = previous.gravityTimer
	ps.lockTimer = previous.lockTimer
	ps.block = previous.block
	ps.canHold = previous.canHold
	ps.onGround = previous.onGround
	ps.blockIndex = previous.blockIndex
	ps.holdBlock = previous.holdBlock
}
func applyInputBuffer(state *GameState, ps *PlayerState, input InputBuffer) {
	// Order apply: Horizontal move -> Rotate -> Vertical drop -> Hold -> hard drop last
	// Horizontal moves (left/right)
	if input[left] {
		ps.cCol--
		if hasCollision(ps.board, ps.block.shape, ps.cRow, ps.cCol) {
			ps.cCol++ // Revert
		}
	}
	if input[right] {
		ps.cCol++
		if hasCollision(ps.board, ps.block.shape, ps.cRow, ps.cCol) {
			ps.cCol-- // Revert
		}
	}

	// Rotate (rotate clockwise, rrotate counterclockwise)
	if input[rotate] || input[rrotate] {

		rotatedShape := copySlice(ps.block.shape)
		x := 1
		if ok := input[rotate]; ok {
			rotatedShape = RotateRight(rotatedShape)
		} else {
			rotatedShape = RotateLeft(rotatedShape)
			x = 3 //3 = -1 in modula
		}
		form := ps.block.form
		newForm := (form + x) % 4
		if !hasCollision(ps.board, rotatedShape, ps.cRow, ps.cCol) {
			ps.block.shape = rotatedShape
			ps.block.form = newForm

		} else {
			wallkickOffset := GetWallKickData(len(ps.block.shape), form, newForm)
			for _, d := range wallkickOffset {
				newRow := ps.cRow + d[1]
				newCol := ps.cCol + d[0]
				if !hasCollision(ps.board, rotatedShape, newRow, newCol) {
					fmt.Printf("Applying wall kick: dx=%v, dy=%v\n", d[0], d[1])
					ps.cRow = newRow
					ps.cCol = newCol
					ps.block.shape = rotatedShape
					ps.block.form = newForm
					break
				}
			}
		}

	}

	// Vertical moves (down soft drop key down set dropspeed -> 100ms)
	if input[down] {
		ps.dropSpeed = SOFT_DROP
	}
	if input[downOff] {
		ps.dropSpeed = DROPSPEED
	}

	// Hold
	if input[hold] && ps.canHold {
		holdBlock := state.listBlock[ps.blockIndex]
		if ps.holdBlock == 0 {
			ps.blockIndex++
			ps.block = Tetromino[state.listBlock[ps.blockIndex]]

		} else {
			ps.block = Tetromino[ps.holdBlock]
		}
		ps.holdBlock = holdBlock
		ps.cRow = 0
		ps.cCol = 4
		ps.canHold = false
	}
	if input[spacebar] {
		ps.cRow = FindLandingPosition(ps.board, ps.block.shape, ps.cRow, ps.cCol)
		ps.onGround = true
		ps.lockTimer = LOCKDELAY
	}
}
func spawnNewPiece(state *GameState, ps *PlayerState) {
	ps.cRow = 0
	ps.cCol = 4
	ps.blockIndex++
	ps.block = Tetromino[state.listBlock[ps.blockIndex]]
	ps.onGround = false
	ps.canHold = true
	ps.lockTimer = 0
	ps.gravityTimer = 0
}
