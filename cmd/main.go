package main

import (
	"fmt"
	"os"

	"github.com/Hu13er/vision"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: vision <input> <output>")
		os.Exit(1)
	}
	var (
		input  = os.Args[1]
		output = os.Args[2]
	)

	m, err := vision.LoadGrayImage(input)
	if err != nil {
		panic(err)
	}
	fmt.Println(m.Dims())

	m2 := vision.Sobel(m)

	err = vision.SaveGrayImage(m2, output)
	if err != nil {
		panic(err)
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
