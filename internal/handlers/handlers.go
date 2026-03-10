package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strings"

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
	resp := map[string]interface{}{
		"entries": []string{token},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SectorGenHandler handles sector generation requests.
func (h *Handlers) SectorGenHandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	token := r.FormValue("token")
	isie := r.FormValue("isie")
	instars := r.FormValue("stars")
	forIE := isie == "1"

	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Seed RNG from token
	seed := util.UntokenizeSeed(token)
	rand.Seed(seed)

	gen := generator.New(h.DB, h.LangData)

	sectorName, starMap, worlds := gen.GenSector()
	npcs := gen.GenNPCs(0)
	corps := gen.GenCorps(0)
	rels := gen.GenReligions(0)
	pols := gen.GenPolParties(0)
	aliens := gen.GenAliens(0)

	// Apply custom star positions if provided
	if instars != "" {
		var customStars []*generator.Star
		if err := json.Unmarshal([]byte(instars), &customStars); err == nil && len(customStars) > 0 {
			starMap = customStars
		}
	}

	if action == "display" {
		// Render map (unless IE)
		var mapData interface{}
		if !forIE {
			mapBytes, _ := render.RenderMap(sectorName, starMap, h.BgPNG, h.DotPNG, h.FontData, false)
			mapData = "data:image/png;base64," + string(mapBytes)
		}

		// Build display tables
		ntbl := [][]string{{"Name", "M/F", "Age", "Height"}}
		sortedNPCs := make([]*generator.NPC, len(npcs))
		copy(sortedNPCs, npcs)
		sort.Slice(sortedNPCs, func(i, j int) bool { return sortedNPCs[i].Sort < sortedNPCs[j].Sort })
		for _, npc := range sortedNPCs {
			ntbl = append(ntbl, []string{npc.Name, string(npc.Gender[0]), npc.Age, npc.Height})
		}

		ctbl := [][]string{{"Company", "Business"}}
		sortedCorps := make([]*generator.Corp, len(corps))
		copy(sortedCorps, corps)
		sort.Slice(sortedCorps, func(i, j int) bool { return sortedCorps[i].Name < sortedCorps[j].Name })
		for _, c := range sortedCorps {
			ctbl = append(ctbl, []string{c.Name, c.Business})
		}

		rtbl := [][]string{{"Name", "Leadership"}}
		sortedRels := make([]*generator.Religion, len(rels))
		copy(sortedRels, rels)
		sort.Slice(sortedRels, func(i, j int) bool { return sortedRels[i].Name < sortedRels[j].Name })
		for _, r := range sortedRels {
			l := strings.SplitN(r.Leadership, ".", 2)[0]
			rtbl = append(rtbl, []string{r.Name, l})
		}

		ptbl := [][]string{{"Organization", "Leadership", "Policy", "Outsiders", "Issues"}}
		sortedPols := make([]*generator.PoliticalParty, len(pols))
		copy(sortedPols, pols)
		sort.Slice(sortedPols, func(i, j int) bool { return sortedPols[i].Name < sortedPols[j].Name })
		for _, p := range sortedPols {
			le := strings.SplitN(p.Leadership, ":", 2)[0]
			pl := strings.SplitN(p.Policy, ":", 2)[0]
			re := strings.SplitN(p.Relationship, ":", 2)[0]
			issues := p.Issues[0].Tag + ", " + p.Issues[1].Tag
			ptbl = append(ptbl, []string{p.Name, le, pl, re, issues})
		}

		wtbl := [][]string{{"Name", "Atmo.", "Temp.", "Biosphere", "Population", "TL", "Tags"}}
		sortedWorlds := make([]*generator.World, len(worlds))
		copy(sortedWorlds, worlds)
		sort.Slice(sortedWorlds, func(i, j int) bool { return sortedWorlds[i].Name < sortedWorlds[j].Name })
		for _, wd := range sortedWorlds {
			wtbl = append(wtbl, []string{wd.Name, wd.Atmosphere.Short, wd.Temperature.Short,
				wd.Biosphere.Short, wd.Population.Short, wd.TechLevel.Short,
				wd.Tags[0].Short + ", " + wd.Tags[1].Short})
		}

		atbl := [][]string{{"Name", "Body Type", "Lenses", "Structure"}}
		sortedAliens := make([]*generator.Alien, len(aliens))
		copy(sortedAliens, aliens)
		sort.Slice(sortedAliens, func(i, j int) bool { return sortedAliens[i].Name < sortedAliens[j].Name })
		for _, a := range sortedAliens {
			btxt := a.Body[0][0]
			if len(a.Body) > 1 {
				btxt = a.Body[0][0] + ", " + a.Body[1][0]
			}
			ltxt := a.Lens[0][0] + ", " + a.Lens[1][0]
			stxt := a.Social[0][0]
			if len(a.Social) > 1 {
				stxt = "Multiple"
			}
			atbl = append(atbl, []string{a.Name, btxt, ltxt, stxt})
		}

		resp := map[string]interface{}{
			"name":   []string{sectorName},
			"token":  token,
			"map":    mapData,
			"stars":  starMap,
			"worlds": wtbl,
			"npcs":   ntbl,
			"corps":  ctbl,
			"rels":   rtbl,
			"pols":   ptbl,
			"aliens": atbl,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	} else if action == "create" {
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
			h.BgPNG, h.DotPNG, h.FontData, forIE)

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition",
			fmt.Sprintf("attachment;filename=SWN_Generator_%s.zip", token))
		w.Write(zipData)
	}
}

// IEMapHandler renders a map image for IE.
func (h *Handlers) IEMapHandler(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
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
