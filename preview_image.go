package photon

import (
	"image"
	"image/color"
	"image/draw"
)

func decodePreview(data []uint16, imageHeight uint32, imageWidth uint32) *image.RGBA {
	/*
		Preview image can override filename text:
		// Square it
		maxDim := uint32(maxInt(int(imageHeight), int(imageWidth)))
	*/

	maxDim := imageWidth

	// Parse image data into PNG
	img := image.NewRGBA(image.Rect(0, 0, int(imageWidth), int(imageHeight)))
	//img := image.NewRGBA(image.Rect(0, 0, int(maxDim), int(maxDim)))
	draw.Draw(img, img.Bounds(), &image.Uniform{PixelSetColor}, image.ZP, draw.Src)

	// Left right, up to down
	pixelIndex := uint32(0)
	for i := 0; i < len(data); i++ {
		s := data[i]

		// Split the uint16
		r, g, b, isFill := SplitRGB5515(s)

		// If the isFill bit is set, the next uint16 contains how many pixels we should fill with the current color.
		if isFill {
			idx := i + 1
			if idx >= len(data) {
				break
			}

			// The slicer sets the 2 bits above 0xFFF by doing:
			//		(s = s | 0x3000)
			// for unknown reasons, so we mask out the lower 12 bits.
			fillCount := data[idx] & 0xFFF

			// Fill gap
			targetPixelIndex := pixelIndex + uint32(fillCount)
			for ; pixelIndex < targetPixelIndex; pixelIndex++ {
				x := pixelIndex % maxDim
				y := pixelIndex / maxDim
				img.Set(int(x), int(y), color.RGBA{r, g, b, 255})
			}

			i += 1
		}

		x := pixelIndex % maxDim
		y := pixelIndex / maxDim

		img.Set(int(x), int(y), color.RGBA{r, g, b, 255})

		pixelIndex += 1
	}

	return img
}

func encodePreview(img *image.RGBA) []uint16 {
	var output []uint16

	var imageHeight = img.Bounds().Max.Y
	var imageWidth = img.Bounds().Max.X

	/*
		Preview image can override filename text:
		// Square it
		maxDim := maxInt(imageHeight, imageWidth)
		maxPixelIndex := maxDim * maxDim
	*/
	maxDim := imageWidth
	maxPixelIndex := imageHeight * imageWidth

	pixelAt := func(pi int) color.RGBA {
		x := pi % maxDim
		y := pi / maxDim
		return img.At(x, y).(color.RGBA)
	}

	for pixelIndex := 0; pixelIndex <= maxPixelIndex; pixelIndex++ {
		p := pixelAt(pixelIndex)

		if p != pixelAt(pixelIndex+1) || p != pixelAt(pixelIndex+2) || pixelIndex+2 >= maxPixelIndex {
			output = append(output, CombineRGB5515(p.R, p.G, p.B, false))
		} else {

			// Count skips
			var skipCount uint16 = 3
			for ; skipCount < 0xFFF && p == pixelAt(pixelIndex+int(skipCount)); skipCount++ {
			}

			output = append(output, CombineRGB5515(p.R, p.G, p.B, true)|0x20)
			output = append(output, skipCount-1|0x3000)
			pixelIndex += int(skipCount - 1)
		}
	}

	return output
}
