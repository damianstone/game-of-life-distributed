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

type Broker struct{}

func worker(
	p schema.Params,
	startY int,
	endY int,
	startX int,
	endX int,
	currentWorld [][]uint8,
	out chan<- [][]uint8) {

	// calculate the dimensions of the piece of the image
	height := endY - startY
	width := endX - startX

	// extra height for the first and last rows
	extraHeight := height + 2

	// create a slice to store the extended piece of the image
	world := utils.CreateWorld(width, extraHeight)

	// extract the piece of the image from the currentWorld
	for i := 0; i < extraHeight; i++ {
		for j := 0; j < width; j++ {
			// if including the first row then add the last row of the image on the top of the piece
			if i == 0 && startY == 0 {
				world[i][j] = currentWorld[p.ImageHeight-1][startX+j]

				// if including the last row then add the first row of the image on the bottom of the piece
			} else if i == height+1 && startY+height == p.ImageHeight {
				world[i][j] = currentWorld[0][startX+j]

				// otherwise just add the piece of the image
			} else {
				a := startY + i - 1
				world[i][j] = currentWorld[a][startX+j]
			}
		}
	}

	world = utils.CalculateNextState(world)

	// Send the resulting piece of the image to the output channel
	out <- world[1 : height+1]
}

func (b *Broker) BrokerHandler(request schema.Request, response *schema.Response) (err error) {
	updatedWorld := make([][]uint8, 0)

	workerHeight := request.Params.ImageHeight / request.Params.Threads
	remaining := request.Params.ImageHeight % request.Params.Threads

	//  channel to collect worker results
	channelSlice := make([]chan [][]uint8, request.Params.Threads)

	for i := 0; i < request.Params.Threads; i++ {

		channelSlice[i] = make(chan [][]uint8)

		if (remaining > 0) && ((i + 1) == request.Params.Threads) {
			startY := i * workerHeight
			endY := ((i + 1) * workerHeight) + remaining

			go worker(
				request.Params,
				startY,
				endY,
				0,
				request.Params.ImageWidth,
				request.World,
				channelSlice[i],
			)

		} else {
			startY := i * workerHeight
			endY := (i + 1) * workerHeight

			go worker(
				request.Params,
				startY,
				endY,
				0,
				request.Params.ImageWidth,
				request.World,
				channelSlice[i],
			)
		}

	}

	for i := 0; i < request.Params.Threads; i++ {
		// receive data from worker goroutines
		receivedData := <-channelSlice[i]
		// append received data to final image data without appending extra rows
		updatedWorld = append(updatedWorld, receivedData...)

	}

	response.World = updatedWorld
	response.Status = "OK"
	return nil
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	// register the server
	err := rpc.Register(&Broker{})
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
