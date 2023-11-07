package gol

import (
	"strconv"
	"uk.ac.bris.cs/gameoflife/util"
)

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

func countLiveNeighbours(i int, j int, world [][]uint8) int {
	// 8 potential neighbours per cell
	neighborOffsets := [8][2]int{
		{-1, -1},
		{-1, 0},
		{-1, 1},
		{0, -1},
		{0, 1},
		{1, -1},
		{1, 0},
		{1, 1},
	}

	liveNeighbours := 0

	for _, offset := range neighborOffsets {

		// calculate neighbour positions
		height := len(world)
		width := len(world[0])

		ni := (i + offset[0] + height) % height
		nj := (j + offset[1] + width) % width

		if world[ni][nj] == 255 {
			liveNeighbours++
		}
	}

	return liveNeighbours
}

func calculateNextState(p Params, currentWorld [][]uint8) [][]uint8 {

	nextWorld := make([][]uint8, len(currentWorld))

	for i := range currentWorld {
		nextWorld[i] = make([]uint8, len(currentWorld[i]))
		copy(nextWorld[i], currentWorld[i])
	}

	for i := range currentWorld {
		// iterate through the rows of the image
		for j := range currentWorld[i] {
			count := countLiveNeighbours(i, j, currentWorld)

			// any live cell with fewer than two live neighbours dies
			// any live cell with more than three live neighbours dies
			if currentWorld[i][j] == 255 && (count < 2 || count > 3) {
				nextWorld[i][j] = 0

				// any dead cell with exactly three live neighbours becomes alive
			} else if currentWorld[i][j] == 0 && count == 3 {
				nextWorld[i][j] = 255
			}

			// any live cell with two or three live neighbours is unaffected
			// so just don't do anything
		}
	}

	return nextWorld
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
