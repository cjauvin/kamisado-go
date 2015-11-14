package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"

	"github.com/fatih/color"
)

type Piece struct {
	player int
	color  string
}

var boardColors [8][8]string = [8][8]string{
	{"orange", "blue", "purple", "pink", "yellow", "red", "green", "brown"}, // player 1 (black)
	{"red", "orange", "pink", "green", "blue", "yellow", "brown", "purple"},
	{"green", "pink", "orange", "red", "purple", "brown", "yellow", "blue"},
	{"pink", "purple", "blue", "orange", "brown", "green", "red", "yellow"},
	{"yellow", "red", "green", "brown", "orange", "blue", "purple", "pink"},
	{"blue", "yellow", "brown", "purple", "red", "orange", "pink", "green"},
	{"purple", "brown", "yellow", "blue", "green", "pink", "orange", "red"},
	{"brown", "green", "red", "yellow", "pink", "purple", "blue", "orange"}} // player 0 (white)

var N int = len(boardColors)

type Board [8][8]*Piece

type Coord struct {
	i, j int
}

type State struct {
	board             Board
	playerPieceCoords [2]map[string]Coord // int -> color -> Coord
}

func (state *State) Copy() *State {
	// the board grid is getting copied because it's an array,
	// but the playerPieceCoords structure must be deeply copied
	newState := *state
	newState.playerPieceCoords = [2]map[string]Coord{}
	for player := 0; player < 2; player++ {
		newState.playerPieceCoords[player] = make(map[string]Coord)
		for color, coord := range state.playerPieceCoords[player] {
			newState.playerPieceCoords[player][color] = coord
		}
	}
	return &newState
}

func ToCoord(a []string) Coord {
	i, _ := strconv.Atoi(a[0])
	j, _ := strconv.Atoi(a[1])
	return Coord{i, j}
}

func (state *State) MovePiece(player int, color string, dst Coord) {
	src := state.playerPieceCoords[player][color]
	piece := state.board[src.i][src.j]
	state.board[dst.i][dst.j] = piece
	state.board[src.i][src.j] = nil
	state.playerPieceCoords[player][color] = Coord{dst.i, dst.j}
}

func (state *State) PrintBoard() {
	board := state.board
	type ColorPair struct {
		fg color.Attribute
		bg color.Attribute
	}
	colors := map[string]ColorPair{
		"orange": ColorPair{color.FgHiRed, color.BgHiRed},
		"blue":   ColorPair{color.FgCyan, color.BgCyan},
		"purple": ColorPair{color.FgBlue, color.BgBlue},
		"pink":   ColorPair{color.FgHiMagenta, color.BgHiMagenta},
		"yellow": ColorPair{color.FgHiYellow, color.BgHiYellow},
		"red":    ColorPair{color.FgRed, color.BgRed},
		"green":  ColorPair{color.FgGreen, color.BgGreen},
		"brown":  ColorPair{color.FgHiBlack, color.BgHiBlack},
	}
	fmt.Println()
	fmt.Print("    ")
	for j := 0; j < N; j++ {
		fmt.Print(fmt.Sprintf("%d    ", j))
	}
	fmt.Println()
	for i := 0; i < N; i++ {
		for k := 0; k < 3; k++ {
			if k == 1 {
				fmt.Print(fmt.Sprintf(" %d", i))
			} else {
				fmt.Print("  ")
			}
			for j := 0; j < N; j++ {
				cp := colors[boardColors[i][j]]
				color.Set(cp.bg)
				fmt.Print("  ")
				piece := board[i][j]
				if k == 1 && piece != nil {
					color.Unset()
					color.Set(colors[piece.color].fg)
					if piece.player == 0 {
						fmt.Print("X")
					} else {
						fmt.Print("O")
					}
				} else {
					fmt.Print(" ")
				}
				color.Set(cp.bg)
				fmt.Print("  ")
			}
			color.Unset()
			if k == 1 {
				fmt.Print(i)
			}
			fmt.Println()
		}
	}
	fmt.Print("    ")
	for j := 0; j < N; j++ {
		fmt.Print(fmt.Sprintf("%d    ", j))
	}
	fmt.Println()
	fmt.Println()
}

func (state *State) getPossibleMoveCoords(player int, color string) []Coord {
	incrs := [3]Coord{Coord{1, -1}, Coord{1, 0}, Coord{1, 1}}
	coords := []Coord{}
	pc := state.playerPieceCoords[player][color]
	piece := state.board[pc.i][pc.j]
	m := 1
	if piece.player == 0 {
		m = -1 // reverse direction of Coord.i component
	}
	for n := 0; n < 3; n++ { // cycle through 3 i directions
		i, j := pc.i, pc.j
		for {
			i += incrs[n].i * m
			j += incrs[n].j
			if i < 0 || i > (N-1) || j < 0 || j > (N-1) || state.board[i][j] != nil {
				break
			}
			coords = append(coords, Coord{i, j})
		}
	}
	return coords
}

