package vision

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"sort"

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
		candidates := o.Init(s.OctaveCount, s.Thersh).Candidates()
		log.Println("Found", len(candidates), "candidates")
		th := candidates.cPercentBest(20.0)
		log.Println(20, "percent best:", th)
		// thOneDim := 0.0 / 2 // Gx and Gy
		kps := candidates.FilterCandidates(10.0, 10.0)
		log.Println("Got", len(kps), "keypoints.")
		result = append(result, kps...)
	}

	return result
}

func (s *Sift) createImageScales() []image.Image {
	outp := make([]image.Image, s.ScaleCount)
	outp[0] = s.Image
	for i := 1; i < len(outp); i++ {
		b := outp[i-1].Bounds()
		outp[i] = resize.Resize(b.Dx()/2, b.Dy()/2, outp[i-1], resize.Bilinear)
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

	th float64
}

func (o *Octave) Init(count int, th float64) *Octave {
	o.th = th
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

func (o *Octave) Candidates() KeyPoints {
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

	kps := make(KeyPoints, 0)
	for idx := 1; idx < len(o.DoGs)-1; idx++ {
		var (
			doga = o.DoGs[idx-1]
			dog  = o.DoGs[idx]
			dogc = o.DoGs[idx+1]
		)

		bx, by := dog.Matrix.Dims()
		IterateSlice(dog.Matrix, 1, 1, bx-2, by-2, func(i, j int, v float64) {

			mini, maxi := 1e6, -1e6
			IterateSlice(doga.Matrix, i-1, j-1, i+1, j+1, func(_, _ int, v float64) {
				mini = min(mini, v)
				maxi = max(maxi, v)
			})

			IterateSlice(dog.Matrix, i-1, j-1, i+1, j+1, func(i2, j2 int, v float64) {
				if i2 == i && j2 == j2 {
					return
				}
				mini = min(mini, v)
				maxi = max(maxi, v)
			})

			IterateSlice(dogc.Matrix, i-1, j-1, i+1, j+1, func(_, _ int, v float64) {
				mini = min(mini, v)
				maxi = max(maxi, v)
			})

			if v <= mini || v >= maxi {
				kp := KeyPoint{
					Octave: o,
					DoG:    &dog,
					X:      i,
					Y:      j,
				}
				if kp.Threshold(o.th) {
					kps = append(kps, kp)
				}
			}
		})
	}
	return kps
}

type KeyPoints []KeyPoint

func (kps KeyPoints) Swap(i, j int) {
	kps[i], kps[j] = kps[j], kps[i]
}

func (kps KeyPoints) Len() int {
	return len(kps)
}

func (kps KeyPoints) Less(i, j int) bool {
	a := kps[i].DoG.G.G.At(kps[i].X, kps[i].Y)
	b := kps[j].DoG.G.G.At(kps[j].X, kps[j].Y)
	return a > b
}

func (kps KeyPoints) Sort() {
	sort.Sort(kps)
}

func (kps KeyPoints) cPercentBest(c float64) float64 {
	kps.Sort()
	n := int(float64(kps.Len())*(c/100.0)) - 1
	if n < 0 {
		n = 0
	}
	x, y := kps[n].X, kps[n].Y
	return kps[n].DoG.G.G.At(x, y)
}

func (kps KeyPoints) FilterCandidates(thX, thY float64) []KeyPoint {
	outp := make([]KeyPoint, 0)
	for _, kp := range kps {
		gx := kp.DoG.G.GX.At(kp.X, kp.Y)
		gy := kp.DoG.G.GY.At(kp.X, kp.Y)
		if gx < thX || gy < thY {
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

func (kp *KeyPoint) Threshold(th float64) bool {
	return kp.DoG.G.G.At(kp.X, kp.Y) >= th
}

func (kp *KeyPoint) Calculate(sigma float64) {
	g := newGaussKernel(1.6, sigma)
	maxOrient := kp.MaxOrientation(kp.X-8, kp.Y-8, kp.X+8, kp.Y+8)
	features := make([]float64, 0)

	for x := kp.X - 8; x < kp.X+8; x += 4 {
		for y := kp.Y - 8; y < kp.Y; y += 4 {
			bin := make([]float64, 8)
			mg := Slice(kp.Octave.GMain.G, x, y, x+3, y+3)
			mg = Sig(mg, g)
			mt := Slice(kp.Octave.GMain.Teta, x, y, x+3, y+3)

			Iterate(mt, func(i, j int, teta float64) {
				td := angelDesct(teta - maxOrient)
				bin[td] += mg.At(i, j)
			})

			features = append(features, bin...)
		}
	}

	kp.Feature = features
	kp.NormalizeFeatures()
}

func (kp *KeyPoint) NormalizeFeatures() {
	var sum float64
	for _, f := range kp.Feature {
		sum += f * f
	}

	h := math.Sqrt(sum)
	for i, f := range kp.Feature {
		kp.Feature[i] = f / h
	}
}

func (kp *KeyPoint) MaxOrientation(x1, y1, x2, y2 int) float64 {
	mx, or := -1e8, 0.0
	IterateSlice(kp.Octave.GMain.G, x1, y1, x2, y2,
		func(i, j int, v float64) {
			if v > mx {
				mx = v
				or = kp.Octave.GMain.Teta.At(i, j)
			}
		})
	return or
}

func (kp KeyPoint) String() string {
	dx, dy := kp.Octave.Main.Dims()
	return fmt.Sprintf("KeyPoint{X: %d, Y: %d, Scale: (%d, %d)}",
		kp.X, kp.Y,
		dx, dy)
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

func angelDesct(t float64) int {
	for i := 0; i < 8; i++ {
		a := float64(i) * 45.0
		b := a + 45.0
		if a <= t && t < b {
			return i
		}
	}
	return 0
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
