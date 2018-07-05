package main

import (
	_ "fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/Andoryuuta/photon"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	extractPreview   = kingpin.Flag("extract-preview", "Extract the preview files").Default("false").Bool()
	debugPrint       = kingpin.Flag("debugprint", "Print debug information about the file").Default("false").Bool()
	replacePreview   = kingpin.Flag("replace-preview", "Replace the preview image with the given .png").HintOptions("custom_preview.png").ExistingFile()
	replaceThumbnail = kingpin.Flag("replace-thumbnail", "Replace the thumbnail image with the given .png").HintOptions("custom_preview.png").ExistingFile()
	extractDir       = kingpin.Flag("extractdir", "Extraction directory.").Default("./").String()
	inputFile        = kingpin.Arg("input", "Input .photon/.cbddlp file").Required().ExistingFile()
	outputFile       = kingpin.Arg("output", "Output .photon/.cbddlp file").String()
)

func main() {
	// Parse command line
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("0.0.1").Author("Andrew Gutekanst")
	kingpin.CommandLine.Help = "photontool is a tool for working with .photon/.cbddlp or any other file that matches the Chitu D series DLP file format.\n\nSee http://github.com/Andoryuuta/photon for more information."
	kingpin.Parse()

	input, err := os.Open(*inputFile)
	if err != nil {
		log.Panicf("Failed to open file '%s': %v\n", *inputFile, err)
	}

	pfi, err := photon.Decode(input)
	if err != nil {
		log.Panicf("Failed to decode input file: %v\n", err)
	}

	if *debugPrint {
		dif, err := os.Open(*inputFile)
		if err != nil {
			log.Panicf("Failed to open file '%s' for debug printing: %v\n", *inputFile, err)
		}
		err = photon.DebugPrint(dif)
		if err != nil {
			log.Panicf("Failed to debug print!: %v\n", err)
		}
	}

	if *extractPreview {
		log.Println("Extracting preview images...")
		err := extractPreviewImages(pfi)
		if err != nil {
			log.Panicf("Error extracting preview images: %v\n", err)
		}

		log.Println("Preview images extracted.")
	}

	if *replacePreview != "" {
		f, err := os.Open(*replacePreview)
		if err != nil {
			log.Panicf("Error replacing preview image: %v\n", err)
		}
		defer f.Close()
		img, err := png.Decode(f)
		customPreview := image.NewRGBA(img.Bounds())
		draw.Draw(customPreview, customPreview.Bounds(), img, image.Pt(0, 0), draw.Src)
		pfi.PreviewImage = customPreview

		log.Println("Replaced preview image.")
	}

	if *replaceThumbnail != "" {
		f, err := os.Open(*replaceThumbnail)
		if err != nil {
			log.Panicf("Error replacing thumbnail image: %v\n", err)
		}
		defer f.Close()
		img, err := png.Decode(f)
		customPreview := image.NewRGBA(img.Bounds())
		draw.Draw(customPreview, customPreview.Bounds(), img, image.Pt(0, 0), draw.Src)
		pfi.ThumbnailImage = customPreview

		log.Println("Replaced thumbnail image.")
	}

	if *outputFile != "" {
		of, err := os.Create(*outputFile)
		if err != nil {
			log.Panicf("Error creating output file '%v': %v\n", *outputFile, err)
		}

		err = pfi.EncodeTo(of)
		if err != nil {
			log.Panicf("Failed to encode output file: %v\n", err)
		}
	}

	log.Println("Completed!")
}

func extractPreviewImages(pf *photon.PhotonFile) error {
	// Write preview
	f, err := os.Create(*extractDir + "preview.png")
	if err != nil {
		return err
	}
	defer f.Close()

	png.Encode(f, pf.PreviewImage)

	// Write thumbnail
	f2, err := os.Create(*extractDir + "preview-thumbnail.png")
	if err != nil {
		return err
	}
	defer f2.Close()

	png.Encode(f2, pf.ThumbnailImage)

	return nil
}
