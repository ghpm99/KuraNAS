package img

import (
	"image"
	"image/color"
	"image/draw"
)

func ThumbnailOnBlack(src image.Image, width, height uint) image.Image {
	fitted := Thumbnail(src, width, height)
	canvas := image.NewRGBA(fitted.Bounds())
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.Black), image.Point{}, draw.Src)
	draw.Draw(canvas, canvas.Bounds(), fitted, fitted.Bounds().Min, draw.Over)
	return canvas
}
