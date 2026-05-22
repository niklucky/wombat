package tray

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// GenerateIcon returns the tray icon with an optional badge overlay.
func GenerateIcon(activeCount int) ([]byte, error) {
	data, err := os.ReadFile("assets/tray-icon.png")
	if err != nil {
		return nil, err
	}
	base, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := base.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, base, bounds.Min, draw.Src)

	if activeCount > 0 {
		badgeColor := color.RGBA{52, 199, 89, 255} // green for 1
		if activeCount > 1 {
			badgeColor = color.RGBA{255, 59, 48, 255} // red for >1
		}

		// Draw badge circle in top-right corner
		cx := bounds.Dx() - 14
		cy := 14
		radius := 11
		fillCircle(img, cx, cy, radius, badgeColor)

		if activeCount > 1 {
			label := fmt.Sprintf("%d", activeCount)
			drawCenteredLabel(img, cx, cy, label)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func fillCircle(img draw.Image, cx, cy, r int, c color.Color) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				img.Set(cx+x, cy+y, c)
			}
		}
	}
}

func drawCenteredLabel(img draw.Image, cx, cy int, label string) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: basicfont.Face7x13,
	}

	// Measure text width
	adv := d.MeasureString(label)
	width := adv.Round()
	height := 13 // basicfont.Face7x13 height

	x := cx - width/2
	y := cy + height/2 - 3 // slight offset for visual centering

	d.Dot = fixed.P(x, y)
	d.DrawString(label)
}
