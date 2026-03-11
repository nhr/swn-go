package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"swn-go/internal/generator"
)

func testHandlers(t *testing.T) *Handlers {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	incDir := filepath.Join(root, "static", "Includes")

	dbPath := filepath.Join(root, "static", "swn.sqlite")
	if _, err := os.Stat(dbPath); err != nil {
		t.Skipf("SQLite database not found: %v", err)
	}
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	langData := make(map[string]string)
	for _, lang := range []string{"barsoomian", "glorantha", "jorune", "klingon", "lovecraftian", "sindarin", "tsolyani"} {
		data, err := os.ReadFile(filepath.Join(incDir, lang+".txt"))
		if err == nil {
			langData[lang] = string(data)
		}
	}

	bgPNG, _ := os.ReadFile(filepath.Join(incDir, "swnmap.png"))
	dotPNG, _ := os.ReadFile(filepath.Join(incDir, "swndot.png"))
	fontData, _ := os.ReadFile(filepath.Join(root, "static", "Fonts", "ProFontWindows.ttf"))
	swnHTML, _ := os.ReadFile(filepath.Join(incDir, "swn.html"))

	return &Handlers{
		DB:       db,
		LangData: langData,
		SwnHTML:  string(swnHTML),
		BgPNG:    bgPNG,
		DotPNG:   dotPNG,
		FontData: fontData,
	}
}

// newMux registers routes the same way main.go does, so PathValue works.
func newMux(h *Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/seed", h.SeedHandler)
	mux.HandleFunc("GET /api/sector/{token}", h.SectorHandler)
	mux.HandleFunc("GET /api/sector/{token}/map", h.MapHandler)
	mux.HandleFunc("POST /api/sector/{token}/export", h.ExportHandler)
	return mux
}

func TestSeedHandler(t *testing.T) {
	h := testHandlers(t)
	mux := newMux(h)
	req := httptest.NewRequest("GET", "/api/seed", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("SeedHandler returned status %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Invalid JSON response: %v", err)
	}

	if resp["token"] == "" {
		t.Error("Expected non-empty token")
	}
}

func TestSectorHandler(t *testing.T) {
	h := testHandlers(t)
	mux := newMux(h)
	req := httptest.NewRequest("GET", "/api/sector/ABC", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("SectorHandler returned status %d", w.Code)
	}

	var resp generator.Sector
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if resp.Name == "" {
		t.Error("Expected non-empty sector name")
	}
	if resp.Token != "ABC" {
		t.Errorf("Token = %q; want ABC", resp.Token)
	}
	if len(resp.Stars) == 0 {
		t.Error("Expected non-empty stars")
	}
	if len(resp.Worlds) == 0 {
		t.Error("Expected non-empty worlds")
	}
	if len(resp.NPCs) == 0 {
		t.Error("Expected non-empty NPCs")
	}
	if len(resp.Corps) == 0 {
		t.Error("Expected non-empty corps")
	}
	if len(resp.Rels) == 0 {
		t.Error("Expected non-empty religions")
	}
	if len(resp.Pols) == 0 {
		t.Error("Expected non-empty political parties")
	}
	if len(resp.Aliens) == 0 {
		t.Error("Expected non-empty aliens")
	}
}

func TestMapHandler(t *testing.T) {
	h := testHandlers(t)
	mux := newMux(h)
	req := httptest.NewRequest("GET", "/api/sector/ABC/map", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("MapHandler returned status %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("Content-Type = %q; want image/png", ct)
	}

	if w.Body.Len() == 0 {
		t.Error("MapHandler returned empty body")
	}
}

func TestExportHandler(t *testing.T) {
	h := testHandlers(t)
	mux := newMux(h)
	req := httptest.NewRequest("POST", "/api/sector/ABC/export", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ExportHandler returned status %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/zip" {
		t.Errorf("Content-Type = %q; want application/zip", ct)
	}

	disp := w.Header().Get("Content-Disposition")
	if !strings.Contains(disp, "SWN_Sector_ABC.zip") {
		t.Errorf("Content-Disposition = %q; want filename containing SWN_Sector_ABC.zip", disp)
	}
}

func TestExportHandler_EmptyBody(t *testing.T) {
	h := testHandlers(t)
	mux := newMux(h)

	// Export with an empty JSON body (no custom stars)
	req := httptest.NewRequest("POST", "/api/sector/ABC/export", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ExportHandler with empty body returned status %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/zip" {
		t.Errorf("Content-Type = %q; want application/zip", ct)
	}
}
