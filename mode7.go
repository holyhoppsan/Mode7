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

var cameraWorldPosition = glm.Vec3{256, 32, 256}
var cameraScale = glm.Vec2{1.0, 1.0}
var cameraRotation = glm.Vec3{0.0, 0.0, 0.0 /*math.Pi / 2*/}

var nearPlaneDistance float32 = 32.0
var horizonScanline = windowSizeY / 2

var frameRateCap uint32 = 1000 / 60

const cameraVelocity = 50.0
const cameraAngluarVelocity = math.Pi / 4

// RenderingMode Type for tracking the current render mode
type RenderingMode int

const (
	Affine2D = iota
	Mode7
)

var currentRenderingMode RenderingMode = Affine2D
var previousRenderingModeKeyToggleState uint8 = 0

func getPixelIndex(samplePosition glm.Vec2, surface *sdl.Surface) int {
	x := int(samplePosition.X())
	y := int(samplePosition.Y())

	if x < 0 || x >= int(surface.W) || y < 0 || y >= int(surface.H) {
		return -1
	}

	index := (x * int(surface.Format.BytesPerPixel)) + (y * int(surface.Pitch))
	return index
}

func writePixelToBuffer(samplePosition glm.Vec2, x int, y int, targetPixels []byte, backgroundPixels []byte, backgroundSurface *sdl.Surface) {
	srcIndex := getPixelIndex(samplePosition, backgroundSurface)
	if srcIndex != -1 {
		destIndex := (x + (y * windowSizeX)) * stride

		targetPixels[destIndex] = backgroundPixels[srcIndex]
		targetPixels[destIndex+1] = backgroundPixels[srcIndex+1]
		targetPixels[destIndex+2] = backgroundPixels[srcIndex+2]
		targetPixels[destIndex+3] = backgroundPixels[srcIndex+3]
	}
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

	for y := 0; y < windowSizeY; y++ {
		for x := 0; x < windowSizeX; x++ {
			// P * (q- q0) + p0 = p
			cameraSpacePosition := glm.Vec2{float32(x), float32(y)}
			cameraCenterOffset := glm.Vec2{float32(windowSizeX / 2), float32(windowSizeY / 2)}

			samplePostionCameraSpace := cameraSpacePosition.Sub(&cameraCenterOffset)

			transformedCameraSpacePosition := affineTransform.Mul2x1(&samplePostionCameraSpace)

			// Translate the camera space to background space
			backgroundSamplePosition := glm.Vec2{cameraWorldPosition.X(), cameraWorldPosition.Y()}
			backgroundSamplePosition.AddWith(&transformedCameraSpacePosition)

			writePixelToBuffer(backgroundSamplePosition, x, y, targetPixels, backgroundPixels, backgroundSurface)
		}
	}
}

func rasterBackgroundMode7Basic(targetPixels []byte, backgroundPixels []byte, backgroundSurface *sdl.Surface) {

	// Projection w = z/d such that  z / w = z / (z / d) = z * (d / z) = d

	// this means that lambda needs to be  y / ay. This means z / (y / ay) = z * (ay / ay)

	rotationMatrix := glm.Rotate2D(cameraRotation[2])
	for y := horizonScanline; y < windowSizeY; y++ {
		h := y - horizonScanline
		lambda := cameraWorldPosition.Y() / float32(h)
		scaleMatrix := glm.Mat2{lambda, 0.0, 0.0, lambda}

		cameraToBackgroundTransform := rotationMatrix.Mul2(&scaleMatrix)

		for x := 0; x < windowSizeX; x++ {

			screenSpacePosition := glm.Vec2{float32(x), float32(y)}

			cameraCenterOffset := glm.Vec2{float32(windowSizeX / 2), nearPlaneDistance + float32(h)}

			samplePostionCameraSpace := screenSpacePosition.Sub(&cameraCenterOffset)

			transformedCameraSpacePosition := cameraToBackgroundTransform.Mul2x1(&samplePostionCameraSpace)

			backgroundSamplePosition := glm.Vec2{cameraWorldPosition.X(), cameraWorldPosition.Z()}
			backgroundSamplePosition.AddWith(&transformedCameraSpacePosition)

			writePixelToBuffer(backgroundSamplePosition, x, y, targetPixels, backgroundPixels, backgroundSurface)
		}
	}
}

func applyTranslationCameraSpace(translationDirection glm.Vec3, rotation float32, deltaTime float32) {
	rotationMatrix := glm.Rotate3DZ(rotation)
	inverseRotationMatrix := rotationMatrix.Inverse()
	directionVector := inverseRotationMatrix.Mul3x1(&translationDirection)
	velocityMultiplier := cameraVelocity * deltaTime
	timeAdjustedDirectionVector := directionVector.Mul(velocityMultiplier)
	cameraWorldPosition.AddWith(&timeAdjustedDirectionVector)
}

