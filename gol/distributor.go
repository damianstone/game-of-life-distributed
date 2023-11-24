package gol

import (
	"fmt"
	"net/rpc"
	"os"
	"strconv"
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

func gameOfLifeController(p Params, c distributorChannels, initialWorld [][]uint8) [][]uint8 {
	ticker := time.NewTicker(2 * time.Second)
	client, _ := rpc.Dial("tcp", "127.0.0.1:8030")
	defer client.Close()
	request := schema.Request{
		World: initialWorld,
		Params: schema.Params{
			Turns:       p.Turns,
			Threads:     p.Threads,
			ImageWidth:  p.ImageWidth,
			ImageHeight: p.ImageHeight,
		},
	}
	response := new(schema.Response)
	done := client.Go(schema.BrokerHandler, request, response, nil)

	for {
		select {
		case <-done.Done:
			return response.World
		case <-ticker.C:
			request := schema.BlakRequest{}
			response := new(schema.CurrentStateResponse)
			err := client.Call(schema.GetCurrentState, request, response)
			if err != nil {
				fmt.Println("Error GetCurrentState -> ", err)
				os.Exit(1)
			}
			c.events <- AliveCellsCount{CompletedTurns: response.Turn, CellsCount: response.AliveCellsCount}
		case key := <-c.ioKeyPress:
			handleSdlEvents(p, turn, c, string(key), currentWorld)
		}
	}

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
