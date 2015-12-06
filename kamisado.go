package main

import (
	"errors"
	"math"
)

const (
	humanPlayer = 0
	cpuPlayer   = 1
)

var boardColors [8][8]string = [8][8]string{
	{"orange", "blue", "purple", "pink", "yellow", "red", "green", "brown"}, // player 1 (black, CPU)
	{"red", "orange", "pink", "green", "blue", "yellow", "brown", "purple"},
	{"green", "pink", "orange", "red", "purple", "brown", "yellow", "blue"},
	{"pink", "purple", "blue", "orange", "brown", "green", "red", "yellow"},
	{"yellow", "red", "green", "brown", "orange", "blue", "purple", "pink"},
	{"blue", "yellow", "brown", "purple", "red", "orange", "pink", "green"},
	{"purple", "brown", "yellow", "blue", "green", "pink", "orange", "red"},
	{"brown", "green", "red", "yellow", "pink", "purple", "blue", "orange"}} // player 0 (white, human)

var N int = len(boardColors)

type coord struct {
	i, j int
}

type piece struct {
	// fields must be capitalized to be JSON exportable
	Player int    `json:"player"`
	Color  string `json:"color"`
}

type state struct {
	board             [8][8]*piece
	playerPieceCoords [2]map[string]coord // int -> color -> coord
}

func (state *state) copy() *state {
	// the board grid is getting copied because it's an array,
	// but the playerPieceCoords structure must be deeply copied
	newState := *state
	newState.playerPieceCoords = [2]map[string]coord{}
	for player := 0; player < 2; player++ {
		newState.playerPieceCoords[player] = make(map[string]coord)
		for color, coord := range state.playerPieceCoords[player] {
			newState.playerPieceCoords[player][color] = coord
		}
	}
	return &newState
}

// "a1" -> Coord{7, 0}
// "d5" -> Coord{3, 3}
func toCoord(a []string) (coord, error) {
	if len(a) != 2 {
		return coord{-1, -1}, errors.New(`Coord must have two elements`)
	}
	// use ascii code conversion
	j := int(a[0][0]) - 97       // 'a' -> 0, 'h' -> 7
	i := N - (int(a[1][0]) - 48) // '0' -> 0, '7' -> 7
	if i < 0 || j < 0 || i >= N || j >= N {
		return coord{-1, -1}, errors.New(`Bad coord`)
	}
	return coord{i, j}, nil
}

func (state *state) movePiece(player int, color string, dst coord) {
	src := state.playerPieceCoords[player][color]
	piece := state.board[src.i][src.j]
	state.board[dst.i][dst.j] = piece
	state.board[src.i][src.j] = nil
	state.playerPieceCoords[player][color] = coord{dst.i, dst.j}
}

func (state *state) getPossibleMoveCoords(player int, color string) []coord {
	incrs := [3]coord{coord{1, -1}, coord{1, 0}, coord{1, 1}}
	coords := []coord{}
	src := state.playerPieceCoords[player][color]
	piece := state.board[src.i][src.j]
	m := 1
	if piece.Player == humanPlayer {
		m = -1 // reverse direction of coord.i component
	}
	for n := 0; n < 3; n++ { // cycle through 3 i directions
		i, j := src.i, src.j
		for {
			i += incrs[n].i * m
			j += incrs[n].j
			if i < 0 || i > (N-1) || j < 0 || j > (N-1) || state.board[i][j] != nil {
				break
			}
			coords = append(coords, coord{i, j})
		}
	}
	return coords
}

func (state *state) findBestMoveCoord(player int, color string, depth int) coord {
	dstCoords := state.getPossibleMoveCoords(player, color)
	var bestCoord coord
	bestValue := math.Inf(-1)
	for _, dst := range dstCoords {
		newState := state.copy()
		newState.movePiece(player, color, dst)
		nextColor := boardColors[dst.i][dst.j]
		v := -negamax(newState, player, player, nextColor, depth)
		if v > bestValue {
			bestCoord = dst
			bestValue = v
		}
	}
	return bestCoord
}

func (state *state) isWinning(player int) bool {
	var i int
	if player == humanPlayer { // if white, check top row
		i = 0
	} else {
		i = N - 1 // if black, check bottom row
	}
	for j := 0; j < N; j++ {
		piece := state.board[i][j]
		// if you find a player's piece in the target row, they won
		if piece != nil && piece.Player == player {
			return true
		}
	}
	return false
}

func (state *state) value(player int) float64 {
	opponent := (player + 1) % 2
	pos := state.getNumberOfWinInOnePlayerPieces(player)
	neg := state.getNumberOfWinInOnePlayerPieces(opponent)
	neg += state.getNumberDistinctColorsForNextMove(opponent)
	return float64(pos - neg)
}

func (state *state) getNumberOfWinInOnePlayerPieces(player int) int {
	nWinningPieces := 0
	var winningRow int
	if player == humanPlayer {
		winningRow = 0
	} else {
		winningRow = N - 1
	}
	for color, _ := range state.playerPieceCoords[player] {
		moveCoords := state.getPossibleMoveCoords(player, color)
		for _, nextCoord := range moveCoords {
			if nextCoord.i == winningRow {
				nWinningPieces++
				break
			}
		}
	}
	return nWinningPieces
}

func (state *state) getNumberDistinctColorsForNextMove(player int) int {
	colors := make(map[string]bool)
	n := 0
	for color, _ := range state.playerPieceCoords[player] {
		moveCoords := state.getPossibleMoveCoords(player, color)
		for _, nextCoord := range moveCoords {
			nextColor := boardColors[nextCoord.i][nextCoord.j]
			if _, ok := colors[nextColor]; !ok {
				colors[nextColor] = true
				n++
				break
			}
		}
	}
	return n
}

func negamax(state *state, initPlayer int, currPlayer int, color string, depth int) float64 {
	nextPlayer := (currPlayer + 1) % 2
	m := float64(1)
	if currPlayer == initPlayer {
		m = float64(-1)
	}
	if state.isWinning(currPlayer) {
		return m * math.Inf(1)
	} else if state.isWinning(nextPlayer) {
		return m * math.Inf(-1)
	} else if depth == 0 {
		return m * state.value(initPlayer)
	}
	dstCoords := state.getPossibleMoveCoords(nextPlayer, color)
	bestValue := float64(-1)
	foundMove := false
	for _, dst := range dstCoords {
		nextColor := boardColors[dst.i][dst.j]
		newState := state.copy()
		newState.movePiece(nextPlayer, nextColor, dst)
		v := -negamax(newState, nextPlayer, nextPlayer, nextColor, depth-1)
		if v > bestValue {
			bestValue = v
			foundMove = true
		}
	}
	if foundMove {
		return bestValue
	} else {
		// board stays the same (src == dst), and next color is src's one
		src := state.playerPieceCoords[nextPlayer][color]
		return negamax(state, initPlayer, nextPlayer, boardColors[src.i][src.j], depth-1)
	}
}

func (state *state) isLegalMove(player int, color string, dst coord) bool {
	dstCoords := state.getPossibleMoveCoords(player, color)
	for _, dstCoord := range dstCoords {
		if dst == dstCoord {
			return true
		}
	}
	return false
}

func (state *state) isBlocked(player int, color string) bool {
	return len(state.getPossibleMoveCoords(player, color)) == 0
}
