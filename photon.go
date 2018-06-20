package photon

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
)

type PhotonFile struct {
	PlateX             float32
	PlateY             float32
	PlateZ             float32
	LayerThickness     float32
	NormalExposureTime float32
	BottomExposureTime float32
	OffTime            float32
	BottomLayers       uint32
	ScreenHeight       uint32
	ScreenWidth        uint32
	LightCuringType    uint32 // ProjectionType

	PreviewImage   *image.RGBA
	ThumbnailImage *image.RGBA

	Layers []Layer
}

type Layer struct {
	RawData         []byte
	AbsoluteHeight  float32
	ExposureTime    float32
	PerLayerOffTime float32
}

type binCompatFileHeader struct {
	Magic1                       uint32 // Always 0x12FD0019
	Magic2                       uint32 // Always 0x01
	PlateX                       float32
	PlateY                       float32
	PlateZ                       float32
	Field_14                     uint32
	Field_18                     uint32
	Field_1C                     uint32
	LayerThickness               float32
	NormalExposureTime           float32
	BottomExposureTime           float32
	OffTime                      float32
	BottomLayers                 uint32
	ScreenHeight                 uint32
	ScreenWidth                  uint32
	PreviewHeaderOffset          uint32
	LayerHeadersOffset           uint32
	TotalLayers                  uint32
	PreviewThumbnailHeaderOffset uint32
	Field_4C                     uint32
	LightCuringType              uint32 // ProjectionType
	Field_54                     uint32
	Field_58                     uint32
	Field_60                     uint32
	Field_5C                     uint32
	Field_64                     uint32
	Field_68                     uint32
}

type binCompatPreviewHeader struct {
	Width             uint32
	Height            uint32
	PreviewDataOffset uint32
	PreviewDataSize   uint32
	Field_10          uint64 // Unused, always 0
	Field_18          uint64 // Unused, always 0
}

type binCompatLayerHeader struct {
	AbsoluteHeight  float32
	ExposureTime    float32
	PerLayerOffTime float32 // This is normally set to the file headers OffTime in all layers.

	// Most significant bit is seek type
	// switch(ImageDataOffset>>31)
	//		case 0: from start of file (Only seen this one actually being used.)
	//		case 1: relative (probably...)
	ImageDataOffset uint32
	ImageDataSize   uint32
	Field_14        uint64 // Unused, always 0
	Field_1C        uint64 // Unused, always 0
}

