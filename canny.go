package vision

import "math"

func Canny(m Matrix) Matrix {
	var (
		r, c = m.Dims()

		gx   = VSobel(m)
		gy   = HSobel(m)
		g    = NewMatrix(r, c)
		teta = NewMatrix(r, c)

		out = NewMatrix(r, c)
	)
	pyth := func(a, b float64) float64 { return math.Sqrt(a*a + b*b) }

	m = Blur(m, 7, 1)

	Iterate(g, func(i, j int, _ float64) {
		g.Set(i, j, pyth(gx.At(i, j), gy.At(i, j)))
	})

	Iterate(teta, func(i, j int, _ float64) {
		t := math.Atan2(gx.At(i, j), gy.At(i, j))
		t = t / math.Pi * 90

		inrange := func(a, b float64) bool { return a <= t && t < b }
		switch {
		case inrange(0, 22.5):
			t = 0
		case inrange(22.5, 67.5):
			t = 45
		case inrange(67.5, 112.5):
			t = 90
		case inrange(112.5, 157.5):
			t = 135
		case inrange(157.5, 180):
			t = 0
		default:
			t = 0
		}

		teta.Set(i, j, t)
	})

	Iterate(out, func(i, j int, v float64) {
		var maximal bool

		// ignore edge pixels
		if r, c := out.Dims(); i <= 1 || i >= r-1 ||
			j <= 1 || j >= c-1 {
			return
		}

		switch teta.At(i, j) {
		case 0:
			maximal = (g.At(i, j) > g.At(i, j+1)) && (g.At(i, j) > g.At(i, j-1))
		case 45:
			maximal = (g.At(i, j) > g.At(i+1, j-1)) && (g.At(i, j) > g.At(i-1, j+1))
		case 90:
			maximal = (g.At(i, j) > g.At(i-1, j)) && (g.At(i, j) > g.At(i+1, j))
		case 135:
			maximal = (g.At(i, j) > g.At(i+1, j+1)) && (g.At(i, j) > g.At(i-1, j-1))
		default:
			panic("unexpected angel")
		}

		const upper = 10
		if maximal && g.At(i, j) > upper {
			out.Set(i, j, 0xff)
		} else {
			out.Set(i, j, 0x0)
		}
	})

	return out
}
