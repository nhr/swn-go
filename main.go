package main

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"

	"swn-go/internal/handlers"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	// Extract SQLite database to temp file (SQLite needs a real file)
	dbData, err := staticFS.ReadFile("static/swn.sqlite")
	if err != nil {
		log.Fatalf("Failed to read embedded database: %v", err)
	}
	tmpDB := filepath.Join(os.TempDir(), "swn_go.sqlite")
	if err := os.WriteFile(tmpDB, dbData, 0644); err != nil {
		log.Fatalf("Failed to write temp database: %v", err)
	}
	defer os.Remove(tmpDB)

	db, err := sql.Open("sqlite", tmpDB+"?mode=ro")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Load language data files for alien name generation
	langData := make(map[string]string)
	langFiles := []string{"barsoomian", "glorantha", "jorune", "klingon", "lovecraftian", "sindarin", "tsolyani"}
	for _, lang := range langFiles {
		data, err := staticFS.ReadFile("static/Includes/" + lang + ".txt")
		if err == nil {
			langData[lang] = string(data)
		}
	}

	// Load TiddlyWiki template
	swnHTML, err := staticFS.ReadFile("static/Includes/swn.html")
	if err != nil {
		log.Fatalf("Failed to read TiddlyWiki template: %v", err)
	}

	// Load map images
	bgPNG, err := staticFS.ReadFile("static/Includes/swnmap.png")
	if err != nil {
		log.Fatalf("Failed to read map background: %v", err)
	}
	dotPNG, err := staticFS.ReadFile("static/Includes/swndot.png")
	if err != nil {
		log.Fatalf("Failed to read dot image: %v", err)
	}

	// Load font
	fontData, _ := staticFS.ReadFile("static/Fonts/ProFontWindows.ttf")

	h := &handlers.Handlers{
		DB:       db,
		LangData: langData,
		SwnHTML:  string(swnHTML),
		BgPNG:    bgPNG,
		DotPNG:   dotPNG,
		FontData: fontData,
	}

	// Set up routes
	mux := http.NewServeMux()

	// CGI routes (matching the original Perl app's URL structure)
	mux.HandleFunc("/CGI/seed.cgi", h.SeedHandler)
	mux.HandleFunc("/CGI/sectorgen.cgi", h.SectorGenHandler)
	mux.HandleFunc("/CGI/iemap.cgi", h.IEMapHandler)

	// Static file serving from embedded FS
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(staticSub))

	// Serve index.html at root
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" || path == "/index.html" || path == "/index.pl" {
			data, err := staticFS.ReadFile("static/index.html")
			if err != nil {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(data)
			return
		}
		// Strip leading slash for static files and check if they exist
		cleanPath := strings.TrimPrefix(path, "/")
		if _, err := staticFS.ReadFile("static/" + cleanPath); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})

	fmt.Printf("SWN Sector Generator listening on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
