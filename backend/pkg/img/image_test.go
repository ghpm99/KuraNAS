package img

import (
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func solidImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestThumbnailAndResize(t *testing.T) {
	src := solidImage(400, 200, color.RGBA{R: 255, A: 255})
	thumb := Thumbnail(src, 100, 100)
	if thumb.Bounds().Dx() != 100 || thumb.Bounds().Dy() != 100 {
		t.Fatalf("expected 100x100 thumbnail, got %dx%d", thumb.Bounds().Dx(), thumb.Bounds().Dy())
	}

	thumbDefault := Thumbnail(src, 0, 0)
	if thumbDefault.Bounds().Dx() != 320 || thumbDefault.Bounds().Dy() != 320 {
		t.Fatalf("expected default 320x320 thumbnail, got %dx%d", thumbDefault.Bounds().Dx(), thumbDefault.Bounds().Dy())
	}

	resized := resizeBilinear(src, 50, 25)
	if resized.Bounds().Dx() != 50 || resized.Bounds().Dy() != 25 {
		t.Fatalf("expected 50x25 resized image")
	}
}

func TestOpenDecodeAndEncode(t *testing.T) {
	tmp := t.TempDir()
	base := solidImage(16, 16, color.RGBA{G: 255, A: 255})

	jpegPath := filepath.Join(tmp, "a.jpg")
	pngPath := filepath.Join(tmp, "b.png")
	gifPath := filepath.Join(tmp, "c.gif")

	jf, err := os.Create(jpegPath)
	if err != nil {
		t.Fatalf("failed to create jpeg file: %v", err)
	}
	if err := jpeg.Encode(jf, base, nil); err != nil {
		t.Fatalf("failed to encode jpeg: %v", err)
	}
	_ = jf.Close()

	pf, err := os.Create(pngPath)
	if err != nil {
		t.Fatalf("failed to create png file: %v", err)
	}
	if err := png.Encode(pf, base); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}
	_ = pf.Close()

	gf, err := os.Create(gifPath)
	if err != nil {
		t.Fatalf("failed to create gif file: %v", err)
	}
	if err := gif.Encode(gf, base, nil); err != nil {
		t.Fatalf("failed to encode gif: %v", err)
	}
	_ = gf.Close()

	if _, _, err := OpenImageFromFile(pngPath); err != nil {
		t.Fatalf("expected open image success, err=%v", err)
	}
	if _, err := DecodeJPEG(jpegPath); err != nil {
		t.Fatalf("expected decode jpeg success, err=%v", err)
	}
	if _, err := DecodePNG(pngPath); err != nil {
		t.Fatalf("expected decode png success, err=%v", err)
	}
	if _, err := DecodeGIF(gifPath); err != nil {
		t.Fatalf("expected decode gif success, err=%v", err)
	}
	if _, err := EncodePNG(base); err != nil {
		t.Fatalf("expected encode png success, err=%v", err)
	}
}

func TestDecodeAndOpenErrors(t *testing.T) {
	if _, _, err := OpenImageFromFile("/non/existent/path.png"); err == nil {
		t.Fatalf("expected open error for missing file")
	}
	if _, err := DecodeJPEG("/non/existent/path.jpg"); err == nil {
		t.Fatalf("expected decode jpeg error")
	}
	if _, err := DecodePNG("/non/existent/path.png"); err == nil {
		t.Fatalf("expected decode png error")
	}
	if _, err := DecodeGIF("/non/existent/path.gif"); err == nil {
		t.Fatalf("expected decode gif error")
	}
}
