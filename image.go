package vision

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"
)

func LoadGrayImage(path string) (*image.Gray, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := jpeg.Decode(f)
	if err != nil {
		return nil, err
	}

	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	gray := image.NewGray(img.Bounds())
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			gray.Set(x, y, color.GrayModel.Convert(img.At(x, y)))
		}
	}

	return gray, nil
}

func LoadGrayImageMatrix(path string) (Matrix, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var decoder func(io.Reader) (image.Image, error)
	if strings.HasSuffix(path, ".png") {
		decoder = png.Decode
	} else {
		decoder = jpeg.Decode
	}

	decoded, err := decoder(f)
	if err != nil {
		return nil, err
	}

	size := decoded.Bounds().Size()
	m := NewMatrix(size.X, size.Y)
	for i := 0; i < size.X; i++ {
		for j := 0; j < size.Y; j++ {
			pixel := color.GrayModel.Convert(decoded.At(i, j))
			g, _, _, _ := pixel.RGBA()
			g >>= 8
			m.Set(i, j, float64(g))
		}
	}
	return m, nil
}

func SaveGrayImageMatrix(m Matrix, path string) error {
	var encoder func(io.Writer, image.Image) error
	if strings.HasSuffix(path, ".png") {
		encoder = png.Encode
	} else {
		encoder = func(w io.Writer, m image.Image) error { return jpeg.Encode(w, m, nil) }
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	r, c := m.Dims()
	gray := image.NewGray(image.Rect(0, 0, r, c))
	Iterate(m, func(i, j int, value float64) {
		gray.Set(i, j, color.Gray{uint8(value)})
	})

	return encoder(f, gray)
}
