package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/schema"
	"uk.ac.bris.cs/gameoflife/server/utils"
)

// 8:32

type GameOfLifeOperations struct{}

// Worker is the function that will be called by the client
// This function implement the logic of the game of life
func (s *GameOfLifeOperations) Worker(request schema.Request, response *schema.Response) (err error) {
	world := request.InitialWorld
	for turn := 0; turn < request.Turns; turn++ {
		world = utils.CalculateNextState(world)
	}

	response.Status = "OK"
	response.World = world
	return err
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	// register the server
	err := rpc.Register(&GameOfLifeOperations{})

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

	// make the server start accepting communication
	rpc.Accept(listener)
}
