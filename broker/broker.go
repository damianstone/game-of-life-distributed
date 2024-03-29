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
var totalTurns int
var turn int
var shutdownFlag bool
var pauseFlag bool

type Broker struct {
	nodeAddresses []string
}

func callDistributor(updatedWorld [][]uint8) {
	client, nodeErr := rpc.Dial("tcp", "127.0.0.1:8020")
	// client, nodeErr := rpc.Dial("tcp", "13.40.158.33:8020")
	if nodeErr != nil {
		fmt.Println("Error when connecting to client: ", nodeErr)
		return
	}
	client.Call(schema.HandleFlipCells, schema.FlipRequest{OldWorld: world, NewWorld: updatedWorld, Turn: turn}, schema.Response{})
	client.Close()
}

func callNode(add string, client *rpc.Client, height int, nodeWorld [][]uint8, out chan [][]uint8) {
	defer client.Close()

	request := schema.Request{
		World: nodeWorld, // portion of the image that corresponds to the node
	}
	response := new(schema.Response)

	fmt.Println("Calling node in address: ", add)

	err := client.Call(schema.HandleWorker, request, response)

	if err != nil {
		fmt.Println("Error calling node: "+add, err)
	}

	out <- response.World[1 : height+1]
}

func (b *Broker) HandleBroker(request schema.Request, response *schema.Response) (err error) {
	world = request.World
	totalTurns = request.Params.Turns
	nodeAddresses := b.nodeAddresses
	// numberNodes := request.Params.Threads // for testing purposes
	numberNodes := len(nodeAddresses)

	workerHeight := len(world) / numberNodes
	remaining := len(world) % numberNodes

	//  channel to collect worker results
	channelSlice := make([]chan [][]uint8, numberNodes)

	for turn = 0; turn < totalTurns; {
		updatedWorld := make([][]uint8, 0)

		for i := 0; i < numberNodes; i++ {
			channelSlice[i] = make(chan [][]uint8)

			startY := i * workerHeight
			endY := ((i + 1) * workerHeight) + remaining
			height := endY - startY

			// get portion of the image
			nodeWorld := utils.GetImagePart(request.Params, startY, endY, world)

			// connect to the node
			nodeAdd := nodeAddresses[i]
			client, nodeErr := rpc.Dial("tcp", nodeAdd)

			if nodeErr != nil {
				fmt.Println("Error when connecting to node: "+nodeAdd+" Details : ", nodeErr)
			}

			go callNode(nodeAdd, client, height, nodeWorld, channelSlice[i])
		}

		for i := 0; i < numberNodes; i++ {
			receivedData := <-channelSlice[i]
			updatedWorld = append(updatedWorld, receivedData...)
		}

		mutex.Lock()
		callDistributor(updatedWorld)
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
	switch string(request.Key) {
	case "q":
		*response = schema.CurrentStateResponse{
			CurrentWorld: world,
			Turn:         turn,
		}
		responseChan := make(chan struct{})
		go func() {
			err := b.HandleBroker(schema.Request{}, &schema.Response{})
			if err != nil {
				fmt.Println("Error calling distributor: ", err)
			}
			responseChan <- struct{}{}
		}()
		<-responseChan
	case "k":
		for i := 0; i < len(b.nodeAddresses); i++ {
			nAddress := b.nodeAddresses[i]
			client, nodeErr := rpc.Dial("tcp", nAddress)
			if nodeErr != nil {
				fmt.Println("Error when connecting to node: "+nAddress+" Details : ", nodeErr)
			}
			done := client.Go(schema.CloseNode, schema.BlankRequest{}, schema.Response{}, nil)
			// waitingn for the node to close
			<-done.Done
			client.Close()
		}

		mutex.Lock()
		shutdownFlag = true
		mutex.Unlock()

	case "p":
		pauseFlag = !pauseFlag
		*response = schema.CurrentStateResponse{
			CurrentWorld: world,
			Turn:         turn,
		}
	}
	return err
}

func main() {
	pAddr := flag.String("port", "127.0.0.1:8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	// Locally
	broker := Broker{
		nodeAddresses: []string{
			"127.0.0.1:8050",
			"127.0.0.1:8051",
			"127.0.0.1:8052",
			"127.0.0.1:8053",
		},
	}

	// AWS
	// broker := Broker{
	// 	nodeAddresses: []string{
	// 		"18.132.63.211:8050",
	// 		"3.10.5.121:8051",
	// 		"35.177.62.130:8052",
	// 		"35.178.190.148:8053",
	// 	},
	// }

	// register the broker
	err := rpc.Register(&broker)

	if err != nil {
		fmt.Println("Error registering broker: ", err)
		return
	}

	listener, _ := net.Listen("tcp", *pAddr)

	fmt.Println("Broker running on port: ", *pAddr)

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Error closing listener: ", err)
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
