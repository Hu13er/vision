package vision

func Convolution(m, kernel Matrix) Matrix {
	out := NewMatrix(m.Dims())
	r, c := m.Dims()
	kr, kc := kernel.Dims()
	midi, midj := (kr-1)/2, (kc-1)/2
	Iterate(m, func(i, j int, _ float64) {
		x1, y1 := i-midi, j-midj
		x2, y2 := i+midi, j+midj
		if x1 < 0 || y1 < 0 ||
			x2 >= r || y2 >= c {
			return
		}
		slice := Slice(m, x1, y1, x2, y2)
		s := cage(Star(slice, kernel))
		out.Set(i, j, s)
	})
	return out
}

func Star(m, kernel Matrix) float64 {
	var s float64
	kernel = Flip(kernel)
	Iterate(m, func(i, j int, value float64) {
		s += kernel.At(i, j) * m.At(i, j)
	})
	return s
}

func cage(n float64) float64 {
	if n <= 0 {
		n = 0
	} else if n >= 0xff {
		n = 0xff
	}
	return n
}
