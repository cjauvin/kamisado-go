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
	player int
	color  string
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

func toCoord(a []string) coord {
	i, _ := strconv.Atoi(a[0])
	j, _ := strconv.Atoi(a[1])
	return coord{i, j}
}

func (state *state) movePiece(player int, color string, dst coord) {
	src := state.playerPieceCoords[player][color]
	piece := state.board[src.i][src.j]
	state.board[dst.i][dst.j] = piece
	state.board[src.i][src.j] = nil
	state.playerPieceCoords[player][color] = coord{dst.i, dst.j}
}

func (state *state) printBoard() {
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

func (state *state) getPossibleMoveCoords(player int, color string) []coord {
	incrs := [3]coord{coord{1, -1}, coord{1, 0}, coord{1, 1}}
	coords := []coord{}
	src := state.playerPieceCoords[player][color]
	piece := state.board[src.i][src.j]
	m := 1
	if piece.player == 0 {
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

func main() {

	if len(os.Args) != 2 {
		fmt.Println("usage: kamisado-go <depth>")
		os.Exit(0)
	}

	depth, ok := strconv.Atoi(os.Args[1])
	if ok != nil {
		fmt.Println("usage: kamisado-go <depth>")
		os.Exit(0)
	}

	board := [8][8]*piece{}
	playerPieceCoords := [2]map[string]coord{}

	// Note that the 16 Piece objects created here will be the only
	// ones used throughout the entire program.
	for player, i := range []int{7, 0} {
		playerPieceCoords[player] = make(map[string]coord)
		for j := 0; j < 8; j++ {
			color := boardColors[i][j]
			board[i][j] = &piece{player, color}
			playerPieceCoords[player][color] = coord{i, j}
		}
	}

	state := state{board, playerPieceCoords}

	/////////////////////////////////////////////////////
	// state.MovePiece(0, "brown", coord{5, 0})
	// state.MovePiece(1, "blue", coord{6, 1})
	// state.MovePiece(0, "brown", coord{3, 2})
	// //state.PrintBoard()
	// //fmt.Println(state.GetPossibleMoveCoords(1, "blue"))
	// fmt.Println(state.FindBestMoveCoord(1, "blue", depth))
	// os.Exit(0)
	/////////////////////////////////////////////////////

	fmt.Println("Welcome to Kamisado! You play White (X)")
	state.printBoard()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your move (src/dst coords as 4 numbers): ")
	text, _ := reader.ReadString('\n')
	if text[0] == 'q' {
		os.Exit(0)
	}

	r, _ := regexp.Compile(`\d+`)
	a := r.FindAllString(text, -1)

	var humanSrcCoord, humanDstCoord, cpuDstCoord coord
	var humanSrcColor, cpuSrcColor string

	humanSrcCoord = toCoord(a[:2])
	humanSrcColor = boardColors[humanSrcCoord.i][humanSrcCoord.j]

	humanDstCoord = toCoord(a[2:])
	state.movePiece(0, humanSrcColor, humanDstCoord)
	cpuSrcColor = boardColors[humanDstCoord.i][humanDstCoord.j]

	for {
		// --> cpuSrcColor must be defined at this point

		cpuDstCoord = state.findBestMoveCoord(1, cpuSrcColor, depth)

		state.movePiece(1, cpuSrcColor, cpuDstCoord)

		if state.isWinning(1) {
			state.printBoard()
			fmt.Println("CPU won")
			os.Exit(0)
		}

		state.printBoard()

		humanSrcColor = boardColors[cpuDstCoord.i][cpuDstCoord.j]

		fmt.Println("CPU has played on " + humanSrcColor)

		// Replay right away is human is blocked
		if state.isBlocked(0, humanSrcColor) {
			fmt.Println("You are blocked, CPU will play again")
			// This can be considered as a zero-length move by the human
			// player, so the CPU src color will be the color of the cell
			// on which the target human piece is on.
			humanDstCoord = state.playerPieceCoords[0][humanSrcColor]
			cpuSrcColor = boardColors[humanDstCoord.i][humanDstCoord.j]
			continue
		}

		for {
			// --> humanSrcColor must be defined at this point

			fmt.Print("Enter your move (dst coords as 2 numbers): ")
			text, _ = reader.ReadString('\n')
			if text[0] == 'q' {
				os.Exit(0)
			}
			a = r.FindAllString(text, -1)

			humanDstCoord = toCoord(a)

			if !state.isLegalMove(0, humanSrcColor, humanDstCoord) {
				fmt.Println("Illegal move!")
				continue
			}

			state.movePiece(0, humanSrcColor, humanDstCoord)
			cpuSrcColor = boardColors[humanDstCoord.i][humanDstCoord.j]

			if state.isBlocked(1, cpuSrcColor) {
				fmt.Println("CPU is blocked, you get to play again")
				// This can be considered as a zero-length move by the CPU player,
				// so the human src color will be the color of the cell on which
				// the target CPU piece is on.
				cpuDstCoord = state.playerPieceCoords[1][cpuSrcColor]
				humanSrcColor = boardColors[cpuDstCoord.i][cpuDstCoord.j]
				continue
			}

			break // let CPU play
		}

		if state.isWinning(0) {
			state.printBoard()
			fmt.Println("You won!")
			os.Exit(0)
		}
	}

}
