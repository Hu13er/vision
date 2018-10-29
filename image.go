package vision

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
)

func LoadGrayImage(path string) (Matrix, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	decoded, err := jpeg.Decode(f)
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

func SaveGrayImage(m Matrix, path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	r, c := m.Dims()
	gray := image.NewGray(image.Rect(0, 0, r, c))
	Iterate(m, func(i, j int, value float64) {
		gray.Set(i, j, color.Gray{uint8(value)})
	})

	return jpeg.Encode(f, gray, nil)
}
