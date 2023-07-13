package brot

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sync"

	"github.com/teadove/gobrot/internal/palette"
)

type Service struct {
	WG              *sync.WaitGroup
	ColorPalette    string
	ColorStep       float64
	XPos            float64
	YPos            float64
	Width, Height   int
	ImageSmoothness int
	MaxIteration    int
	EscapeRadius    float64
	OutputFile      string
}

func (s *Service) InterpolateColors(paletteCode *string, numberOfColors float64) []color.RGBA {
	var factor float64
	var steps []float64
	var cols []uint32
	var interpolated []uint32
	var interpolatedColors []color.RGBA

	for _, v := range palette.ColorPalettes {
		factor = 1.0 / numberOfColors
		if v.Keyword != *paletteCode || paletteCode == nil {
			continue
		}
		for index, col := range v.Colors {
			if col.Step == 0.0 && index != 0 {
				stepRatio := float64(index+1) / float64(len(v.Colors))
				step := float64(int(stepRatio*100)) / 100 // truncate to 2 decimal precision
				steps = append(steps, step)
			} else {
				steps = append(steps, col.Step)
			}
			r, g, b, a := col.Color.RGBA()
			r /= 0xff
			g /= 0xff
			b /= 0xff
			a /= 0xff
			uintColor := r<<24 | g<<16 | b<<8 | a
			cols = append(cols, uintColor)
		}

		var min, max, minColor, maxColor float64
		if len(v.Colors) == len(steps) && len(v.Colors) == len(cols) {
			for i := 0.0; i <= 1; i += factor {
				for j := 0; j < len(v.Colors)-1; j++ {
					if i >= steps[j] && i < steps[j+1] {
						min = steps[j]
						max = steps[j+1]
						minColor = float64(cols[j])
						maxColor = float64(cols[j+1])
						uintColor := cosineInterpolation(
							maxColor,
							minColor,
							(i-min)/(max-min),
						)
						interpolated = append(interpolated, uint32(uintColor))
					}
				}
			}
		}

		for _, pixelValue := range interpolated {
			r := pixelValue >> 24 & 0xff
			g := pixelValue >> 16 & 0xff
			b := pixelValue >> 8 & 0xff
			a := 0xff

			interpolatedColors = append(
				interpolatedColors,
				color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)},
			)
		}
	}

	return interpolatedColors
}

func (s *Service) Render(maxIteration int, colors []color.RGBA, done chan struct{}) {
	s.Width *= s.ImageSmoothness
	s.Height *= s.ImageSmoothness
	ratio := float64(s.Height) / float64(s.Width)
	xmin, xmax := s.XPos-s.EscapeRadius/2.0, math.Abs(s.XPos+s.EscapeRadius/2.0)
	ymin, ymax := s.YPos-s.EscapeRadius*ratio/2.0, math.Abs(s.YPos+s.EscapeRadius*ratio/2.0)

	rgbaImage := image.NewRGBA(
		image.Rectangle{Min: image.Point{}, Max: image.Point{X: s.Width, Y: s.Height}},
	)

	for iy := 0; iy < s.Height; iy++ {
		s.WG.Add(1)
		go func(iy int) {
			defer s.WG.Done()

			for ix := 0; ix < s.Width; ix++ {
				x := xmin + (xmax-xmin)*float64(ix)/float64(s.Width-1)
				y := ymin + (ymax-ymin)*float64(iy)/float64(s.Width-1)
				norm, it := mandelIteration(x, y, maxIteration)
				iteration := float64(maxIteration-it) + math.Log(norm)

				if int(math.Abs(iteration)) < len(colors)-1 {
					color1 := colors[int(math.Abs(iteration))]
					color2 := colors[int(math.Abs(iteration))+1]
					compiledColor := linearInterpolation(
						rgbaToUint(color1),
						rgbaToUint(color2),
						uint32(iteration),
					)

					rgbaImage.Set(ix, iy, uint32ToRgba(compiledColor))
				}
			}
		}(iy)
	}

	s.WG.Wait()

	// TODO add err check
	output, _ := os.Create(s.OutputFile)
	// TODO add err check
	png.Encode(output, rgbaImage)

	done <- struct{}{}
}

func cosineInterpolation(c1, c2, mu float64) float64 {
	mu2 := (1 - math.Cos(mu*math.Pi)) / 2.0
	return c1*(1-mu2) + c2*mu2
}

func linearInterpolation(c1, c2, mu uint32) uint32 {
	return c1*(1-mu) + c2*mu
}

func mandelIteration(cx, cy float64, maxIter int) (float64, int) {
	var x, y, xx, yy float64 = 0.0, 0.0, 0.0, 0.0

	for i := 0; i < maxIter; i++ {
		xy := x * y
		xx = x * x
		yy = y * y
		if xx+yy > 4 {
			return xx + yy, i
		}
		x = xx - yy + cx
		y = 2*xy + cy
	}

	logZn := (x*x + y*y) / 2
	return logZn, maxIter
}

func rgbaToUint(color color.RGBA) uint32 {
	r, g, b, a := color.RGBA()
	r /= 0xff
	g /= 0xff
	b /= 0xff
	a /= 0xff
	return r<<24 | g<<16 | b<<8 | a
}

func uint32ToRgba(col uint32) color.RGBA {
	r := col >> 24 & 0xff
	g := col >> 16 & 0xff
	b := col >> 8 & 0xff
	a := 0xff
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
}
