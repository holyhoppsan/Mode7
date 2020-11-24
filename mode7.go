package main

import (
	"errors"
	"fmt"
	"github.com/EngoEngine/glm"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const windowSizeX = 640
const windowSizeY = 480
const stride = 4

var cameraToBackgroundTranslation = glm.Vec2{0.0, 0.0}
var affineTransformationMatrix = glm.Mat2{1.0, 0.0, 0.0, 1.0}

func getPixelIndex(x int, y int, surface *sdl.Surface) (int, error) {
	if x < 0 || x >= int(surface.W) || y < 0 || y >= int(surface.H) {
		return -1, errors.New("getPixelIndex: out of bounds")
	}

	index := (x * int(surface.Format.BytesPerPixel)) + (y * int(surface.Pitch))
	return index, nil
}

func rasterBackground(targetPixels []byte, backgroundPixels []byte, backgroundSurface *sdl.Surface) {
	// dx = displacement vector
	// T(dx) = Translation from camera space to background space
	// p = position in background space
	// q = position in camera space
	// T(dx)p = p + dx
	// T(dx)^-1 = T(-dx)

	// T(dx)q = p

	// P * q = p
	// P = A^-1

	// q = A * T(-dx)p
	// A^-1 * q = A^-1 * A * T(-dx)p
	// A^-1 * q = I * T(-dx)p
	// P * q = T(-dx)p
	// P * q = p - dx
	// dx + P * q = p

	// Adding the rotation point
	// TBD

	for y := 0; y < windowSizeY; y++ {
		for x := 0; x < windowSizeX; x++ {
			destIndex := (x + (y * windowSizeX)) * stride

			cameraSpacePosition := glm.Vec2{float32(x), float32(y)}

			translatedCameraPosition := cameraSpacePosition.Sub(&cameraToBackgroundTranslation)
			backgroundSamplePosition := affineTransformationMatrix.Mul2x1(&translatedCameraPosition)

			srcIndex, err := getPixelIndex(int(backgroundSamplePosition.X()), int(backgroundSamplePosition.Y()), backgroundSurface)

			if err == nil {
				targetPixels[destIndex] = backgroundPixels[srcIndex]
				targetPixels[destIndex+1] = backgroundPixels[srcIndex+1]
				targetPixels[destIndex+2] = backgroundPixels[srcIndex+2]
				targetPixels[destIndex+3] = backgroundPixels[srcIndex+3]
			}
		}
	}
}

func clearRenderTarget(targetBuffer []byte) {
	for index := 0; index < len(targetBuffer); index++ {
		targetBuffer[index] = 0
	}
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

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				fmt.Println("Quit")
				running = false
				break
			case *sdl.KeyboardEvent:
				if t.Type == sdl.KEYDOWN {
					switch t.Keysym.Sym {
					case sdl.K_ESCAPE:
						running = false
						break
					case sdl.K_UP:
						cameraToBackgroundTranslation.AddWith(&glm.Vec2{0.0, -1.0})
						break
					case sdl.K_DOWN:
						cameraToBackgroundTranslation.AddWith(&glm.Vec2{0.0, 1.0})
						break
					case sdl.K_LEFT:
						cameraToBackgroundTranslation.AddWith(&glm.Vec2{-1.0, 0.0})
						break
					case sdl.K_RIGHT:
						cameraToBackgroundTranslation.AddWith(&glm.Vec2{1.0, 0.0})
						break
					case sdl.K_a:
						affineTransformationMatrix.Set(0, 0, affineTransformationMatrix.At(0, 0)-0.1)
						break
					case sdl.K_d:
						affineTransformationMatrix.Set(0, 0, affineTransformationMatrix.At(0, 0)+0.1)
						break
					case sdl.K_w:
						affineTransformationMatrix.Set(1, 1, affineTransformationMatrix.At(1, 1)+0.1)
						break
					case sdl.K_s:
						affineTransformationMatrix.Set(1, 1, affineTransformationMatrix.At(1, 1)-0.1)
						break
					}
				}
			}
		}

		// Render logic

		clearRenderTarget(targetPixels)
		rasterBackground(targetPixels, mapPixels, mapImage)

		texture.Update(nil, targetPixels, windowSizeX*stride)
		window.UpdateSurface()

		renderer.Clear()
		renderer.Copy(texture, nil, nil)
		renderer.Present()
	}
}