func clearRenderTarget(targetBuffer []byte) {
	for index := 0; index < len(targetBuffer); index++ {
		targetBuffer[index] = 0
	}
}

func processSDLEvents() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			fmt.Println("Quit")
			return false
		case *sdl.KeyboardEvent:
			if t.Type == sdl.KEYDOWN {
				switch t.Keysym.Sym {
				case sdl.K_ESCAPE:
					return false
				}
			}
		}
	}
	return true
}

func processInput(deltaTime float32) {
	keyboardState := sdl.GetKeyboardState()

	directionVector := glm.Vec3{0.0, 0.0, 0.0}
	if keyboardState[sdl.SCANCODE_UP] == 1 {
		directionVector[1] += -1.0
	}

	if keyboardState[sdl.SCANCODE_DOWN] == 1 {
		directionVector[1] += 1.0
	}

	if keyboardState[sdl.SCANCODE_LEFT] == 1 {
		directionVector[0] += -1.0
	}

	if keyboardState[sdl.SCANCODE_RIGHT] == 1 {
		directionVector[0] += 1.0
	}

	if directionVector.Len() > 0.0 {
		directionVector.Normalize()

		applyTranslationCameraSpace(directionVector, cameraRotation.Z(), deltaTime)
	}

	if keyboardState[sdl.SCANCODE_A] == 1 {
		cameraScale[0] = float32(math.Max(float64(cameraScale[0]-deltaTime), 0.001))
		cameraScale[1] = float32(math.Max(float64(cameraScale[1]-deltaTime), 0.001))
	}

	if keyboardState[sdl.SCANCODE_D] == 1 {
		cameraScale[0] = float32(math.Max(float64(cameraScale[0]+deltaTime), 0.001))
		cameraScale[1] = float32(math.Max(float64(cameraScale[1]+deltaTime), 0.001))
	}

	if keyboardState[sdl.SCANCODE_W] == 1 {
		cameraScale[1] = cameraScale[1] * 2.0
	}

	if keyboardState[sdl.SCANCODE_S] == 1 {
		cameraScale[1] = cameraScale[1] * 0.5
	}

	if keyboardState[sdl.SCANCODE_Q] == 1 {
		cameraRotation[2] += cameraAngluarVelocity * deltaTime
	}

	if keyboardState[sdl.SCANCODE_E] == 1 {
		cameraRotation[2] -= cameraAngluarVelocity * deltaTime
	}

	if keyboardState[sdl.SCANCODE_R] == 1 {
		nearPlaneDistance += float32(8.0 * deltaTime)
	}

	if keyboardState[sdl.SCANCODE_F] == 1 {
		nearPlaneDistance -= float32(8.0 * deltaTime)
	}

	if previousRenderingModeKeyToggleState == 1 && keyboardState[sdl.SCANCODE_TAB] == 0 {
		switch currentRenderingMode {
		case Affine2D:
			currentRenderingMode = Mode7
		case Mode7:
			currentRenderingMode = Affine2D
		}
	}

	previousRenderingModeKeyToggleState = keyboardState[sdl.SCANCODE_TAB]
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
	currentTimeStamp := sdl.GetTicks()
	lastTimeStamp := uint32(0)
	countedFrames := 0
	timeSinceLastTick := uint32(0)
	for running {
		lastTimeStamp = currentTimeStamp
		currentTimeStamp = sdl.GetTicks()

		timeSinceLastTick += currentTimeStamp - lastTimeStamp
		if timeSinceLastTick > frameRateCap {
			running = processSDLEvents()

			frameDeltaTime := float32(timeSinceLastTick) / 1000.0
			processInput(frameDeltaTime)

			// Render logic
			countedFrames++
			averageFramesPerSecond := float32(countedFrames) / float32(sdl.GetTicks()-startTime) * 1000.0
			window.SetTitle(fmt.Sprintf("Avg FPS: %f framerate cap: %d scale: %f near %f", averageFramesPerSecond, frameRateCap, cameraScale.X(), nearPlaneDistance))

			clearRenderTarget(targetPixels)

			switch currentRenderingMode {
			case Affine2D:
				rasterBackground(targetPixels, mapPixels, mapImage)
			case Mode7:
				rasterBackgroundMode7Basic(targetPixels, mapPixels, mapImage)
			}

			texture.Update(nil, targetPixels, windowSizeX*stride)
			window.UpdateSurface()

			renderer.Clear()
			renderer.Copy(texture, nil, nil)
			renderer.Present()

			timeSinceLastTick = 0
		}
	}
}
