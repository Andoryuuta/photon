package photon

import (
	"image"
	"image/color"
	"image/draw"
)

const (
	FLAG_SET_PIXELS = 0x80
)

var (
	PixelSetColor   = color.RGBA{255, 255, 255, 255}
	PixelUnsetColor = color.RGBA{0, 0, 0, 255}
)

func decodeLayerImageData(imageData []byte, screenHeight uint32, screenWidth uint32) *image.RGBA {

	// Parse image data into PNG
	img := image.NewRGBA(image.Rect(0, 0, int(screenWidth), int(screenHeight)))
	draw.Draw(img, img.Bounds(), &image.Uniform{PixelUnsetColor}, image.ZP, draw.Src)

	// Left right, up to down
	pixelIndex := uint32(0)
	for i := 0; i < len(imageData); i++ {
		val := imageData[i] & 0x7F

		// If MSB is set, then the remaining 7 bits of the byte represent how many pixels to skip.
		if imageData[i] < 0x80 {
			pixelIndex += uint32(val)
			continue
		}

		if val != 0 {
			// val contains the amount of pixels to fill on this column downwards
			for j := 0; j < int(val); j++ {
				if pixelIndex > screenHeight*screenWidth {
					break
				}

				y := pixelIndex % screenHeight
				x := pixelIndex / screenHeight

				img.Set(int(x), int(y), PixelSetColor)
				pixelIndex++
			}
		}
	}

	return img
}

func encodeLayerImageData(img *image.RGBA) []byte {
	var output []byte

	var screenHeight = img.Bounds().Max.Y
	var screenWidth = img.Bounds().Max.X

	var unsetCount uint8 = 0
	var setCount uint8 = 0

	maxPixelIndex := screenWidth * screenHeight
	for pixelIndex := 0; pixelIndex < maxPixelIndex; pixelIndex++ {

		y := pixelIndex % screenHeight
		x := pixelIndex / screenHeight

		if img.At(x, y) == PixelUnsetColor {
			if setCount != 0 {
				// Previous pixels were set, this was not.
				output = append(output, setCount|FLAG_SET_PIXELS)
				setCount = 0
			}

			unsetCount++
			if unsetCount >= 0x7f-2 { // why -2?
				output = append(output, unsetCount)
				unsetCount = 0
			}
		} else if img.At(x, y) == PixelSetColor {
			if unsetCount != 0 {
				// Previous pixels were unset, this was not.
				output = append(output, unsetCount)
				unsetCount = 0
			}

			setCount++
			if setCount >= 0x7f-2 { // why -2?
				output = append(output, setCount|FLAG_SET_PIXELS)
				setCount = 0
			}
		}
	}

	// Set any leftover data
	if setCount != 0 {
		// Previous pixels were set, this was not.
		output = append(output, setCount|FLAG_SET_PIXELS)
		setCount = 0
	}

	if unsetCount != 0 {
		// Previous pixels were unset, this was not.
		output = append(output, unsetCount)
		unsetCount = 0
	}

	return output
}
