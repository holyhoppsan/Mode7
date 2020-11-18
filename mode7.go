package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"math/rand"
)

func main() {
	fmt.Println("Mode7 begins!")

	const windowSizeX = 640
	const windowSizeY = 480
	const stride = 4

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Mode 7", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, windowSizeX, windowSizeY, sdl.WINDOW_SHOWN)

	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 640, 480)

	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	pixels := make([]byte, windowSizeX*windowSizeY*stride)

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				fmt.Println("Quit")
				running = false
				break
			}
		}

		// Render logic
		for i := range pixels {
			pixels[i] = byte(rand.Intn(256))
		}

		texture.Update(nil, pixels, windowSizeX*stride)
		window.UpdateSurface()

		renderer.Clear()
		renderer.Copy(texture, nil, nil)
		renderer.Present()
	}
}
