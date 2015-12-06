package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

type colorPair struct {
	fg color.Attribute
	bg color.Attribute
}

var colors map[string]colorPair = map[string]colorPair{
	"orange": colorPair{color.FgHiRed, color.BgHiRed},
	"blue":   colorPair{color.FgCyan, color.BgCyan},
	"purple": colorPair{color.FgBlue, color.BgBlue},
	"pink":   colorPair{color.FgHiMagenta, color.BgHiMagenta},
	"yellow": colorPair{color.FgHiYellow, color.BgHiYellow},
	"red":    colorPair{color.FgRed, color.BgRed},
	"green":  colorPair{color.FgGreen, color.BgGreen},
	"brown":  colorPair{color.FgHiBlack, color.BgHiBlack},
}

func (state *state) printBoard() {
	board := state.board
	fmt.Println()
	fmt.Print("    ")
	for _, c := range "abcdefgh" {
		fmt.Print(fmt.Sprintf("%c    ", c))
	}
	fmt.Println()
	for i := 0; i < N; i++ {
		for k := 0; k < 3; k++ {
			if k == 1 {
				fmt.Print(fmt.Sprintf(" %d", N-i))
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
					color.Set(colors[piece.Color].fg)
					if piece.Player == humanPlayer {
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
				fmt.Print(N - i)
			}
			fmt.Println()
		}
	}
	fmt.Print("    ")
	for _, c := range "abcdefgh" {
		fmt.Print(fmt.Sprintf("%c    ", c))
	}
	fmt.Println()
	fmt.Println()
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

	// boardJ, _ := json.Marshal(board)
	// fmt.Println(string(boardJ))
	// os.Exit(0)

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

	var humanSrcCoord, humanDstCoord, cpuDstCoord coord
	var humanSrcColor, cpuSrcColor string
	var text string
	var a [][]string
	var reader *bufio.Reader
	var r *regexp.Regexp
	var err error

	for {

		reader = bufio.NewReader(os.Stdin)
		fmt.Print("Enter your move (e.g. a1 d4): ")
		text, _ = reader.ReadString('\n')
		text = strings.ToLower(text)
		if strings.HasPrefix(text, "q") {
			os.Exit(0)
		}

		r, _ = regexp.Compile(`.*([a-h]).*([1-8]).*([a-h]).*([1-8]).*`)
		a = r.FindAllStringSubmatch(text, -1)

		if len(a) == 0 {
			fmt.Println("Invalid move!")
			continue
		}

		humanSrcCoord, err = toCoord(a[0][1:3])
		if err != nil {
			fmt.Println("Invalid move!")
			continue
		}
		humanSrcColor = boardColors[humanSrcCoord.i][humanSrcCoord.j]

		humanDstCoord, err = toCoord(a[0][3:])
		if err != nil {
			fmt.Println("Invalid move!")
			continue
		}

		if !state.isLegalMove(humanPlayer, humanSrcColor, humanDstCoord) {
			fmt.Println("Illegal move!")
			continue
		}

		break
	}

	state.movePiece(humanPlayer, humanSrcColor, humanDstCoord)
	cpuSrcColor = boardColors[humanDstCoord.i][humanDstCoord.j]

	for {
		// --> cpuSrcColor must be defined at this point

		cpuDstCoord = state.findBestMoveCoord(cpuPlayer, cpuSrcColor, depth)

		state.movePiece(cpuPlayer, cpuSrcColor, cpuDstCoord)

		if state.isWinning(cpuPlayer) {
			state.printBoard()
			fmt.Println("CPU won")
			os.Exit(0)
		}

		state.printBoard()

		humanSrcColor = boardColors[cpuDstCoord.i][cpuDstCoord.j]

		fmt.Println("CPU has played on " + humanSrcColor)

		// Replay right away is human is blocked
		if state.isBlocked(humanPlayer, humanSrcColor) {
			fmt.Println("You are blocked, CPU will play again")
			// This can be considered as a zero-length move by the human
			// player, so the CPU src color will be the color of the cell
			// on which the target human piece is on.
			humanDstCoord = state.playerPieceCoords[humanPlayer][humanSrcColor]
			cpuSrcColor = boardColors[humanDstCoord.i][humanDstCoord.j]
			continue
		}

		for {
			// --> humanSrcColor must be defined at this point

			fmt.Print("Enter your move (e.g. d3): ")
			text, _ = reader.ReadString('\n')
			text = strings.ToLower(text)
			if strings.HasPrefix(text, "q") {
				os.Exit(0)
			}

			r, _ = regexp.Compile(`.*([a-h]).*([1-8]).*`)
			a = r.FindAllStringSubmatch(text, -1)

			if len(a) == 0 {
				fmt.Println("Invalid move!")
				continue
			}

			humanDstCoord, err = toCoord(a[0][1:])
			if err != nil {
				fmt.Println("Invalid move!")
				continue
			}

			if !state.isLegalMove(humanPlayer, humanSrcColor, humanDstCoord) {
				fmt.Println("Illegal move!")
				continue
			}

			state.movePiece(humanPlayer, humanSrcColor, humanDstCoord)
			cpuSrcColor = boardColors[humanDstCoord.i][humanDstCoord.j]

			if state.isBlocked(cpuPlayer, cpuSrcColor) {
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

		if state.isWinning(humanPlayer) {
			state.printBoard()
			fmt.Println("You won!")
			os.Exit(0)
		}
	}

}
