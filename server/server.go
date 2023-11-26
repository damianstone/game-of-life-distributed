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

type Broker struct{}

func getImagePart(
	p schema.Params,
	startY int,
	endY int,
	currentWorld [][]uint8) [][]uint8 {

	// calculate the dimensions of the piece of the image
	height := endY - startY
	width := p.ImageWidth

	// extra height for the first and last rows
	extraHeight := height + 2

	// create a slice to store the extended piece of the image
	nodeWorld := utils.CreateWorld(width, extraHeight)

	// extract the piece of the image from the currentWorld
	for i := 0; i < extraHeight; i++ {
		for j := 0; j < width; j++ {
			// if including the first row then add the last row of the image on the top of the piece
			if i == 0 && startY == 0 {
				nodeWorld[i][j] = currentWorld[p.ImageHeight-1][j]

				// if including the last row then add the first row of the image on the bottom of the piece
			} else if i == height+1 && startY+height == p.ImageHeight {
				nodeWorld[i][j] = currentWorld[0][j]

				// otherwise just add the piece of the image
			} else {
				a := startY + i - 1
				nodeWorld[i][j] = currentWorld[a][j]
			}
		}
	}

	return nodeWorld
}

func callNode(add string, client *rpc.Client, height int, nodeWorld [][]uint8, out chan [][]uint8) {
	defer client.Close()

	request := schema.Request{
		World: nodeWorld,
	}
	response := new(schema.Response)

	fmt.Println("Calling node in address: ", add)

	call := client.Go(schema.HandleWorker, request, response, nil)

	fmt.Println("Response from node: ", response.Status)

	<-call.Done
	fmt.Println("CALL DONE")
	out <- response.World[1 : height+1]
}

// HandleBroker is the function that will be called by the client
// This function implement the logic of the game of life
func (b *Broker) HandleBroker(request schema.Request, response *schema.Response) (err error) {
	world = request.World
	nodeAddresses := []string{
		"127.0.0.1:8050",
		"127.0.0.1:8051",
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

				nodeWorld := getImagePart(request.Params, startY, endY, world)

				nodeAdd := nodeAddresses[i]
				client, _ := rpc.Dial("tcp", nodeAdd)
				callNode(nodeAdd, client, height, nodeWorld, channelSlice[i])

			} else {
				startY := i * workerHeight
				endY := (i + 1) * workerHeight
				height := endY - startY
				nodeWorld := getImagePart(request.Params, startY, endY, world)

				nodeAdd := nodeAddresses[i]
				client, _ := rpc.Dial("tcp", nodeAdd)
				callNode(nodeAdd, client, height, nodeWorld, channelSlice[i])

			}

		}

		for i := 0; i < request.Params.Threads; i++ {
			fmt.Println("WAITING FOR CHANNEL")
			receivedData := <-channelSlice[i]
			updatedWorld = append(updatedWorld, receivedData...)
		}

		mutex.Lock()
		fmt.Println("NEXT ITERATION")
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
