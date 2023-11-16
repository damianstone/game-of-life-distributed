package gol

import (
	"fmt"
	"os"
	"strconv"
	"uk.ac.bris.cs/gameoflife/util"
)

func createWorld(height int, width int) [][]uint8 {
	world := make([][]uint8, height)
	for i := range world {
		world[i] = make([]uint8, width)
	}
	return world
}

func writeImage(p Params, c distributorChannels, turn int, world [][]uint8) {
	// command to write image in io.go
	c.ioCommand <- 0

	// send the filename after sent the appropriate command (to write the image)
	w := strconv.Itoa(p.ImageWidth)
	h := strconv.Itoa(p.ImageHeight)
	t := strconv.Itoa(p.Turns)
	filename := w + "x" + h + "x" + t
	c.ioFilename <- filename

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.ioOutput <- world[y][x]
		}
	}

	//c.events <- ImageOutputComplete{CompletedTurns: turn, Filename: filename + ".png"}
}

func getImage(p Params, c distributorChannels, world [][]uint8) [][]uint8 {

	// command to read image in io.go
	c.ioCommand <- 1

	// send the filename after sent the appropriate command (to read the image)
	w := strconv.Itoa(p.ImageWidth)
	h := strconv.Itoa(p.ImageHeight)
	c.ioFilename <- w + "x" + h

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			world[y][x] = <-c.ioInput
		}
	}

	// then we continue with the following TODO that is execute the turns of the game of life
	return world
}

func calculateAliveCells(p Params, world [][]uint8) []util.Cell {
	var cells []util.Cell
	for i := range world {
		for j := range world[i] {
			if world[i][j] == 255 {
				cells = append(cells, util.Cell{X: j, Y: i})
			}
		}
	}
	return cells
}

func handleSdlEvents(p Params, turn int, c distributorChannels, key string, world [][]uint8) {
	switch key {
	// terminate the program
	case "q":
		writeImage(p, c, turn, world)
		c.events <- StateChange{CompletedTurns: turn, NewState: Quitting}
		close(c.events)
		os.Exit(0)
	// generate a PGM file with the current state of the board
	case "s":
		writeImage(p, c, turn, world)
	// pause the execution
	case "p":
		c.events <- StateChange{CompletedTurns: turn, NewState: Paused}
		fmt.Println("Turn" + strconv.Itoa(p.Turns))
		for {
			if <-c.ioKeyPress == 'p' {
				c.events <- StateChange{turn, Executing}
				fmt.Println("Continuing")
				break
			}
		}
	default:
		fmt.Println("Invalid key")
	}

}
