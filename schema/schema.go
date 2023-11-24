package schema

var BrokerHandler = "Broker.BrokerHandler"

// Structured data types for communication between the client and the server

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type Response struct {
	Status string
	World  [][]uint8
}

type Request struct {
	World  [][]uint8
	Params Params
}
