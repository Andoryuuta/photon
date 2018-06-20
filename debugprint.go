package photon

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Prototyping/ debug function.
// Dont even try to use.
func DebugPrint(rdr io.ReadSeeker) error {
	// Read main file header
	var header binCompatFileHeader
	err := binary.Read(rdr, binary.LittleEndian, &header)
	if err != nil {
		return err
	}

	fmt.Printf("%#v\n", header)

	// Read layers
	rdr.Seek(int64(header.LayerHeadersOffset), io.SeekStart)
	layerHeaders := make([]binCompatLayerHeader, header.TotalLayers)

	err = binary.Read(rdr, binary.LittleEndian, &layerHeaders)
	if err != nil {
		return err
	}

	// Read in the image data
	var previewHeader binCompatPreviewHeader

	rdr.Seek(int64(header.PreviewHeaderOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &previewHeader)
	if err != nil {
		return err
	}

	fmt.Printf("previewHeader %#v\n", previewHeader)

	previewData := make([]uint16, previewHeader.PreviewDataSize/2)
	rdr.Seek(int64(previewHeader.PreviewDataOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &previewData)
	if err != nil {
		return err
	}

	var thumbnailHeader binCompatPreviewHeader

	rdr.Seek(int64(header.PreviewthumbnailHeaderOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &thumbnailHeader)
	if err != nil {
		return err
	}

	fmt.Printf("thumbnailHeader %#v\n", thumbnailHeader)

	thumbnailData := make([]uint16, thumbnailHeader.PreviewDataSize/2)
	rdr.Seek(int64(thumbnailHeader.PreviewDataOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &thumbnailData)
	if err != nil {
		return err
	}

	for layerIdx, layer := range layerHeaders {
		fmt.Printf("Layer %v:%#v\n", layerIdx, layer)
	}

	return nil
}

/*
// Prototyping/ debug function.
// Dont even try to use.
func Read(rdr io.ReadSeeker) error {
	// Read main file header
	var header binCompatFileHeader
	err := binary.Read(rdr, binary.LittleEndian, &header)
	if err != nil {
		return err
	}

	fmt.Printf("%#v\n", header)

	// Read layers
	rdr.Seek(int64(header.LayerHeadersOffset), io.SeekStart)
	layerHeaders := make([]binCompatLayerHeader, header.TotalLayers)

	err = binary.Read(rdr, binary.LittleEndian, &layerHeaders)
	if err != nil {
		return err
	}

	////////////////////////////////////////////////////
	// Read in the image data
	var previewHeader binCompatPreviewHeader

	rdr.Seek(int64(header.PreviewHeaderOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &previewHeader)
	if err != nil {
		return err
	}

	fmt.Printf("previewHeader %#v\n", previewHeader)

	previewData := make([]uint16, previewHeader.PreviewDataSize/2)
	rdr.Seek(int64(previewHeader.PreviewDataOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &previewData)
	if err != nil {
		return err
	}

	var thumbnailHeader binCompatPreviewHeader

	rdr.Seek(int64(header.PreviewthumbnailHeaderOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &thumbnailHeader)
	if err != nil {
		return err
	}

	fmt.Printf("thumbnailHeader %#v\n", thumbnailHeader)

	thumbnailData := make([]uint16, thumbnailHeader.PreviewDataSize/2)
	rdr.Seek(int64(thumbnailHeader.PreviewDataOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &thumbnailData)
	if err != nil {
		return err
	}

	fmt.Println("Decoding start")
	img := decodePreview(previewData, previewHeader.Height, previewHeader.Width)

	fmt.Println("Encoding start")
	encImg := encodePreview(img)

	fmt.Println("comp start")
	fmt.Printf("img == encImg: %v\n", comparePixelBuf(previewData, encImg))

	ioutil.WriteFile("orgP.bin", U16ToU8Slice(previewData), 0666)
	ioutil.WriteFile("encP.bin", U16ToU8Slice(encImg), 0666)

	ioutil.WriteFile(fmt.Sprintf("preview_%vx%v.bin", previewHeader.Width, previewHeader.Height), U16ToU8Slice(previewData), 0666)
	ioutil.WriteFile(fmt.Sprintf("thumbnail_%vx%v.bin", thumbnailHeader.Width, thumbnailHeader.Height), U16ToU8Slice(thumbnailData), 0666)

	fmt.Println("Decoding enc start")
	img_reenc := decodePreview(encImg, previewHeader.Height, previewHeader.Width)

	thumb := decodePreview(thumbnailData, thumbnailHeader.Height, thumbnailHeader.Width)
	encThumb := encodePreview(thumb)
	thumb_reenc := decodePreview(encThumb, thumbnailHeader.Height, thumbnailHeader.Width)

	// Write to png file
	f, err := os.Create("preview.png")
	if err != nil {
		return err
	}
	png.Encode(f, img)
	f.Close()

	// Write to png file
	f, err = os.Create("preview_reenc.png")
	if err != nil {
		return err
	}
	png.Encode(f, img_reenc)
	f.Close()

	// Write to png file
	f, err = os.Create("thumbnail.png")
	if err != nil {
		return err
	}
	png.Encode(f, thumb)
	f.Close()

	// Write to png file
	f, err = os.Create("thumbnail_reenc.png")
	if err != nil {
		return err
	}
	png.Encode(f, thumb_reenc)
	f.Close()

	//ioutil.WriteFile("previewData.bin", previewData, 0666)

	return nil

	////////////////////////////////////////////////////

	for layerIdx, layer := range layerHeaders {
		fmt.Printf("Layer %v:%#v\n", layerIdx, layer)

		// Read in the image data
		imageData := make([]byte, layer.ImageDataSize)

		rdr.Seek(int64(layer.ImageDataOffset), io.SeekStart)
		err = binary.Read(rdr, binary.LittleEndian, &imageData)
		if err != nil {
			return err
		}

		// Deocode image data
		img := decodeLayerImageData(imageData, header.ScreenHeight, header.ScreenWidth)

		// Write to png file
		f, err := os.Create(fmt.Sprintf("layer_%v.png", layerIdx))
		if err != nil {
			return err
		}
		defer f.Close()

		png.Encode(f, img)
	}

	return nil
}
*/
