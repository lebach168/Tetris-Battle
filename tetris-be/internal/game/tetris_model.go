package game

import (
	"fmt"
	"time"
)

const BOARD_WIDTH = 10
const BOARD_HEIGHT = 22
const SOFT_DROP = 100
const LOCKDELAY float64 = 300
const DROPSPEED float64 = 800
const QUEUE_SIZE = 100
const TICK = 30

var INTERVAL = float64(time.Second.Milliseconds()) / float64(time.Duration(TICK))

var Tetromino = map[int]Block{
	1: { //I
		shape: [][]int{
			{0, 0, 0, 0},
			{1, 1, 1, 1},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		form: 0,
	},
	2: { //O
		shape: [][]int{
			{1, 1},
			{1, 1},
		},
		form: 0,
	},
	3: { //T
		shape: [][]int{
			{0, 1, 0},
			{1, 1, 1},
			{0, 0, 0},
		},
		form: 0,
	},
	4: { //S
		shape: [][]int{
			{0, 2, 2},
			{2, 2, 0},
			{0, 0, 0},
		},
		form: 0,
	},
	5: { //Z
		shape: [][]int{
			{3, 3, 0},
			{0, 3, 3},
			{0, 0, 0},
		},
		form: 0,
	},
	6: { //J
		shape: [][]int{
			{2, 0, 0},
			{2, 2, 2},
			{0, 0, 0},
		},
		form: 0,
	},
	7: { //L
		shape: [][]int{
			{0, 0, 3},
			{3, 3, 3},
			{0, 0, 0},
		},
		form: 0,
	},
}

var ReverseTetrominoMap = map[int]string{
	1: "I", 2: "O", 3: "T", 4: "S", 5: "Z", 6: "J", 7: "L",
}

type Block struct {
	shape [][]int
	form  int //0,1,2,3,
}
type KickOffsets [][2]int
type WallKickTable map[string]KickOffsets

var WallKickJLSTZ = WallKickTable{
	"0->1": {{0, 0}, {-1, 0}, {-1, 1}, {0, -2}, {-1, -2}}, // 0->R
	"1->2": {{0, 0}, {1, 0}, {1, -1}, {0, 2}, {1, 2}},     // R->2
	"2->3": {{0, 0}, {1, 0}, {1, 1}, {0, -2}, {1, -2}},    // 2->L
	"3->0": {{0, 0}, {-1, 0}, {-1, -1}, {0, 2}, {-1, 2}},  // L->0
	"1->0": {{0, 0}, {1, 0}, {1, -1}, {0, 2}, {1, 2}},     // R->0
	"2->1": {{0, 0}, {-1, 0}, {-1, 1}, {0, -2}, {-1, -2}}, // 2->R
	"3->2": {{0, 0}, {-1, 0}, {-1, -1}, {0, 2}, {-1, 2}},  // L->2
	"0->3": {{0, 0}, {1, 0}, {1, 1}, {0, -2}, {1, -2}},    // 0->L
}
var WallKickI = WallKickTable{
	"0->1": {{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}, // 0->R
	"1->2": {{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}, // R->2
	"2->3": {{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -2}}, // 2->L
	"3->0": {{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {-2, 1}}, // L->0
	"1->0": {{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -2}}, // R->0
	"2->1": {{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {-2, 1}}, // 2->R
	"3->2": {{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}, // L->2
	"0->3": {{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}, // 0->L
}
var WallKickIAkira = WallKickTable{
	"0->1": {{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}, // 0->R
	"1->2": {{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}, // R->2
	"2->3": {{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -1}}, // 2->L
	"3->0": {{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {-2, 1}}, // L->0
	"1->0": {{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -2}}, // R->0
	"2->1": {{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {1, -1}}, // 2->R
	"3->2": {{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}, // L->2
	"0->3": {{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}, // 0->L
}

func GetWallKickData(size int, fromShape, toShape int) KickOffsets {
	//if size ==2 -> tetromino O -> no need wallkick for this
	//if size == 4 -> tetromino I -> game for only I
	if size == 2 {
		return KickOffsets{{0, 0}}
	}
	k := fmt.Sprintf("%d->%d", fromShape, toShape)
	if size == 4 {
		return WallKickI[k]
	}
	return WallKickJLSTZ[k]
}
