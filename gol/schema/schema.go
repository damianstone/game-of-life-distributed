package schema

var HandleWorker = "GameOfLifeOperations.Worker"

type Response struct {
	Status string
	World  [][]uint8
}

type Request struct {
	Message      string
	InitialState [][]uint8
	Turns        int
	ImageWidth   int
	ImageHeight  int
}
