package schema

var BrokerHandler = "Broker.HandleBroker"
var GetCurrentState = "Broker.GetCurrentState"
var HandleKey = "Broker.HandleKey"
var HandleWorker = "Worker.HandleNextState"

// Structured data types for communication between the client and the broker

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type Request struct {
	Message string
	World   [][]uint8
	Params  Params
}

type Response struct {
	Message string
	Status  string
	World   [][]uint8
}

type BlankRequest struct{}

type CurrentStateResponse struct {
	AliveCellsCount int
	CurrentWorld    [][]uint8
	Turn            int
}

type KeyRequest struct {
	Key string
}
