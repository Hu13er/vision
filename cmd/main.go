package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Hu13er/vision"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: vision <algorithm> <input> <output>")
		os.Exit(1)
	}
	var (
		algo   = os.Args[1]
		input  = os.Args[2]
		output = os.Args[3]
	)

	m, err := vision.LoadGrayImage(input)
	if err != nil {
		panic(err)
	}
	fmt.Println(m.Dims())

	var m2 vision.Matrix
	switch strings.ToLower(algo) {
	case "sobel":
		m2 = vision.Sobel(m)
	case "canny":
		m2 = vision.Canny(m)
	case "blur":
		m2 = vision.Blur(m, 13, 1)
	default:
		fmt.Println("not supported")
		os.Exit(1)
	}

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
