package schema

var BrokerHandler = "Broker.HandleBroker"
var GetCurrentState = "Broker.GetCurrentState"
var HandleKey = "Broker.HandleKey"

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

type BlankRequest struct{}

type CurrentStateResponse struct {
	AliveCellsCount int
	CurrentWorld    [][]uint8
	Turn            int
}

type Request struct {
	World  [][]uint8
	Params Params
}

type KeyRequest struct {
	Key string
}
