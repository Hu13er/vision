package vision

import (
	"fmt"
	"math"
	"sort"
)

func Hough(m Matrix) Matrix {
	r, c := m.Dims()
	diam := math.Hypot(float64(r), float64(c))
	fmt.Println("Start.. diam:", diam)

	eq := func(r, teta, x, y int, t float64) bool {
		rad := float64(teta) * math.Pi / 180.0
		right := float64(x)*math.Cos(rad) + float64(y)*math.Sin(rad)
		return math.Abs(float64(r)-right) <= t
	}

	type P struct {
		x int
		y int
	}
	cand := make(map[P]struct{})

	Iterate(m, func(i, j int, v float64) {
		if v > 0xff/2 {
			cand[P{i, j}] = struct{}{}
		}
	})
	fmt.Println("Pointset Done", len(cand))

	type L struct {
		r int
		t int
	}
	lines := make(map[L]int)

	for r := 0; r <= int(diam); r++ {
		if r%100 == 0 {
			fmt.Println("Scan R", r)
		}
		for teta := 0; teta < 180; teta++ {
			for p, _ := range cand {
				if eq(r, teta, p.x, p.y, 0.5) {
					l := L{r, teta}
					lines[l]++
					if lines[l] == 50 {
						fmt.Println("Line Found", r, teta)
					}
				}
			}
		}
	}

	sortedLines := make([]L, 0)
	for l := range lines {
		sortedLines = append(sortedLines, l)
	}
	sort.Slice(sortedLines, func(i, j int) bool {
		return lines[sortedLines[i]] >= lines[sortedLines[j]]
	})

	sortedLines = sortedLines[:100]

	out := NewMatrix(r, c)
	Iterate(out, func(i, j int, v float64) {
		for _, line := range sortedLines {
			if eq(line.r, line.t, i, j, 1) {
				if lines[line] < 2 {
					continue
				}

				v += float64(lines[line]) * 0xff / float64(lines[sortedLines[0]])
				if v >= 0xff {
					v = 0xff
				}
				out.Set(i, j, v)
			}
		}
	})
	fmt.Println("Plot done")

	return out
}
