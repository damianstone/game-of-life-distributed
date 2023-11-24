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

func callWorkerEndpoint(client *rpc.Client, p Params, world [][]uint8) *schema.Response {
	// schemas
	request := schema.Request{
		World: world,
		Params: schema.Params{
			Turns:       p.Turns,
			Threads:     p.Threads,
			ImageWidth:  p.ImageWidth,
			ImageHeight: p.ImageHeight,
		},
	}

	response := new(schema.Response)

	// request to the server
	callError := client.Call(schema.BrokerHandler, request, response)

	if callError != nil {
		fmt.Println("Something when wrong in the request: callWorkerEndpoint -> ", callError)
		os.Exit(1)
	}

	return response
}

func gameOfLifeController(p Params, c distributorChannels, initialWorld [][]uint8) [][]uint8 {
	client, _ := rpc.Dial("tcp", "127.0.0.1:8030")
	defer client.Close()

	turn := 0
	currentWorld := initialWorld
	// ticker to send events every 2 seconds
	ticker := time.NewTicker(2 * time.Second)

	for turn < p.Turns {

		select {
		case <-ticker.C:
			currentAliveCells := calculateAliveCells(currentWorld)
			c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(currentAliveCells)}
		case key := <-c.ioKeyPress:
			handleSdlEvents(p, turn, c, string(key), currentWorld)
		default:

			response := callWorkerEndpoint(client, p, currentWorld)
			updatedWorld := response.World

			// compare which cells have changed and send CellFlipped events
			compareAndSendCellFlippedEvents(c, turn, currentWorld, updatedWorld)
			currentWorld = updatedWorld
			c.events <- TurnComplete{CompletedTurns: turn}
			turn++
		}
	}

	return currentWorld
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
	aliveCells := calculateAliveCells(finalWorld)
	c.events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: aliveCells}
	writeImage(p, c, p.Turns, finalWorld)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