func (state *State) findBestMoveCoord(player int, color string, depth int) Coord {
	dstCoords := state.getPossibleMoveCoords(player, color)
	var bestCoord Coord
	bestValue := math.Inf(-1)
	for _, dst := range dstCoords {
		newState := state.Copy()
		newState.MovePiece(player, color, dst)
		nextColor := boardColors[dst.i][dst.j]
		v := negamax(newState, player, player, nextColor, depth)
		if v > bestValue {
			bestCoord = dst
			bestValue = v
		}
	}
	return bestCoord
}

func (state *State) IsWinning(player int) bool {
	var i int
	if player == 0 { // if white, check top row
		i = 0
	} else {
		i = N - 1 // if black, check bottom row
	}
	for j := 0; j < N; j++ {
		piece := state.board[i][j]
		// if you find a player's piece in the target row, they won
		if piece != nil && piece.player == player {
			return true
		}
	}
	return false
}

func (state *State) Value(player int) float64 {
	opponent := (player + 1) % 2
	pos := state.GetNumberOfWinInOnePlayerPieces(player)
	neg := state.GetNumberOfWinInOnePlayerPieces(opponent)
	neg += state.GetNumberDistinctColorsForNextMove(opponent)
	return float64(pos - neg)
}

func (state *State) GetNumberOfWinInOnePlayerPieces(player int) int {
	nWinningPieces := 0
	var winningRow int
	if player == 0 {
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

func (state *State) GetNumberDistinctColorsForNextMove(player int) int {
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

func negamax(state *State, initPlayer int, currPlayer int, color string, depth int) float64 {
	nextPlayer := (currPlayer + 1) % 2
	m := float64(1)
	if currPlayer == initPlayer {
		m = -1
	}
	if state.IsWinning(currPlayer) {
		return m * math.Inf(1)
	} else if state.IsWinning(nextPlayer) {
		return m * math.Inf(-1)
	} else if depth == 0 {
		return m * state.Value(initPlayer)
	}
	dstCoords := state.getPossibleMoveCoords(nextPlayer, color)
	bestValue := float64(-1)
	foundMove := false
	for _, dst := range dstCoords {
		nextColor := boardColors[dst.i][dst.j]
		newState := state.Copy()
		newState.MovePiece(nextPlayer, nextColor, dst)
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

func main() {

	depth := 3

	board := Board{}
	playerPieceCoords := [2]map[string]Coord{}

	// Note that the 16 Piece objects created here will be the only
	// ones used throughout the entire program.
	for player, i := range []int{7, 0} {
		playerPieceCoords[player] = make(map[string]Coord)
		for j := 0; j < 8; j++ {
			color := boardColors[i][j]
			board[i][j] = &Piece{player, color}
			playerPieceCoords[player][color] = Coord{i, j}
		}
	}

	state := State{board, playerPieceCoords}

	fmt.Println("Welcome to Kamisado! You play White (X)")
	state.PrintBoard()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your move (src/dst coords as 4 numbers): ")
	text, _ := reader.ReadString('\n')
	if text[0] == 'q' {
		os.Exit(0)
	}

	r, _ := regexp.Compile(`\d+`)
	a := r.FindAllString(text, -1)

	var humanSrcCoord, humanDstCoord, cpuDstCoord Coord
	var humanSrcColor, cpuSrcColor string

	humanSrcCoord = ToCoord(a[:2])
	humanSrcColor = boardColors[humanSrcCoord.i][humanSrcCoord.j]

	humanDstCoord = ToCoord(a[2:])
	state.MovePiece(0, humanSrcColor, humanDstCoord)

	for {

		cpuSrcColor = boardColors[humanDstCoord.i][humanDstCoord.j]
		fmt.Println(cpuSrcColor)
		cpuDstCoord = state.findBestMoveCoord(1, cpuSrcColor, depth)
		state.MovePiece(1, cpuSrcColor, cpuDstCoord)

		if state.IsWinning(1) {
			fmt.Println("CPU won")
			os.Exit(0)
		}

		state.PrintBoard()

		humanSrcColor = boardColors[cpuDstCoord.i][cpuDstCoord.j]

		fmt.Println("CPU has played on " + humanSrcColor)

		fmt.Print("Enter your move (dst coords as 2 numbers): ")
		text, _ = reader.ReadString('\n')
		if text[0] == 'q' {
			os.Exit(0)
		}
		a = r.FindAllString(text, -1)

		humanDstCoord = ToCoord(a)
		state.MovePiece(0, humanSrcColor, humanDstCoord)

		if state.IsWinning(0) {
			fmt.Println("You won!")
			os.Exit(0)
		}
	}

}
