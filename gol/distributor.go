package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"os"
	"uk.ac.bris.cs/gameoflife/gol/schema"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func controller() {
	server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	flag.Parse()
	fmt.Println("Server: ", *server)

	// create a client
	client, err := rpc.Dial("tcp", *server)

	// check for errors
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// close the client
	defer func(client *rpc.Client) {
		err := client.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(client)

	// create a request
	request := schema.Request{Message: "level"}
	response := new(schema.Response)
	callError := client.Call(schema.HandleWorker, request, response)

	if callError != nil {
		fmt.Println("Something when wrong in the request", err)
		os.Exit(1)
	}
	fmt.Println("Responded: " + response.Status)
}

func gameOfLife(p Params, initialWorld [][]uint8) [][]uint8 {
	world := initialWorld
	for turn := 0; turn < p.Turns; turn++ {
		world = calculateNextState(p, world)
	}
	return world
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
	finalWorld := gameOfLife(p, initialWorld)

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
