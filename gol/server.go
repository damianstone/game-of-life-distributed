package gol

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/gol/schema"
)

type GameOfLifeOperations struct{}

func (s *GameOfLifeOperations) Worker(request schema.Request, response schema.Response) (err error) {
	// number of iterations in turns

	// use the calculate next state function
	return
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
