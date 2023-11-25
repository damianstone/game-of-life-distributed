package gol

import (
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

func calculateAliveCells(world [][]uint8) []util.Cell {
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

func compareAndSendCellFlippedEvents(c distributorChannels, turn int, currentWorld, updatedWorld [][]uint8) {
	for i := range currentWorld {
		for j := range currentWorld[i] {
			if currentWorld[i][j] != updatedWorld[i][j] {
				c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: j, Y: i}}
			}
		}
	}
}
