package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/schema"
	"uk.ac.bris.cs/gameoflife/server/utils"
)

var world [][]uint8
var mutex sync.Mutex
var turn int
var shutdownFlag bool
var pauseFlag bool

//var turnSignalChannel = make(chan schema.TurnSignal)

type Broker struct{}

// HandleBroker is the function that will be called by the client
// This function implement the logic of the game of life
func (b *Broker) HandleBroker(request schema.Request, response *schema.Response) (err error) {
	world = request.World
	for turn = 0; turn < request.Params.Turns; {
		mutex.Lock()
		//oldWorld := world
		world = utils.CalculateNextState(world)
		turn++
		//turnSignalChannel <- schema.TurnSignal{Turn: turn, CurrentWorld: world, OldWorld: oldWorld}
		mutex.Unlock()

		// check for pause flag and wait if set
		for pauseFlag {
			time.Sleep(100 * time.Millisecond)
		}
	}
	response.Status = "OK"
	response.World = world
	return err
}

// GetTurnSignal is a method to get the current turn signal
//func (b *Broker) GetTurnSignal(request schema.BlankRequest, response *schema.TurnSignal) (err error) {
//	*response = <-turnSignalChannel
//	return err
//}

func (b *Broker) GetCurrentState(request schema.Request, response *schema.CurrentStateResponse) (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	response.CurrentWorld = world
	response.AliveCellsCount = utils.CountAliveCells(world)
	response.Turn = turn
	return err
}

func (b *Broker) HandleKey(request schema.KeyRequest, response *schema.CurrentStateResponse) (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	switch string(request.Key) {
	case "q":
		world = [][]uint8{}
		turn = 0
	case "k":
		shutdownFlag = true
	case "p":
		// TODO: pause the execution of the server and send the current state to the client
		// TODO: the client should be able to resume the execution of the server
		pauseFlag = !pauseFlag
		*response = schema.CurrentStateResponse{
			CurrentWorld: world,
			Turn:         turn,
		}
	}
	return err
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	// register the server
	err := rpc.Register(&Broker{})

	if err != nil {
		fmt.Println("Error registering server: ", err)
		return
	}

	listener, _ := net.Listen("tcp", ":"+*pAddr)

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)

	// goroutine to handle server shutdown
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			mutex.Lock()
			if shutdownFlag {
				mutex.Unlock()
				fmt.Println("Shutting down the server...")
				os.Exit(0)
			}
			mutex.Unlock()
		}
	}()

	// make the server start accepting communication
	rpc.Accept(listener)
}
