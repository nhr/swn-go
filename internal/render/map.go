package render

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"swn-go/internal/generator"
	"swn-go/internal/util"
)

const (
	mapX   = 38
	mapCol = 67
	mapMod = 38
	mapRow = 78
	fsize  = 9
)

// RenderMap generates a star map PNG image and area coordinate data.
// Returns (imageData, areas). If forIE is true, imageData is raw PNG bytes;
// otherwise it is base64-encoded.
func RenderMap(sectorName string, starMap []*generator.Star, bgPNG, dotPNG, fontData []byte, forIE bool) ([]byte, string) {
	// Load background image
	bgImg, err := png.Decode(bytes.NewReader(bgPNG))
	if err != nil {
		return nil, ""
	}

	// Load dot image
	dotImg, err := png.Decode(bytes.NewReader(dotPNG))
	if err != nil {
		return nil, ""
	}
	dotBounds := dotImg.Bounds()
	dotW := dotBounds.Dx()
	dotH := dotBounds.Dy()

	// Create mutable image from background
	bounds := bgImg.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, bgImg.At(x, y))
		}
	}

	// Load font
	var face font.Face
	if len(fontData) > 0 {
		tt, err := opentype.Parse(fontData)
		if err == nil {
			face, err = opentype.NewFace(tt, &opentype.FaceOptions{
				Size:    float64(fsize),
				DPI:     72,
				Hinting: font.HintingFull,
			})
			if err != nil {
				face = nil
			}
		}
	}

	// Build star reference map
	starRef := make(map[string]int)
	for _, s := range starMap {
		starRef[s.Cell] = s.ID
	}

	black := color.RGBA{0, 0, 0, 255}

	x := mapX
	y := 32

	var areas strings.Builder
	first := true

	for row := 0; row < 10; row++ {
		for col := 0; col < 8; col++ {
			cell := util.CellID(col, row)
			cname := ""
			if _, ok := starRef[cell]; ok {
				cname = starMap[starRef[cell]].Name
			}

			ly := y
			if col%2 == 1 {
				ly += mapMod
			} else {
				x++
			}

			if cname != "" {
				// Draw dot
				for dy := 0; dy < dotH; dy++ {
					for dx := 0; dx < dotW; dx++ {
						sc := dotImg.At(dotBounds.Min.X+dx, dotBounds.Min.Y+dy)
						_, _, _, a := sc.RGBA()
						if a > 0 {
							dst.Set(x+dx, ly+dy, sc)
						}
					}
				}

				// Draw star name
				nameUpper := strings.ToUpper(cname)
				if face != nil {
					textWidth := measureText(face, nameUpper)
					cx := x - textWidth/2 + 8
					cy := ly + 22
					drawText(dst, face, cx, cy, nameUpper, black)

					// Build coordinate data for area map
					coords := fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d",
						x, ly+dotH, x, ly, x+dotW, ly, x+dotW, ly+dotH,
						cx+textWidth, cy-fsize, cx+textWidth, cy+2, cx, cy+2, cx, cy-fsize,
						x, ly+dotH)

					if first {
						first = false
					} else {
						areas.WriteString("----\n")
					}
					areas.WriteString(util.SysName(cname, cell) + "\n" + coords + "\n")
				}
			}

			x += mapCol
		}
		x = mapX
		y += mapRow
	}

	// Encode image
	var buf bytes.Buffer
	png.Encode(&buf, dst)

	if forIE {
		return buf.Bytes(), areas.String()
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return []byte(encoded), areas.String()
}

func measureText(f font.Face, text string) int {
	advance := font.MeasureString(f, text)
	return advance.Ceil()
}

func drawText(img *image.RGBA, f font.Face, x, y int, text string, c color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: f,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}
