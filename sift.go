package vision

import (
	"image"
	"image/color"
	"math"

	"github.com/coraldane/resize"
)

type Sift struct {
	Image       image.Image
	ScaleCount  int
	OctaveCount int
	InitSigma   float64
	K           float64
	Radius      float64
	Thersh      float64
}

func (s *Sift) KeyPoints() []KeyPoint {
	scaledImages := s.createImageScales()

	scaledMatrixes := make([]Matrix, len(scaledImages))
	for i := range scaledMatrixes {
		scaledMatrixes[i] = imageToMatrix(scaledImages[i])
	}

	result := make([]KeyPoint, 0)
	for i := range scaledMatrixes {
		o := &Octave{
			Main:   scaledMatrixes[i],
			Sigma:  s.InitSigma,
			K:      s.K,
			Radius: s.Radius,
		}
		can := o.Init(s.OctaveCount).Candidates()
		result = append(result, FilterCandidates(can, s.Thersh)...)
	}

	return result
}

func (s *Sift) createImageScales() []image.Image {
	outp := make([]image.Image, s.ScaleCount)
	iterImg := s.Image
	for i := range outp {
		b := iterImg.Bounds()
		scaled := resize.Resize(b.Dx()/2, b.Dy()/2, iterImg, resize.Bilinear)
		iterImg = scaled
		outp[i] = scaled
	}
	return outp
}

type DoG struct {
	Matrix Matrix
	G      Gradian
}

type Octave struct {
	Main   Matrix
	GMain  Gradian
	Sigma  float64
	K      float64
	Radius float64

	Fadeing []Matrix
	DoGs    []DoG
}

func (o *Octave) Init(count int) *Octave {
	o.GMain = G(o.Main)
	o.Fade(count)
	o.DoG()
	return o
}

func (o *Octave) Fade(count int) {
	o.Fadeing = make([]Matrix, count)
	o.Fadeing[0] = o.Main
	for i := 1; i < count; i++ {
		o.Fadeing[i] = Blur(o.Fadeing[i-1], o.Radius, o.Sigma)
	}
}

func (o *Octave) DoG() {
	o.DoGs = make([]DoG, len(o.Fadeing)-1)
	for i := 1; i < len(o.Fadeing); i++ {
		f, pf := o.Fadeing[i], o.Fadeing[i-1]

		dog := NewMatrix(f.Dims())
		Iterate(dog, func(i, j int, v float64) {
			dog.Set(i, j, f.At(i, j)-pf.At(i, j))
		})
		o.DoGs[i-1] = DoG{
			Matrix: dog,
			G:      G(dog),
		}
	}
}

func (o *Octave) Candidates() []KeyPoint {
	max := func(a, b float64) float64 {
		if a > b {
			return a
		}
		return b
	}
	min := func(a, b float64) float64 {
		if a < b {
			return a
		}
		return b
	}

	kp := make([]KeyPoint, 0)
	for i, dog := range o.DoGs {
		if i == 0 || i == len(o.DoGs)-1 {
			continue
		}

		bx, by := dog.Matrix.Dims()
		IterateSlice(dog.Matrix, 1, 1, bx-2, by-2, func(i, j int, v float64) {

			mini, maxi := 1e6, -1e6

			doga := o.DoGs[i-1].Matrix
			IterateSlice(doga, i-1, j-1, i+1, j+1, func(i, j int, v float64) {
				mini = min(mini, v)
				maxi = max(maxi, v)
			})

			IterateSlice(dog.Matrix, i-1, j-1, i+1, j+1, func(i, j int, v float64) {
				mini = min(mini, v)
				maxi = max(maxi, v)
			})

			dogc := o.DoGs[i+1].Matrix
			IterateSlice(dogc, i-1, j-1, i+1, j+1, func(i, j int, v float64) {
				mini = min(mini, v)
				maxi = max(maxi, v)
			})

			if v <= mini || v >= maxi {
				kp = append(kp, KeyPoint{
					Octave: o,
					DoG:    &dog,
					X:      i,
					Y:      j,
				})
			}
		})
	}
	return kp
}

func FilterCandidates(kps []KeyPoint, th float64) []KeyPoint {
	outp := make([]KeyPoint, 0)
	for _, kp := range kps {
		g := kp.DoG.G.G.At(kp.X, kp.Y)
		if g < th {
			continue
		}
		outp = append(outp, kp)
	}
	return outp
}

type KeyPoint struct {
	Octave  *Octave
	DoG     *DoG
	X, Y    int
	Feature []float64
}

func (kp *KeyPoint) Calculate() {
	// TODO
}

type Gradian struct {
	G      Matrix
	GX, GY Matrix
	Teta   Matrix
}

func G(m Matrix) Gradian {
	gx := func() Matrix {
		kernel := NewMatrix(3, 3)
		Fill(kernel, [][]float64{
			{-1, -2, -1},
			{0, 0, 0},
			{1, 2, 1},
		})
		return Convolution(m, kernel)
	}
	gy := func() Matrix {
		kernel := NewMatrix(3, 3)
		Fill(kernel, [][]float64{
			{-1, 0, 1},
			{-2, 0, 2},
			{-1, 0, 1}})

		return Convolution(m, kernel)
	}

	g := Gradian{}

	g.GX = gx()
	g.GY = gy()

	g.G = NewMatrix(m.Dims())
	Iterate(g.G, func(i, j int, _ float64) {
		dx := g.GX.At(i, j)
		dy := g.GY.At(i, j)
		g.G.Set(i, j, math.Hypot(dx, dy))
	})

	g.Teta = NewMatrix(m.Dims())
	Iterate(g.Teta, func(i, j int, _ float64) {
		dx := g.GX.At(i, j)
		dy := g.GY.At(i, j)
		t := math.Atan2(dx, dy)
		t = t / math.Pi * 90
		g.Teta.Set(i, j, t)
	})

	return g
}

func imageToMatrix(img image.Image) Matrix {
	size := img.Bounds().Size()
	m := NewMatrix(size.X, size.Y)
	for i := 0; i < size.X; i++ {
		for j := 0; j < size.Y; j++ {
			pixel := color.GrayModel.Convert(img.At(i, j))
			g, _, _, _ := pixel.RGBA()
			g >>= 8
			m.Set(i, j, float64(g))
		}
	}
	return m
}
