package main

import (
	"fmt"
	"math"

	"github.com/EngoEngine/glm"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const windowSizeX = 640
const windowSizeY = 480
const stride = 4

var upVector = glm.Vec3{0.0, 1.0, 0.0}
var rightVector = glm.Vec3{1.0, 0.0, 0.0}

var cameraWorldPosition = glm.Vec3{0, 0, 0}
var cameraScale = glm.Vec2{1.0, 1.0}
var cameraRotation = glm.Vec3{0.0, 0.0, math.Pi / 2}

func getPixelIndex(x int, y int, surface *sdl.Surface) int {
	if x < 0 || x >= int(surface.W) || y < 0 || y >= int(surface.H) {
		return -1
	}

	index := (x * int(surface.Format.BytesPerPixel)) + (y * int(surface.Pitch))
	return index
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

	rotationMatrix := glm.Rotate2D(cameraRotation[2])
	scaleMatrix := glm.Mat2{cameraScale.X(), 0.0, 0.0, cameraScale.Y()}
	affineTransform := rotationMatrix.Mul2(&scaleMatrix)
	invertedAffineTransformationMatrix := affineTransform.Inverse()

	for y := 0; y < windowSizeY; y++ {
		for x := 0; x < windowSizeX; x++ {
			destIndex := (x + (y * windowSizeX)) * stride

			// P * (q- q0) + p0 = p
			cameraSpacePosition := glm.Vec2{float32(x), float32(y)}
			cameraCenterOffset := glm.Vec2{float32(windowSizeX / 2), float32(windowSizeY / 2)}

			samplePostionCameraSpace := cameraSpacePosition.Sub(&cameraCenterOffset)

			transformedCameraSpacePosition := invertedAffineTransformationMatrix.Mul2x1(&samplePostionCameraSpace)

			backgroundSamplePosition := glm.Vec2{cameraWorldPosition.X(), cameraWorldPosition.Y()}
			backgroundSamplePosition.AddWith(&transformedCameraSpacePosition)

			srcIndex := getPixelIndex(int(backgroundSamplePosition.X()), int(backgroundSamplePosition.Y()), backgroundSurface)
			if srcIndex != -1 {
				targetPixels[destIndex] = backgroundPixels[srcIndex]
				targetPixels[destIndex+1] = backgroundPixels[srcIndex+1]
				targetPixels[destIndex+2] = backgroundPixels[srcIndex+2]
				targetPixels[destIndex+3] = backgroundPixels[srcIndex+3]
			}
		}
	}
}

func applyTranslationCameraSpace(translationDirection glm.Vec3) {
	var rotationMatrix = glm.Rotate3DZ(cameraRotation.Z())
	var inverseRotationMatrix = rotationMatrix.Inverse()
	var directionVector = inverseRotationMatrix.Mul3x1(&translationDirection)
	cameraWorldPosition.AddWith(&directionVector)
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

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, windowSizeX, windowSizeY)

	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	targetPixels := make([]byte, windowSizeX*windowSizeY*stride)

	mapPixels := mapImage.Pixels()

	running := true
	startTime := sdl.GetTicks()
	countedFrames := 0
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
						applyTranslationCameraSpace(glm.Vec3{0.0, -1.0, 0.0})
						break
					case sdl.K_DOWN:
						applyTranslationCameraSpace(glm.Vec3{0.0, 1.0, 0.0})
						break
					case sdl.K_LEFT:
						applyTranslationCameraSpace(glm.Vec3{-1.0, 0.0, 0.0})
						break
					case sdl.K_RIGHT:
						applyTranslationCameraSpace(glm.Vec3{1.0, 0.0, 0.0})
						break
					case sdl.K_a:
						cameraScale[0] = cameraScale[0] * 0.5
						cameraScale[1] = cameraScale[1] * 0.5
						break
					case sdl.K_d:
						cameraScale[0] = cameraScale[0] * 2.0
						cameraScale[1] = cameraScale[1] * 2.0
						break
					case sdl.K_w:
						cameraScale[1] = cameraScale[1] * 2.0
						break
					case sdl.K_s:
						cameraScale[1] = cameraScale[1] * 0.5
						break
					case sdl.K_q:
						cameraRotation[2] += 0.01 * math.Pi
						break
					case sdl.K_e:
						cameraRotation[2] -= 0.01 * math.Pi
						break
					}
				}
			}
		}

		// Render logic
		countedFrames++
		averageFramesPerSecond := float32(countedFrames) / float32(sdl.GetTicks()-startTime) * 1000.0
		window.SetTitle(fmt.Sprintf("Avg FPS: %f", averageFramesPerSecond))

		clearRenderTarget(targetPixels)
		rasterBackground(targetPixels, mapPixels, mapImage)

		texture.Update(nil, targetPixels, windowSizeX*stride)
		window.UpdateSurface()

		renderer.Clear()
		renderer.Copy(texture, nil, nil)
		renderer.Present()
	}
}
