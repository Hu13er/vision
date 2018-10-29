package vision

import "math"

func Blur(m Matrix, radius float64, sigma float64) Matrix {
	return Convolution(m, newGaussKernel(radius, sigma))
}

func newGaussKernel(radius float64, sigma float64) Matrix {
	length := int(math.Ceil(2*radius + 1))
	kernel := NewMatrix(length, length)
	Iterate(kernel, func(i, j int, v float64) {
		kernel.Set(i, j, gaussianFunc(float64(i)-radius, float64(j)-radius, sigma))
	})
	return kernel
}

func gaussianFunc(x, y, sigma float64) float64 {
	sigSqr := sigma * sigma
	return (1.0 / (2 * math.Pi * sigSqr)) * math.Exp(-(x*x+y*y)/(2*sigSqr))
}
