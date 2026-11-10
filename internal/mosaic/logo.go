package mosaic

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	stddraw "image/draw"
	"image/png"

	xdraw "golang.org/x/image/draw"
)

// The logo, faint in the corner of the emblems' backdrop.
//
//go:embed mira.png
var logoPNG []byte

var logoImage = func() image.Image {
	img, err := png.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		panic("the embedded mosaic logo is not a PNG: " + err.Error())
	}
	return img
}()

// drawLogo paints the logo in the bottom right corner of canvas, faint
// enough to sit under the emblems' signs.
func drawLogo(canvas *image.RGBA) {
	const size = 420
	scaled := image.NewRGBA(image.Rect(0, 0, size, size))
	xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), logoImage, logoImage.Bounds(), xdraw.Over, nil)
	at := image.Pt(1400, 560)
	stddraw.DrawMask(canvas, image.Rectangle{Min: at, Max: at.Add(image.Pt(size, size))},
		scaled, image.Point{}, image.NewUniform(color.Alpha{A: 0x45}), image.Point{}, stddraw.Over)
}
