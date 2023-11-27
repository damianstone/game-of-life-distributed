package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"uk.ac.bris.cs/gameoflife/schema"
)

type Worker struct{}

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

func (w *Worker) HandleNextState(request schema.Request, response *schema.Response) (err error) {
	currentWorld := request.World

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

	response.Status = "OK"
	response.World = nextWorld
	return err
}

func (w *Worker) CloseNode(request schema.BlankRequest, response *schema.Response) (err error) {
	fmt.Println("Closing node...")
	os.Exit(0)
	return err
}

func main() {
	// RPC broker
	pAddr := flag.String("port", "127.0.0.1:8050", "IP and port to listen on")
	flag.Parse()

	// Register Worker as an RPC broker
	rpc.Register(&Worker{})

	listener, lisError := net.Listen("tcp", *pAddr)
	if lisError != nil {
		fmt.Println(lisError)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on " + *pAddr)

	for {
		conn, acceptError := listener.Accept()
		if acceptError != nil {
			fmt.Println("Error accepting connection: ", acceptError)
			continue
		}

		go rpc.ServeConn(conn)
	}
}
