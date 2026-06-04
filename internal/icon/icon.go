package icon

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// State holds tunnel counts for badge generation.
type State struct {
	Connected int
	Total     int
}

// Generate overlays a badge with "connected/total" onto the base PNG icon.
// It returns PNG-encoded bytes suitable for systray.SetIcon.
func Generate(basePNG []byte, state State) ([]byte, error) {
	baseImg, err := png.Decode(bytes.NewReader(basePNG))
	if err != nil {
		return nil, fmt.Errorf("decode base icon: %w", err)
	}

	bounds := baseImg.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, baseImg, bounds.Min, draw.Src)

	if state.Total > 0 {
		drawBadge(img, state)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode icon: %w", err)
	}
	return buf.Bytes(), nil
}

func drawBadge(img *image.RGBA, state State) {
	badgeColor := color.Black
	badgeW, badgeH := 128, 80
	margin := 4

	bounds := img.Bounds()
	x0 := bounds.Max.X - badgeW - margin
	y0 := bounds.Max.Y - badgeH - margin
	x1 := x0 + badgeW
	y1 := y0 + badgeH

	// Draw rounded-rect badge (simple 2px corner cut)
	cornerRadius := 4
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			// Skip 2px corners for rounded effect
			if (x < x0+cornerRadius && y < y0+cornerRadius) ||
				(x >= x1-cornerRadius && y < y0+cornerRadius) ||
				(x < x0+cornerRadius && y >= y1-cornerRadius) ||
				(x >= x1-cornerRadius && y >= y1-cornerRadius) {
				continue
			}
			img.Set(x, y, badgeColor)
		}
	}

	// Draw text
	text := fmt.Sprintf("%d/%d", state.Connected, state.Total)
	face := basicfont.Face7x13
	textW := face.Width * len(text) // 7px per char
	textH := face.Height * 3        // 13px

	// Center text inside badge
	tx := x0 + (badgeW-textW)/2
	ty := y0 + (badgeH+textH)/2 - 2 // -2 for visual centering

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: face,
		Dot:  fixed.P(tx, ty),
	}
	d.DrawString(text)
}
