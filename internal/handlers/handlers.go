package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"

	"swn-go/internal/generator"
	"swn-go/internal/render"
	"swn-go/internal/util"
	"swn-go/internal/wiki"
)

// Handlers holds shared resources for HTTP handlers.
type Handlers struct {
	DB       *sql.DB
	LangData map[string]string
	SwnHTML  string
	BgPNG    []byte
	DotPNG   []byte
	FontData []byte
}

// SeedHandler returns a random seed token.
func (h *Handlers) SeedHandler(w http.ResponseWriter, r *http.Request) {
	token := util.TokenizeSeed(util.RandomSeed())
	resp := map[string]string{
		"token": token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// generateSector seeds the RNG and generates all sector data for the given token.
func (h *Handlers) generateSector(token string) (string, []*generator.Star, []*generator.World, []*generator.NPC, []*generator.Corp, []*generator.Religion, []*generator.PoliticalParty, []*generator.Alien) {
	seed := util.UntokenizeSeed(token)
	rand.Seed(seed)

	gen := generator.New(h.DB, h.LangData)

	sectorName, starMap, worlds := gen.GenSector()
	npcs := gen.GenNPCs(0)
	corps := gen.GenCorps(0)
	rels := gen.GenReligions(0)
	pols := gen.GenPolParties(0)
	aliens := gen.GenAliens(0)

	return sectorName, starMap, worlds, npcs, corps, rels, pols, aliens
}

// SectorHandler returns structured sector data as JSON.
func (h *Handlers) SectorHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	sectorName, starMap, worlds, npcs, corps, rels, pols, aliens := h.generateSector(token)

	sort.Slice(npcs, func(i, j int) bool { return npcs[i].Sort < npcs[j].Sort })
	sort.Slice(corps, func(i, j int) bool { return corps[i].Name < corps[j].Name })
	sort.Slice(rels, func(i, j int) bool { return rels[i].Name < rels[j].Name })
	sort.Slice(pols, func(i, j int) bool { return pols[i].Name < pols[j].Name })
	sort.Slice(worlds, func(i, j int) bool { return worlds[i].Name < worlds[j].Name })
	sort.Slice(aliens, func(i, j int) bool { return aliens[i].Name < aliens[j].Name })

	resp := &generator.Sector{
		Name:   sectorName,
		Token:  token,
		Stars:  starMap,
		Worlds: worlds,
		NPCs:   npcs,
		Corps:  corps,
		Rels:   rels,
		Pols:   pols,
		Aliens: aliens,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// MapHandler renders a sector map as a PNG image.
func (h *Handlers) MapHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	seed := util.UntokenizeSeed(token)
	rand.Seed(seed)

	gen := generator.New(h.DB, h.LangData)
	sectorName, starMap, _ := gen.GenSector()

	mapData, _ := render.RenderMap(sectorName, starMap, h.BgPNG, h.DotPNG, h.FontData, true)

	w.Header().Set("Content-Type", "image/png")
	w.Write(mapData)
}

// ExportHandler generates a TiddlyWiki ZIP export of the sector.
func (h *Handlers) ExportHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	sectorName, starMap, worlds, npcs, corps, rels, pols, aliens := h.generateSector(token)

	// Apply custom star positions if provided in the request body
	if r.Body != nil && r.ContentLength > 0 {
		var body struct {
			Stars []*generator.Star `json:"stars"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil && len(body.Stars) > 0 {
			starMap = body.Stars
		}
	}

	sector := &generator.Sector{
		Name:   sectorName,
		Token:  token,
		Stars:  starMap,
		Worlds: worlds,
		NPCs:   npcs,
		Corps:  corps,
		Rels:   rels,
		Pols:   pols,
		Aliens: aliens,
	}

	zipData := wiki.GenerateWiki(sector, h.SwnHTML, starMap, worlds,
		h.BgPNG, h.DotPNG, h.FontData, false)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment;filename=SWN_Sector_%s.zip", token))
	w.Write(zipData)
}

