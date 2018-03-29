package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sync"
	"time"

	"github.com/esimov/gobrot/palette"
)

var (
	colorPalette    string
	colorStep       float64
	xpos, ypos      float64
	width, height   int
	imageSmoothness int
	maxIteration    int
	escapeRadius    float64
	outputFile      string
)

var waitGroup sync.WaitGroup

func init() {
	flag.Float64Var(&colorStep, "step", 6000, "Color smooth step. Value should be greater than iteration count, otherwise the value will be adjusted to the iteration count.")
	flag.IntVar(&width, "width", 1024, "Rendered image width")
	flag.IntVar(&height, "height", 768, "Rendered image height")
	flag.Float64Var(&xpos, "xpos", -0.00275, "Point position on the real axis (defined on `x` axis)")
	flag.Float64Var(&ypos, "ypos", 0.78912, "Point position on the imaginary axis (defined on `y` axis)")
	flag.Float64Var(&escapeRadius, "radius", .125689, "Escape Radius")
	flag.IntVar(&maxIteration, "iteration", 800, "Iteration count")
	flag.IntVar(&imageSmoothness, "smoothness", 8, "The rendered mandelbrot set smoothness. For a more detailded and clear image use higher numbers. For 4xAA (AA = antialiasing) use -smoothness 4")
	flag.StringVar(&colorPalette, "palette", "Hippi", "Hippi | Plan9 | AfternoonBlue | SummerBeach | Biochimist | Fiesta")
	flag.StringVar(&outputFile, "file", "mandelbrot.png", "The rendered mandelbrot image filname")
	flag.Parse()
}

func main() {
	done := make(chan struct{})
	ticker := time.NewTicker(time.Millisecond * 100)

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Print(".")
			case <-done:
				ticker.Stop()
				fmt.Printf("\n\nMandelbrot set rendered into `%s`\n", outputFile)
			}
		}
	}()

	if colorStep < float64(maxIteration) {
		colorStep = float64(maxIteration)
	}
	colors := interpolateColors(&colorPalette, colorStep)

	if len(colors) > 0 {
		fmt.Print("Rendering image...")
		render(maxIteration, colors, done)
	}
	time.Sleep(time.Second)
}

func interpolateColors(paletteCode *string, numberOfColors float64) []color.RGBA {
	var factor float64
	steps := []float64{}
	cols := []uint32{}
	interpolated := []uint32{}
	interpolatedColors := []color.RGBA{}

	for _, v := range palette.ColorPalettes {
		factor = 1.0 / numberOfColors
		switch v.Keyword {
		case *paletteCode:
			if paletteCode != nil {
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
					uintColor := uint32(r)<<24 | uint32(g)<<16 | uint32(b)<<8 | uint32(a)
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
								uintColor := cosineInterpolation(maxColor, minColor, (i-min)/(max-min))
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

					interpolatedColors = append(interpolatedColors, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
				}
			}
		}
	}

	return interpolatedColors
}

func render(maxIteration int, colors []color.RGBA, done chan struct{}) {
	width = width * imageSmoothness
	height = height * imageSmoothness
	ratio := float64(height) / float64(width)
	xmin, xmax := xpos-escapeRadius/2.0, math.Abs(xpos+escapeRadius/2.0)
	ymin, ymax := ypos-escapeRadius*ratio/2.0, math.Abs(ypos+escapeRadius*ratio/2.0)

	image := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})

	for iy := 0; iy < height; iy++ {
		waitGroup.Add(1)
		go func(iy int) {
			defer waitGroup.Done()

			for ix := 0; ix < width; ix++ {
				var x = xmin + (xmax-xmin)*float64(ix)/float64(width-1)
				var y = ymin + (ymax-ymin)*float64(iy)/float64(height-1)
				norm, it := mandelIteration(x, y, maxIteration)
				iteration := float64(maxIteration-it) + math.Log(norm)

				if int(math.Abs(iteration)) < len(colors)-1 {
					color1 := colors[int(math.Abs(iteration))]
					color2 := colors[int(math.Abs(iteration))+1]
					color := linearInterpolation(rgbaToUint(color1), rgbaToUint(color2), uint32(iteration))

					image.Set(ix, iy, uint32ToRgba(color))
				}
			}
		}(iy)
	}

	waitGroup.Wait()

	output, _ := os.Create(outputFile)
	png.Encode(output, image)

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
	return uint32(r)<<24 | uint32(g)<<16 | uint32(b)<<8 | uint32(a)
}

func uint32ToRgba(col uint32) color.RGBA {
	r := col >> 24 & 0xff
	g := col >> 16 & 0xff
	b := col >> 8 & 0xff
	a := 0xff
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}
