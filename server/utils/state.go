package utils

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

func CalculateNextState(currentWorld [][]uint8) [][]uint8 {

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
