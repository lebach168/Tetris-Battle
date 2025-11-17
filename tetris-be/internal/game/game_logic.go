package game

import (
	"fmt"
	"math"
	"math/rand"
)

type BoardState struct {
	board [][]int

	blockIndex   int
	block        Block
	holdBlock    int //1->7 convert to Block : Tetromino[int]
	cRow         int
	cCol         int
	canHold      bool
	dropSpeed    float64
	inputBuffer  InputBuffer // map-fat pointer
	gravityTimer float64
	lockTimer    float64
	onGround     bool

	//for attack mechanism
	combo  int
	b2b    string //back to back
	send   int
	cancel int
}

func NewBoardState(board [][]int, blockIndex int, block Block, holdBlock int, cRow, cCol int, canHold bool,
	dropSpeed float64, input InputBuffer, accumulator float64, onGround bool) *BoardState {
	return &BoardState{
		board:        board,
		blockIndex:   blockIndex,
		block:        block,
		holdBlock:    holdBlock, //0 is empty value
		cRow:         cRow,
		cCol:         cCol,
		canHold:      canHold,
		dropSpeed:    dropSpeed,
		inputBuffer:  input,
		gravityTimer: accumulator,
		onGround:     onGround,
		combo:        0,
	}
}
func NewDefaultBoardState() *BoardState {
	return &BoardState{
		board:       CreateEmptyBoard(),
		canHold:     true,
		dropSpeed:   DROPSPEED,
		inputBuffer: make(InputBuffer),
	}
}

type key string

const (
	down     key = "down"
	downOff  key = "downOff"
	left     key = "left"
	right    key = "right"
	rotate   key = "rotate" //arrow up
	rrotate  key = "rrotate"
	spacebar key = "space"
	hold     key = "hold"
)

type InputBuffer map[key]bool

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
	for i := range extend {
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

func CheckGameOver(board [][]int, block [][]int, cRow, cCol int) bool {
	if hasCollision(board, block, cRow, cCol) {
		return true
	}
	return false
}

// ClearLines return number of rows has cleared and update board after applied board
func ClearLines(board [][]int) int {
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
			//Todo: potential bug
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
func PlaceBlock(board [][]int, block [][]int, row, col int) {
	//fmt.Printf("place block row:%v col:%v\n block:%v\n", row, col, block)
	for i := range block {
		for j := range block[i] {
			if block[i][j] != 0 {
				board[row+i][col+j] = block[i][j]
			}
		}
	}

}
func PropagateState(previous *BoardState, bs *BoardState) {
	bs.board = copySlice(previous.board)
	bs.dropSpeed = previous.dropSpeed
	bs.cRow = previous.cRow
	bs.cCol = previous.cCol
	bs.gravityTimer = previous.gravityTimer
	bs.lockTimer = previous.lockTimer
	bs.block = previous.block
	bs.canHold = previous.canHold
	bs.onGround = previous.onGround
	bs.blockIndex = previous.blockIndex
	bs.holdBlock = previous.holdBlock

	bs.combo = previous.combo
	bs.send = previous.send
	bs.cancel = previous.cancel
}

func ApplyInputBuffer(listBlock []int, bs *BoardState, input InputBuffer) {
	// Order apply: Horizontal move -> Rotate -> Vertical drop -> Hold -> hard drop last
	// Horizontal moves (left/right)
	if input[left] {
		bs.cCol--
		if hasCollision(bs.board, bs.block.shape, bs.cRow, bs.cCol) {
			bs.cCol++ // Revert
		}
	}
	if input[right] {
		bs.cCol++
		if hasCollision(bs.board, bs.block.shape, bs.cRow, bs.cCol) {
			bs.cCol-- // Revert
		}
	}

	// Rotate (rotate clockwise, rrotate counterclockwise)
	if input[rotate] || input[rrotate] {

		rotatedShape := copySlice(bs.block.shape)
		x := 1
		if ok := input[rotate]; ok {
			rotatedShape = RotateRight(rotatedShape)
		} else {
			rotatedShape = RotateLeft(rotatedShape)
			x = 3 //3 = -1 in modula
		}
		form := bs.block.form
		newForm := (form + x) % 4
		if !hasCollision(bs.board, rotatedShape, bs.cRow, bs.cCol) {
			bs.block.shape = rotatedShape
			bs.block.form = newForm

		} else {
			wallkickOffset := GetWallKickData(len(bs.block.shape), form, newForm)
			for _, d := range wallkickOffset {
				newRow := bs.cRow + d[1]
				newCol := bs.cCol + d[0]
				if !hasCollision(bs.board, rotatedShape, newRow, newCol) {
					fmt.Printf("Applying wall kick: dx=%v, dy=%v\n", d[0], d[1])
					bs.cRow = newRow
					bs.cCol = newCol
					bs.block.shape = rotatedShape
					bs.block.form = newForm
					break
				}
			}
		}

	}

	// Vertical moves (down soft drop key down set dropspeed -> 100ms)
	if input[down] {
		bs.dropSpeed = SOFT_DROP
	}
	if input[downOff] {
		bs.dropSpeed = DROPSPEED
	}

	// Hold
	if input[hold] && bs.canHold {
		holdBlock := listBlock[bs.blockIndex]
		if bs.holdBlock == 0 {
			bs.blockIndex++
			bs.block = Tetromino[listBlock[bs.blockIndex]]

		} else {
			bs.block = Tetromino[bs.holdBlock]
		}
		bs.holdBlock = holdBlock
		bs.cRow = 0
		bs.cCol = 4
		bs.canHold = false
	}
	if input[spacebar] {
		bs.cRow = FindLandingPosition(bs.board, bs.block.shape, bs.cRow, bs.cCol)
		bs.onGround = true
		bs.lockTimer = LOCKDELAY
	}
}
func SpawnNewPiece(listBlock []int, bs *BoardState) {
	bs.cRow = 0
	bs.cCol = 4
	bs.blockIndex++
	bs.block = Tetromino[listBlock[bs.blockIndex]]
	bs.onGround = false
	bs.canHold = true
	bs.lockTimer = 0
	bs.gravityTimer = 0
}
func isPerfect(board [][]int) bool {
	for _, row := range board {
		for _, cell := range row {
			if cell != 0 {
				return false
			}
		}
	}
	return true
}
func TakeGarbage(lines int, board [][]int) {
	if lines == 0 {
		return
	}
	for r := 0; r < BOARD_HEIGHT-lines; r++ {
		for c := 0; c < BOARD_WIDTH; c++ {
			board[r][c] = board[r+lines][c]
		}
	}
	emptyCol := rand.Intn(BOARD_WIDTH)
	// Thêm garbage lines vào dưới cùng
	for i := 0; i < lines; i++ {
		row := BOARD_HEIGHT - lines + i
		for c := 0; c < BOARD_WIDTH; c++ {
			board[row][c] = 8 // 8: garbage value
		}
		board[row][emptyCol] = 0
	}

}
