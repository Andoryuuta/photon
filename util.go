package photon

import (
	_ "fmt"
	"math"
)

func changeRange(fromMin uint32, fromMax uint32, toMin uint32, toMax uint32, number uint32) uint32 {
	return uint32(math.Round(float64(number-fromMin)*float64(toMax-toMin)/float64(fromMax-fromMin) + float64(toMin)))
}

func SplitRGB5515(x uint16) (r uint8, g uint8, b uint8, isFill bool) {
	// Split the bits
	rBits := uint8(x & 0x1f)
	fillSet := (x >> 5) & 1
	gBits := uint8((x >> 6) & 0x1f)
	bBits := uint8((x >> 11) & 0x1f)

	// Scale colors from the range of 0-31 to 0-255
	r = uint8(changeRange(0, 31, 0, 255, uint32(rBits)))
	g = uint8(changeRange(0, 31, 0, 255, uint32(gBits)))
	b = uint8(changeRange(0, 31, 0, 255, uint32(bBits)))

	return r, g, b, fillSet == 1
}

func CombineRGB5515(r uint8, g uint8, b uint8, isFill bool) uint16 {
	// Scale colors from the range of 0-255 to 0-31
	rBits := uint16(changeRange(0, 255, 0, 31, uint32(r)))
	gBits := uint16(changeRange(0, 255, 0, 31, uint32(g)))
	bBits := uint16(changeRange(0, 255, 0, 31, uint32(b)))

	fillBit := uint16(0)
	if isFill {
		fillBit = 1
	}

	var x uint16
	x |= ((rBits & 0x1F) << 0)
	x |= ((fillBit & 0x1) << 5)
	x |= ((gBits & 0x1F) << 6)
	x |= ((bBits & 0x1F) << 11)

	return x
}

func maxInt(x int, y int) int {
	if x > y {
		return x
	}
	return y
}

func comparePixelBuf(x []uint16, y []uint16) bool {
	for i := 0; i < len(x) && i < len(y); i++ {
		if x[i] != y[i] {
			return false
		}
	}
	if len(x) != len(y) {
		return false
	}
	return true
}

func U16ToU8Slice(in []uint16) []uint8 {
	out := make([]byte, 0, 2*len(in))
	for _, v := range in {
		out = append(out, byte((v)&0xFF))
		out = append(out, byte((v>>8)&0xFF))
	}
	return out
}
