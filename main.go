package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"runtime"
	"strconv"
	"sync"

	"github.com/alexflint/go-arg"
)

const (
	width  = 2560
	height = 1440
	detail = 30
)

func mandelbrot(z, c complex128) complex128 {
	new_real := real(z)*real(z) - imag(z)*imag(z) - real(c)
	new_imag := real(z)*imag(z)*2.0 + imag(c)
	return complex(new_real, new_imag)
}

func escape_check(z complex128) bool {
	return (real(z)*real(z) + imag(z)*imag(z)) > 4.0
}

func pixel_to_mandelbrot(pixel, offset, zoom float64, zoom_correction int) float64 {
	return (pixel-float64(zoom_correction))*zoom + offset
}

func render_mandelbrot(x_offset, y_offset, zoom float64, width, height int, wg *sync.WaitGroup) *image.RGBA {
	const MAX_ITER = 255 * detail
	width_chan := make(chan int)
	defer close(width_chan)
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	// dividing by two makes the zoom center on the middle of the screen
	width_correction := width / 2
	height_correction := height / 2
	var canvas_mutex sync.Mutex
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			for x := range width_chan {
				for y := 0; y < height; y++ {
					cr := pixel_to_mandelbrot(float64(x), x_offset, zoom, width_correction)
					ci := pixel_to_mandelbrot(float64(y), y_offset, zoom, height_correction)
					c := complex(cr, ci)
					z := complex(0.0, 0.0)

					num_of_iters := 0
					for ; num_of_iters <= MAX_ITER; num_of_iters++ {
						z = mandelbrot(z, c)
						escaped := escape_check(z)
						if escaped {
							break
						}
					}
					colour := uint8(num_of_iters % 255)
					canvas_mutex.Lock()
					canvas.SetRGBA(x, y, color.RGBA{R: colour, G: colour, B: colour, A: 255})
					canvas_mutex.Unlock()
				}
			}
			wg.Done()
		}()
	}
	for i := 0; i < width; i++ {
		width_chan <- i
	}
	return canvas
}

func main() {
	var args struct {
		Zoom     string `default:"0.002"`
		X_offset string
		Y_offset string
	}
	arg.MustParse(&args)
	zoom, _ := strconv.ParseFloat(args.Zoom, 64)
	x_offset, _ := strconv.ParseFloat(args.X_offset, 64)
	y_offset, _ := strconv.ParseFloat(args.Y_offset, 64)

	var wg sync.WaitGroup
	canvas := render_mandelbrot(x_offset, y_offset, zoom, width, height, &wg)

	file, err := os.Create("result.png")
	if err != nil {
		panic(err)
	}
	// ensure image has finished rendering before encoding
	wg.Wait()
	if err := png.Encode(file, canvas); err != nil {
		panic(err)
	}
}
