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
		fmt.Println("Error registering server at port: 8030 -> ", err)
		return
	} else {
		fmt.Println("Connection successful")
	}

	listener, _ := net.Listen("tcp", "127.0.0.1:"+*pAddr)

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Listener error at port: 8030 -> ", err)
		}
	}(listener)

	fmt.Println("Server listening on port", listener.Addr())

	// make the server start accepting communication
	rpc.Accept(listener)
}
