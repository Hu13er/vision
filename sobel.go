package vision

import "math"

func Sobel(m Matrix) Matrix {
	m1, m2 := HSobel(m), VSobel(m)
	Iterate(m1, func(i, j int, value float64) {
		m1.Set(i, j,
			math.Sqrt(math.Pow(m1.At(i, j), 2)+math.Pow(m2.At(i, j), 2)))
	})
	return m1
}

func HSobel(m Matrix) Matrix {
	kernel := NewMatrix(3, 3)
	Fill(kernel, [][]float64{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	})

	return Convolution(m, kernel)
}

func VSobel(m Matrix) Matrix {
	kernel := NewMatrix(3, 3)
	Fill(kernel, [][]float64{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1}})

	return Convolution(m, kernel)
}
