package schema

var HandleWorker = "GameOfLifeOperations.Worker"

// Structured data types for communication between the client and the server

type Response struct {
	Status string
	World  [][]uint8
}

type Request struct {
	Message      string
	InitialWorld [][]uint8
	Turns        int
	ImageWidth   int
	ImageHeight  int
}
