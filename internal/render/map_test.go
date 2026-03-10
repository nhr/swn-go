package render

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"swn-go/internal/generator"
)

func loadTestAssets(t *testing.T) (bgPNG, dotPNG, fontData []byte) {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	incDir := filepath.Join(root, "static", "Includes")
	fontDir := filepath.Join(root, "static", "Fonts")

	var err error
	bgPNG, err = os.ReadFile(filepath.Join(incDir, "swnmap.png"))
	if err != nil {
		t.Skipf("Map background not found: %v", err)
	}
	dotPNG, err = os.ReadFile(filepath.Join(incDir, "swndot.png"))
	if err != nil {
		t.Skipf("Dot image not found: %v", err)
	}
	fontData, _ = os.ReadFile(filepath.Join(fontDir, "ProFontWindows.ttf"))
	return
}

func TestRenderMap_ReturnsData(t *testing.T) {
	bgPNG, dotPNG, fontData := loadTestAssets(t)

	stars := []*generator.Star{
		{Name: "Alpha", Count: 1, Cell: "0101", ID: 0},
		{Name: "Beta", Count: 1, Cell: "0305", ID: 1},
	}

	imgData, areas := RenderMap("Test Sector", stars, bgPNG, dotPNG, fontData, false)
	if len(imgData) == 0 {
		t.Error("RenderMap returned empty image data")
	}
	if areas == "" {
		t.Error("RenderMap returned empty areas")
	}

	// Should be base64 encoded (not forIE)
	_, err := base64.StdEncoding.DecodeString(string(imgData))
	if err != nil {
		t.Errorf("Image data is not valid base64: %v", err)
	}
}

func TestRenderMap_ForIE_ReturnsRawPNG(t *testing.T) {
	bgPNG, dotPNG, fontData := loadTestAssets(t)

	stars := []*generator.Star{
		{Name: "Alpha", Count: 1, Cell: "0101", ID: 0},
	}

	imgData, _ := RenderMap("Test Sector", stars, bgPNG, dotPNG, fontData, true)
	if len(imgData) == 0 {
		t.Fatal("RenderMap returned empty image data")
	}

	// Should be a valid PNG
	_, err := png.Decode(bytes.NewReader(imgData))
	if err != nil {
		t.Errorf("Image data is not valid PNG: %v", err)
	}
}

func TestRenderMap_AreasContainStarNames(t *testing.T) {
	bgPNG, dotPNG, fontData := loadTestAssets(t)

	stars := []*generator.Star{
		{Name: "Alpha", Count: 1, Cell: "0101", ID: 0},
		{Name: "Beta", Count: 1, Cell: "0305", ID: 1},
	}

	_, areas := RenderMap("Test Sector", stars, bgPNG, dotPNG, fontData, false)
	if !strings.Contains(areas, "ALPHA") {
		t.Error("Areas do not contain ALPHA")
	}
	if !strings.Contains(areas, "BETA") {
		t.Error("Areas do not contain BETA")
	}
}

func TestRenderMap_NoStars(t *testing.T) {
	bgPNG, dotPNG, fontData := loadTestAssets(t)

	imgData, areas := RenderMap("Empty Sector", []*generator.Star{}, bgPNG, dotPNG, fontData, false)
	if len(imgData) == 0 {
		t.Error("RenderMap returned empty image data for empty star list")
	}
	if areas != "" {
		t.Errorf("Expected empty areas for no stars; got %q", areas)
	}
}
