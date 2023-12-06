package gol

import (
	"fmt"
	"net"
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

var (
	distributorRegistered bool
	channels              distributorChannels
	params                Params
)

type Distributor struct{}

func gameOfLifeController(p Params, c distributorChannels, initialWorld [][]uint8) [][]uint8 {
	ticker := time.NewTicker(2 * time.Second)
	client, _ := rpc.Dial("tcp", "127.0.0.1:8030")
	// client, _ := rpc.Dial("tcp", "18.133.185.254:8030")
	defer client.Close()
	request := schema.Request{
		World: initialWorld,
		Params: schema.Params{
			Turns:       p.Turns,
			Threads:     p.Threads, // just for testing purposes
			ImageWidth:  p.ImageWidth,
			ImageHeight: p.ImageHeight,
		},
	}
	response := new(schema.Response)
	done := client.Go(schema.BrokerHandler, request, response, nil)

	for {
		select {
		case <-done.Done:
			ticker.Stop()
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
				// NOTE: generate a PGM file of the current state
				request := schema.BlankRequest{}
				response := new(schema.CurrentStateResponse)
				err := client.Call(schema.GetCurrentState, request, response)
				if err != nil {
					fmt.Println("Error GetCurrentState -> ", err)
					os.Exit(1)
				}
				writeImage(p, c, response.Turn, response.CurrentWorld)
			case "q":
				// NOTE: close the client and reset the broker
				keyRequest := schema.KeyRequest{Key: "q"}
				keyResponse := new(schema.CurrentStateResponse)
				keyError := client.Call(schema.HandleKey, keyRequest, keyResponse)
				if keyError != nil {
					fmt.Println("Error HandleKey -> ", keyError)
					os.Exit(1)
				}
				writeImage(p, c, keyResponse.Turn, keyResponse.CurrentWorld)
				c.events <- StateChange{CompletedTurns: keyResponse.Turn, NewState: Quitting}
				close(c.events)

				time.Sleep(500 * time.Millisecond)
				os.Exit(0)
			case "k":
				// NOTE: get the current state
				keyRequest := schema.KeyRequest{Key: "q"}
				keyResponse := new(schema.CurrentStateResponse)
				keyError := client.Call(schema.HandleKey, keyRequest, keyResponse)
				if keyError != nil {
					fmt.Println("Error HandleKey -> ", keyError)
					os.Exit(1)
				}
				writeImage(p, c, keyResponse.Turn, keyResponse.CurrentWorld)
				c.events <- StateChange{CompletedTurns: keyResponse.Turn, NewState: Quitting}
				close(c.events)

				// NOTE: shutdown the broker and nodes
				shutDownRequest := schema.KeyRequest{Key: "k"}
				shutDownResponse := new(schema.CurrentStateResponse)
				done := client.Go(schema.HandleKey, shutDownRequest, shutDownResponse, nil)
				<-done.Done

				time.Sleep(500 * time.Millisecond)
				os.Exit(0)
			case "p":
				// NOTE: print he current turn and pause the game
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

func initGame(p Params, c distributorChannels) {
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

	finalWorld := gameOfLifeController(p, c, initialWorld)

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

func (d *Distributor) HandleFlipCells(request schema.FlipRequest, response *schema.Response) (err error) {
	oldWorld := request.OldWorld
	newWorld := request.NewWorld
	turn := request.Turn

	for i := range oldWorld {
		for j := range oldWorld[i] {
			if oldWorld[i][j] != newWorld[i][j] {
				channels.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: j, Y: i}}
			}
		}
	}

	channels.events <- TurnComplete{CompletedTurns: turn}
	return err
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	channels = c
	params = p

	if !distributorRegistered {
		err := rpc.Register(&Distributor{})
		if err != nil {
			fmt.Println("Error registering distributor: ", err)
			return
		}
		distributorRegistered = true
	}

	pAddr := "127.0.0.1:8020"
	// pAddr := "13.40.158.33:8020"
	listener, err := net.Listen("tcp", pAddr)

	if err != nil {
		fmt.Println("Error listening: ", err)
		os.Exit(1)
	}

	defer listener.Close()
	fmt.Println("Distributor running on port: ", pAddr)

	// allow the program to continue running while waiting for connections
	go rpc.Accept(listener)

	initGame(p, c)
}
