package main

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"image"
	"image/png"

	"golang.org/x/image/draw"
)

// The tray's icons, drawn from the logo: in its colors for the systems
// that show it so, and as a monogram for macOS's menu bar, which tints a
// template icon itself.
//
//go:embed logo.png
var logoColor []byte

//go:embed logo-template.png
var logoTemplate []byte

var (
	logoImage     = decodeIcon(logoColor)
	templateImage = decodeIcon(logoTemplate)
)

func decodeIcon(data []byte) image.Image {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		panic("the embedded tray icon is not a PNG: " + err.Error())
	}
	return img
}

// icon draws the logo size pixels wide, as a PNG: in its colors, or in
// black for macOS's menu bar, which tints a template icon itself. A faded
// icon tells that Mira does not run.
func icon(size int, template, faded bool) []byte {
	source := logoImage
	if template {
		source = templateImage
	}
	rgba := image.NewNRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(rgba, rgba.Bounds(), source, source.Bounds(), draw.Over, nil)
	if faded {
		for i := 3; i < len(rgba.Pix); i += 4 {
			rgba.Pix[i] = uint8(uint16(rgba.Pix[i]) * 0x70 / 0xFF)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, rgba)
	return buf.Bytes()
}

// ico wraps a PNG in an icon file, which Windows' tray loads.
func ico(pngData []byte, size int) []byte {
	var buf bytes.Buffer
	// The header: reserved, type 1 (icon), one image.
	_ = binary.Write(&buf, binary.LittleEndian, [3]uint16{0, 1, 1})
	dimension := uint8(size)
	if size >= 256 {
		dimension = 0
	}
	buf.Write([]byte{dimension, dimension, 0, 0})
	_ = binary.Write(&buf, binary.LittleEndian, [2]uint16{1, 32})
	_ = binary.Write(&buf, binary.LittleEndian, [2]uint32{uint32(len(pngData)), 22})
	buf.Write(pngData)
	return buf.Bytes()
}
