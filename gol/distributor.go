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
			request := schema.BlankRequest{}
			response := new(schema.CurrentStateResponse)
			err := client.Call(schema.GetCurrentState, request, response)
			if err != nil {
				fmt.Println("Error GetCurrentState -> ", err)
				os.Exit(1)
			}
			c.events <- AliveCellsCount{CompletedTurns: response.Turn, CellsCount: response.AliveCellsCount}
		case key := <-c.ioKeyPress:
			switch string(key) {

			case "s":
				// TODO: generate a PGM file of the current state
				request := schema.BlankRequest{}
				response := new(schema.CurrentStateResponse)
				err := client.Call(schema.GetCurrentState, request, response)
				if err != nil {
					fmt.Println("Error GetCurrentState -> ", err)
					os.Exit(1)
				}
				writeImage(p, c, response.Turn, response.CurrentWorld)

			case "q":
				// TODO: close the client without closing the broker
				request := schema.BlankRequest{}
				response := new(schema.CurrentStateResponse)
				err := client.Call(schema.GetCurrentState, request, response)
				if err != nil {
					fmt.Println("Error GetCurrentState -> ", err)
					os.Exit(1)
				}
				writeImage(p, c, response.Turn, response.CurrentWorld)
				c.events <- StateChange{CompletedTurns: response.Turn, NewState: Quitting}
				close(c.events)

				// wait for broker to restart before exiting the client
				shutDownRequest := schema.KeyRequest{Key: "q"}
				shutDownResponse := new(schema.CurrentStateResponse)
				done := client.Go(schema.HandleKey, shutDownRequest, shutDownResponse, nil)
				<-done.Done

				os.Exit(0)

			case "k":
				// TODO: get the current state
				request := schema.BlankRequest{}
				response := new(schema.CurrentStateResponse)
				err := client.Call(schema.GetCurrentState, request, response)
				if err != nil {
					fmt.Println("Error GetCurrentState -> ", err)
					os.Exit(1)
				}

				// channel to signal image writing completion
				imageWriteDone := make(chan struct{})

				go func() {
					writeImage(p, c, response.Turn, response.CurrentWorld)
					close(imageWriteDone)
				}()

				// Wait for image writing to complete
				<-imageWriteDone

				// TODO: shutdown the broker and nodes
				shutDownRequest := schema.KeyRequest{Key: "k"}
				shutDownResponse := new(schema.CurrentStateResponse)
				errShutDown := client.Call(schema.HandleKey, shutDownRequest, shutDownResponse)
				if errShutDown != nil {
					fmt.Println("Error HandleKey -> ", errShutDown)
					os.Exit(1)
				}

			case "p":
				// TODO: print he current turn and pause the game
				request := schema.KeyRequest{
					Key: "p",
				}
				response := new(schema.CurrentStateResponse)
				err := client.Call(schema.HandleKey, request, response)
				if err != nil {
					fmt.Println("Error HandleKey -> ", err)
					os.Exit(1)
				}
				c.events <- StateChange{CompletedTurns: response.Turn, NewState: Paused}
				fmt.Println("Turn" + strconv.Itoa(response.Turn) + "paused")

				for {
					if <-c.ioKeyPress == 'p' {
						request := schema.KeyRequest{
							Key: "p",
						}
						response := new(schema.CurrentStateResponse)
						err := client.Call(schema.HandleKey, request, response)
						if err != nil {
							fmt.Println("Error HandleKey -> ", err)
							os.Exit(1)
						}
						c.events <- StateChange{CompletedTurns: response.Turn, NewState: Executing}
						fmt.Println("Continuing")
						break
					}
				}
			default:
				fmt.Println("Invalid key")
			}
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