func Decode(rdr io.ReadSeeker) (*PhotonFile, error) {
	// Read main file header
	var header binCompatFileHeader
	err := binary.Read(rdr, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}

	// Read layers
	rdr.Seek(int64(header.LayerHeadersOffset), io.SeekStart)
	layerHeaders := make([]binCompatLayerHeader, header.TotalLayers)

	err = binary.Read(rdr, binary.LittleEndian, &layerHeaders)
	if err != nil {
		return nil, err
	}

	// Read in the preview image data
	var previewHeader binCompatPreviewHeader
	rdr.Seek(int64(header.PreviewHeaderOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &previewHeader)
	if err != nil {
		return nil, err
	}
	previewData := make([]uint16, previewHeader.PreviewDataSize/2)
	rdr.Seek(int64(previewHeader.PreviewDataOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &previewData)
	if err != nil {
		return nil, err
	}
	previewImg := decodePreview(previewData, previewHeader.Height, previewHeader.Width)

	// Read in the thumbnail image data
	var thumbnailHeader binCompatPreviewHeader
	rdr.Seek(int64(header.PreviewThumbnailHeaderOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &thumbnailHeader)
	if err != nil {
		return nil, err
	}
	thumbnailData := make([]uint16, thumbnailHeader.PreviewDataSize/2)
	rdr.Seek(int64(thumbnailHeader.PreviewDataOffset), io.SeekStart)
	err = binary.Read(rdr, binary.LittleEndian, &thumbnailData)
	if err != nil {
		return nil, err
	}
	thumbnailImg := decodePreview(thumbnailData, thumbnailHeader.Height, thumbnailHeader.Width)

	var layers []Layer
	for _, layer := range layerHeaders {
		// Read in the image data
		imageData := make([]byte, layer.ImageDataSize)

		rdr.Seek(int64(layer.ImageDataOffset), io.SeekStart)
		err = binary.Read(rdr, binary.LittleEndian, &imageData)
		if err != nil {
			return nil, err
		}

		// Deocode image data
		//img := decodeLayerImageData(imageData, header.ScreenHeight, header.ScreenWidth)

		layers = append(layers, Layer{
			//Image:           img,
			RawData:         imageData,
			AbsoluteHeight:  layer.AbsoluteHeight,
			ExposureTime:    layer.ExposureTime,
			PerLayerOffTime: layer.PerLayerOffTime,
		})
	}

	return &PhotonFile{
		PlateX:             header.PlateX,
		PlateY:             header.PlateY,
		PlateZ:             header.PlateZ,
		LayerThickness:     header.LayerThickness,
		NormalExposureTime: header.NormalExposureTime,
		BottomExposureTime: header.BottomExposureTime,
		OffTime:            header.OffTime,
		BottomLayers:       header.BottomLayers,
		ScreenHeight:       header.ScreenHeight,
		ScreenWidth:        header.ScreenWidth,
		LightCuringType:    header.LightCuringType,
		PreviewImage:       previewImg,
		ThumbnailImage:     thumbnailImg,
		Layers:             layers,
	}, nil
}

/*
header

previewHeader
previewData
thumbnailHeader
thumbnailData

layer0Header
layer1Header
layer2Header
...
layer9Header

layer0Data
layer1Data
layer2Data
...
layer9Data
*/

// Encodes the data in .photon / .cbddlp file format to the given writer.
func (pf *PhotonFile) EncodeTo(writer io.Writer) error {

	err := error(nil)
	previewData := U16ToU8Slice(encodePreview(pf.PreviewImage))
	thumbnailData := U16ToU8Slice(encodePreview(pf.ThumbnailImage))

	/*
		previewData, err := ioutil.ReadFile("saved_preview_835x321.bin")
		if err != nil {
			return err
		}

		thumbnailData, err := ioutil.ReadFile("saved_thumbnail_199x72.bin")
		if err != nil {
			return err
		}
	*/

	var layerDatas [][]byte
	for i := 0; i < len(pf.Layers); i++ {
		layerDatas = append(layerDatas, pf.Layers[i].RawData)
	}

	// Pre-calculate offsets so that we don't have to fixup the offset fields later.
	pos := 0
	pos += binary.Size(binCompatFileHeader{})

	// Preview offsets
	previewHeaderOffset := pos
	pos += binary.Size(binCompatPreviewHeader{})
	previewDataOffset := pos
	pos += len(previewData)

	// Thumbnail offsets
	thumbnailHeaderOffset := pos
	pos += binary.Size(binCompatPreviewHeader{})
	thumbnailDataOffset := pos
	pos += len(thumbnailData)

	// Layer headers offsets
	var layerHeaderOffsets []int
	for i := 0; i < len(pf.Layers); i++ {
		layerHeaderOffsets = append(layerHeaderOffsets, pos)
		pos += binary.Size(binCompatLayerHeader{})
	}

	// Layer data offsets
	var layerDataOffsets []int
	for i := 0; i < len(pf.Layers); i++ {
		layerDataOffsets = append(layerDataOffsets, pos)
		pos += len(layerDatas[i])
	}

	// Start forming and writing the file from here
	header := binCompatFileHeader{
		Magic1:                       0x12FD0019,
		Magic2:                       0x01,
		PlateX:                       pf.PlateX,
		PlateY:                       pf.PlateY,
		PlateZ:                       pf.PlateZ,
		LayerThickness:               pf.LayerThickness,
		NormalExposureTime:           pf.NormalExposureTime,
		BottomExposureTime:           pf.BottomExposureTime,
		OffTime:                      pf.OffTime,
		BottomLayers:                 pf.BottomLayers,
		ScreenHeight:                 pf.ScreenHeight,
		ScreenWidth:                  pf.ScreenWidth,
		PreviewHeaderOffset:          uint32(previewHeaderOffset),
		LayerHeadersOffset:           uint32(layerHeaderOffsets[0]),
		TotalLayers:                  uint32(len(pf.Layers)),
		PreviewThumbnailHeaderOffset: uint32(thumbnailHeaderOffset),
		LightCuringType:              pf.LightCuringType,
	}

	previewHeader := binCompatPreviewHeader{
		Width:/*835, // */ uint32(pf.PreviewImage.Bounds().Max.X),
		Height:/*321, //*/ uint32(pf.PreviewImage.Bounds().Max.Y),
		PreviewDataOffset: uint32(previewDataOffset),
		PreviewDataSize:   uint32(len(previewData)),
	}

	thumbnailHeader := binCompatPreviewHeader{
		Width:/*199, //*/ uint32(pf.ThumbnailImage.Bounds().Max.X),
		Height:/*72,  //*/ uint32(pf.ThumbnailImage.Bounds().Max.Y),
		PreviewDataOffset: uint32(thumbnailDataOffset),
		PreviewDataSize:   uint32(len(thumbnailData)),
	}

	var layerHeaders []binCompatLayerHeader
	for idx, l := range pf.Layers {
		layerHeaders = append(layerHeaders, binCompatLayerHeader{
			AbsoluteHeight:  l.AbsoluteHeight,
			ExposureTime:    l.ExposureTime,
			PerLayerOffTime: l.PerLayerOffTime,
			ImageDataOffset: uint32(layerDataOffsets[idx]),
			ImageDataSize:   uint32(len(layerDatas[idx])),
		})
	}

	err = binary.Write(writer, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.LittleEndian, previewHeader)
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.LittleEndian, previewData)
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.LittleEndian, thumbnailHeader)
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.LittleEndian, thumbnailData)
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.LittleEndian, layerHeaders)
	if err != nil {
		return err
	}

	for _, s := range layerDatas {
		err = binary.Write(writer, binary.LittleEndian, s)
		if err != nil {
			return err
		}
	}

	return nil
}
