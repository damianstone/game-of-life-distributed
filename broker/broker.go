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
	"uk.ac.bris.cs/gameoflife/broker/utils"
	"uk.ac.bris.cs/gameoflife/schema"
)

var world [][]uint8
var mutex sync.Mutex
var turn int
var shutdownFlag bool
var pauseFlag bool

type Broker struct{}

func callNode(add string, client *rpc.Client, height int, nodeWorld [][]uint8, out chan [][]uint8) {
	defer client.Close()

	request := schema.Request{
		World: nodeWorld,
	}
	response := new(schema.Response)

	fmt.Println("Calling node in address: ", add)

	err := client.Call(schema.HandleWorker, request, response)

	if err != nil {
		fmt.Println("Error calling node: ", err)
	}

	out <- response.World[1 : height+1]
}

func (b *Broker) HandleBroker(request schema.Request, response *schema.Response) (err error) {
	world = request.World

	nodeAddresses := []string{
		"127.0.0.1:8050",
		"127.0.0.1:8051",
		"127.0.0.1:8052",
		"127.0.0.1:8053",
	}

	workerHeight := len(world) / request.Params.Threads
	remaining := len(world) % request.Params.Threads

	//  channel to collect worker results
	channelSlice := make([]chan [][]uint8, request.Params.Threads)

	for turn = 0; turn < request.Params.Turns; {

		updatedWorld := make([][]uint8, 0)

		for i := 0; i < request.Params.Threads; i++ {

			channelSlice[i] = make(chan [][]uint8)

			if (remaining > 0) && ((i + 1) == request.Params.Threads) {
				startY := i * workerHeight
				endY := ((i + 1) * workerHeight) + remaining
				height := endY - startY
				nodeWorld := utils.GetImagePart(request.Params, startY, endY, world)
				nodeAdd := nodeAddresses[i]
				client, _ := rpc.Dial("tcp", nodeAdd)
				go callNode(nodeAdd, client, height, nodeWorld, channelSlice[i])

			} else {
				startY := i * workerHeight
				endY := (i + 1) * workerHeight
				height := endY - startY
				nodeWorld := utils.GetImagePart(request.Params, startY, endY, world)
				nodeAdd := nodeAddresses[i]
				client, _ := rpc.Dial("tcp", nodeAdd)
				go callNode(nodeAdd, client, height, nodeWorld, channelSlice[i])
			}

		}

		for i := 0; i < request.Params.Threads; i++ {
			receivedData := <-channelSlice[i]
			updatedWorld = append(updatedWorld, receivedData...)
		}

		mutex.Lock()
		world = updatedWorld
		turn++
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
		// TODO: pause the execution of the broker and send the current state to the client
		// TODO: the client should be able to resume the execution of the broker
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
	// register the broker
	err := rpc.Register(&Broker{})

	if err != nil {
		fmt.Println("Error registering broker: ", err)
		return
	}

	listener, _ := net.Listen("tcp", ":"+*pAddr)

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)

	// goroutine to handle broker shutdown
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			mutex.Lock()
			if shutdownFlag {
				mutex.Unlock()
				fmt.Println("Shutting down the broker...")
				os.Exit(0)
			}
			mutex.Unlock()
		}
	}()

	// make the broker start accepting communication
	rpc.Accept(listener)
}