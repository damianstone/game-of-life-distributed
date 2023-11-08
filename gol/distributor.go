package gol

import (
	"fmt"
	"net/rpc"
	"os"
	"uk.ac.bris.cs/gameoflife/schema"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
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
		fmt.Println("Something when wrong in the request", callError)
		os.Exit(1)
	}

	return response
}

func gameOfLifeController(p Params, c distributorChannels, initialWorld [][]uint8) [][]uint8 {
	
	// create a client
	client, err := rpc.Dial("tcp", "127.0.0.1:8030")

	// check for errors
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// close the client
	defer client.Close()

	response := callWorkerEndpoint(client, p, initialWorld)

	fmt.Println("Response status --> : " + response.Status)

	return response.World
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	worldSlice := make([][]uint8, p.ImageHeight)
	for i := range worldSlice {
		worldSlice[i] = make([]uint8, p.ImageWidth)
	}
	initialWorld := getImage(p, c, worldSlice)

	// TODO: Execute all turns of the Game of Life.
	finalWorld := gameOfLifeController(p, c, initialWorld)

	// TODO: Report the final state using FinalTurnCompleteEvent.
	aliveCells := calculateAliveCells(p, finalWorld)
	c.events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: aliveCells}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
