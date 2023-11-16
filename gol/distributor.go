package gol

import (
	"fmt"
	"net/rpc"
	"os"
	"time"
	"uk.ac.bris.cs/gameoflife/schema"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	ioKeyPress <-chan rune
}

func callWorkerEndpoint(client *rpc.Client, p Params, initialWorld [][]uint8) *schema.Response {
	// schemas
	request := schema.Request{
		Message:      "some message",
		InitialWorld: initialWorld,
		Turns:        p.Turns,
		ImageWidth:   p.ImageWidth,
		ImageHeight:  p.ImageHeight,
	}
	response := new(schema.Response)

	// request to the server
	callError := client.Call(schema.HandleWorker, request, response)

	if callError != nil {
		fmt.Println("Something when wrong in the request: callWorkerEndpoint -> ", callError)
		os.Exit(1)
	}

	return response
}

func getAliveCellsEndpoint(client *rpc.Client, p Params, initialWorld [][]uint8) int {
	return 0
}

func gameOfLifeController(p Params, c distributorChannels, initialWorld [][]uint8) [][]uint8 {
	finalWorld := createWorld(p.ImageHeight, p.ImageWidth)
	ticker := time.NewTicker(2 * time.Second)
	client, err := rpc.Dial("tcp", "127.0.0.1:8030")
	if err != nil {
		fmt.Println("Something when wrong when trying to connect to the server at port: 8030 ", err)
		os.Exit(1)
	}
	defer client.Close()
	
	select {
	case <-ticker.C:
		currentAliveCells := calculateAliveCells(p, finalWorld)
		c.events <- AliveCellsCount{p.Turns, len(currentAliveCells)}
	case key := <-c.ioKeyPress:
		handleSdlEvents(p, p.Turns, c, string(key), finalWorld)
	default:
		response := callWorkerEndpoint(client, p, initialWorld)
		fmt.Println("Response status --> : " + response.Status)
		finalWorld = response.World
	}
	return finalWorld

}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	worldSlice := createWorld(p.ImageHeight, p.ImageWidth)
	initialWorld := getImage(p, c, worldSlice)
	//send CellFlipped for all cells that are alive when the image is loaded in
	for i := range initialWorld {
		for j := range initialWorld[i] {
			if initialWorld[i][j] == 255 {
				c.events <- CellFlipped{CompletedTurns: 0, Cell: util.Cell{X: j, Y: i}}
			}
		}
	}

	// TODO: Execute all turns of the Game of Life.
	finalWorld := gameOfLifeController(p, c, initialWorld)

	// TODO: Report the final state using FinalTurnCompleteEvent.
	aliveCells := calculateAliveCells(p, finalWorld)
	c.events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: aliveCells}
	writeImage(p, c, p.Turns, finalWorld)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
