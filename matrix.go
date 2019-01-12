package vision

import (
	"fmt"
)

type Matrix interface {
	At(i, j int) float64
	Set(i, j int, value float64)
	Dims() (r, c int)
}

type matrixImpl [][]float64

var _ Matrix = new(matrixImpl)

func (m *matrixImpl) At(i, j int) float64         { return (*m)[i][j] }
func (m *matrixImpl) Set(i, j int, value float64) { (*m)[i][j] = value }
func (m *matrixImpl) Dims() (r int, c int)        { return len(*m), len((*m)[0]) }

func NewMatrix(r, c int) Matrix {
	m := make([][]float64, r)
	for i := range m {
		m[i] = make([]float64, c)
	}
	matrix := matrixImpl(m)
	return &matrix
}

func IterateSlice(m Matrix, x1, y1, x2, y2 int, f func(i, j int, value float64)) {
	for i := x1; i <= x2; i++ {
		for j := y1; j <= y2; j++ {
			f(i, j, m.At(i, j))
		}
	}
}

func Iterate(m Matrix, f func(i, j int, value float64)) {
	r, c := m.Dims()
	IterateSlice(m, 0, 0, r-1, c-1, f)
}

func Copy(m Matrix) Matrix {
	out := NewMatrix(m.Dims())
	Iterate(m, func(i, j int, value float64) {
		out.Set(i, j, value)
	})
	return out
}

func FlipVertical(m Matrix) Matrix {
	r, c := m.Dims()
	out := NewMatrix(r, c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			out.Set(r-i-1, j, m.At(i, j))
		}
	}
	return out
}

func FlipHorizontal(m Matrix) Matrix {
	r, c := m.Dims()
	out := NewMatrix(r, c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			out.Set(r, c-j-1, m.At(i, j))
		}
	}
	return out
}

func Flip(m Matrix) Matrix {
	// m = FlipVertical(m)
	// m = FlipHorizontal(m)
	return m
}

func Slice(m Matrix, x1, y1, x2, y2 int) Matrix {
	out := NewMatrix(x2-x1+1, y2-y1+1)
	IterateSlice(m, x1, y1, x2, y2, func(i, j int, value float64) {
		out.Set(i-x1, j-y1, value)
	})
	return out
}

func Sum(m Matrix) float64 {
	var s float64
	Iterate(m, func(i, j int, value float64) {
		s += m.At(i, j)
	})
	return s
}

func Fill(m Matrix, table [][]float64) {
	Iterate(m, func(i, j int, _ float64) {
		m.Set(i, j, table[i][j])
	})
}

func Mult(m Matrix, c float64) Matrix {
	out := NewMatrix(m.Dims())
	Iterate(m, func(i, j int, value float64) {
		out.Set(i, j, c*value)
	})
	return out
}

func Maybe(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	f()
	return
}
