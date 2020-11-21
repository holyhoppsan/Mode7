package main

import (
	"errors"
	"fmt"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const windowSizeX = 640
const windowSizeY = 480
const stride = 4

func getPixelIndex(x int, y int, surface *sdl.Surface) (int, error) {
	if x < 0 || x >= int(surface.W) || y < 0 || y >= int(surface.H) {
		return -1, errors.New("getPixelIndex: out of bounds")
	}

	index := (x * int(surface.Format.BytesPerPixel)) + (y * int(surface.Pitch))
	return index, nil
}

func main() {
	fmt.Println("Mode7 begins!")

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Mode 7", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, windowSizeX, windowSizeY, sdl.WINDOW_SHOWN)

	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	img.Init(img.INIT_PNG)

	mapImage, err := img.Load("content/map.png")
	if err != nil {
		panic(err)
	}
	defer mapImage.Free()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 640, 480)

	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	targetPixels := make([]byte, windowSizeX*windowSizeY*stride)

	mapPixels := mapImage.Pixels()

	for y := 0; y < windowSizeY; y++ {
		for x := 0; x < windowSizeX; x++ {
			destIndex := (x + (y * windowSizeX)) * stride

			srcIndex, err := getPixelIndex(x, y, mapImage)

			if err == nil {
				targetPixels[destIndex] = mapPixels[srcIndex]
				targetPixels[destIndex+1] = mapPixels[srcIndex+1]
				targetPixels[destIndex+2] = mapPixels[srcIndex+2]
				targetPixels[destIndex+3] = mapPixels[srcIndex+3]
			}
		}
	}

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				fmt.Println("Quit")
				running = false
				break
			case *sdl.KeyboardEvent:
				if t.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}
			}
		}

		// Render logic
		texture.Update(nil, targetPixels, windowSizeX*stride)
		window.UpdateSurface()

		renderer.Clear()
		renderer.Copy(texture, nil, nil)
		renderer.Present()
	}
}
