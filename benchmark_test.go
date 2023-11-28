package main

import (
	"fmt"
	"os"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

func BenchmarkGOL(b *testing.B) {

	// Disable all program output apart from benchmark results
	os.Stdout = nil

	for threads := 1; threads <= 4; threads++ {
		param := gol.Params{
			ImageWidth:  512,
			ImageHeight: 512,
			Turns:       10000,
			Threads:     threads,
		}

		testName := fmt.Sprintf(
			"%dx%dx%d-%d",
			param.ImageWidth,
			param.ImageHeight,
			param.Turns,
			threads)

		b.Run(testName, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				events := make(chan gol.Event)
				go gol.Run(param, events, nil)
				for range events {

				}
			}
		})
	}

}
