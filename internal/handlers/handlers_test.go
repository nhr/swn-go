package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
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

func TestSeedHandler(t *testing.T) {
	h := testHandlers(t)
	req := httptest.NewRequest("GET", "/CGI/seed.cgi", nil)
	w := httptest.NewRecorder()

	h.SeedHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("SeedHandler returned status %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Invalid JSON response: %v", err)
	}

	entries, ok := resp["entries"].([]interface{})
	if !ok || len(entries) != 1 {
		t.Errorf("Expected entries array with 1 element; got %v", resp["entries"])
	}
}

func TestSectorGenHandler_Display(t *testing.T) {
	h := testHandlers(t)
	form := url.Values{
		"action": {"display"},
		"token":  {"ABC"},
	}
	req := httptest.NewRequest("POST", "/CGI/sectorgen.cgi", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.SectorGenHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("SectorGenHandler display returned status %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	for _, key := range []string{"name", "token", "map", "stars", "worlds", "npcs", "corps", "rels", "pols", "aliens"} {
		if _, ok := resp[key]; !ok {
			t.Errorf("Response missing key %q", key)
		}
	}
}

func TestSectorGenHandler_Create(t *testing.T) {
	h := testHandlers(t)
	form := url.Values{
		"action": {"create"},
		"token":  {"ABC"},
	}
	req := httptest.NewRequest("POST", "/CGI/sectorgen.cgi", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.SectorGenHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("SectorGenHandler create returned status %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/zip" {
		t.Errorf("Content-Type = %q; want application/zip", ct)
	}

	disp := w.Header().Get("Content-Disposition")
	if !strings.Contains(disp, "SWN_Generator_ABC.zip") {
		t.Errorf("Content-Disposition = %q; want filename containing SWN_Generator_ABC.zip", disp)
	}
}

func TestSectorGenHandler_MissingToken(t *testing.T) {
	h := testHandlers(t)
	form := url.Values{"action": {"display"}}
	req := httptest.NewRequest("POST", "/CGI/sectorgen.cgi", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.SectorGenHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing token; got %d", w.Code)
	}
}

func TestIEMapHandler(t *testing.T) {
	h := testHandlers(t)
	req := httptest.NewRequest("GET", "/CGI/iemap.cgi?token=ABC", nil)
	w := httptest.NewRecorder()

	h.IEMapHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("IEMapHandler returned status %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("Content-Type = %q; want image/png", ct)
	}

	if w.Body.Len() == 0 {
		t.Error("IEMapHandler returned empty body")
	}
}

func TestIEMapHandler_MissingToken(t *testing.T) {
	h := testHandlers(t)
	req := httptest.NewRequest("GET", "/CGI/iemap.cgi", nil)
	w := httptest.NewRecorder()

	h.IEMapHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing token; got %d", w.Code)
	}
}
