package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Hu13er/vision"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: vision <algorithm> <input> [<output>]")
		os.Exit(1)
	}
	var (
		algo   = os.Args[1]
		input  = os.Args[2]
		output string
	)
	if len(os.Args) >= 4 {
		output = os.Args[3]
	}

	m, err := vision.LoadGrayImageMatrix(input)
	if err != nil {
		panic(err)
	}
	fmt.Println(m.Dims())

	var m2 vision.Matrix
	switch strings.ToLower(algo) {
	case "blur":
		m2 = vision.Blur(m, 13, 1)
		err = vision.SaveGrayImageMatrix(m2, output)
		if err != nil {
			panic(err)
		}
	case "sobel":
		m2 = vision.Sobel(m)
		err = vision.SaveGrayImageMatrix(m2, output)
		if err != nil {
			panic(err)
		}
	case "canny":
		m2 = vision.Canny(m)
		err = vision.SaveGrayImageMatrix(m2, output)
		if err != nil {
			panic(err)
		}
	case "hough":
		m2 = vision.Hough(m)
		err = vision.SaveGrayImageMatrix(m2, output)
		if err != nil {
			panic(err)
		}
	case "sift":
		img, err := vision.LoadGrayImage(input)
		if err != nil {
			panic(err)
		}
		sift := &vision.Sift{
			Image:       img,
			InitSigma:   1,
			K:           1.6,
			Thersh:      15,
			OctaveCount: 5,
			ScaleCount:  4,
			Radius:      13,
		}
		kps := sift.KeyPoints()
		fmt.Println(len(kps), "key points found.")
	default:
		fmt.Println("not supported")
		os.Exit(1)
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
