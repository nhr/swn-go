package wiki

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"swn-go/internal/generator"
)

func minimalSector() *generator.Sector {
	return &generator.Sector{
		Name:  "Test Sector",
		Token: "ABC",
		Stars: []*generator.Star{
			{Name: "Alpha", Count: 1, Cell: "0101", ID: 0},
		},
		Worlds: []*generator.World{
			{
				Name:        "TestWorld",
				Atmosphere:  generator.WorldComponent{Roll: 7, Name: "Breathable", Desc: "A breathable atmosphere", Short: "Breathable"},
				Temperature: generator.WorldComponent{Roll: 7, Name: "Temperate", Desc: "Temperate climate", Short: "Temperate"},
				Biosphere:   generator.WorldComponent{Roll: 7, Name: "Human-miscible", Desc: "Compatible biosphere", Short: "Miscible"},
				Population:  generator.WorldComponent{Roll: 7, Name: "Millions", Desc: "Several million", Short: "Millions"},
				TechLevel:   generator.WorldComponent{Roll: 7, Name: "TL4", Desc: "Baseline tech", Short: "TL4"},
				Tags: []generator.WorldTag{
					{Roll: "101", Name: "Tag1", Desc: "Desc1", Enemies: "E1", Friends: "F1", Complications: "C1", Things: "T1", Places: "P1", Short: "Tag1"},
					{Roll: "202", Name: "Tag2", Desc: "Desc2", Enemies: "E2", Friends: "F2", Complications: "C2", Things: "T2", Places: "P2", Short: "Tag2"},
				},
				SysCount: 1, SysPos: 1, SysName: "Alpha III", SysNum: "III", StarID: 0, Cell: "0101", CellPos: "010101",
			},
		},
		NPCs: []*generator.NPC{
			{Name: "John Doe", Sort: "Doe John", Gender: "Male", Age: "Middle-aged", Height: "Average", Problem: "Debt", Motive: "Greed", Quirk: "Fidgets"},
		},
		Corps: []*generator.Corp{
			{Name: "ACME Corp", Business: "Mining", Reputation: "Reliable"},
		},
		Rels: []*generator.Religion{
			{Name: "Test Faith", Origin: "Christianity", Leadership: "Council", Evolution: "Evolved",
				Offshoots: []generator.Heresy{{Name: "Reform", Attitude: "Hostile", Founder: "Priest", HerDesc: "Rejected dogma", Quirk: "Secretive"}}},
		},
		Pols: []*generator.PoliticalParty{
			{Name: "Unity Party", Issues: []generator.PoliticalIssue{{Issue: "Trade", Tag: "Trade"}, {Issue: "Defense", Tag: "Defense"}},
				Leadership: "Elected", Relationship: "Open", Policy: "Liberal"},
		},
		Aliens: []*generator.Alien{
			{Name: "Zyx", Body: [][]string{{"Humanlike", "Bipedal"}}, Lens: [][]string{{"Warlike", "Aggressive"}, {"Curious", "Inquisitive"}},
				Vars: []string{"4 eyes"}, Social: [][]string{{"Hive", "Collective"}},
				Arch: generator.Architecture{Name: "Architecture 1", Towers: "Spires", Foundations: "Stone", WallDecor: "Carved", Supports: "Columns", Arches: "Pointed", Extras: "Gargoyles"}},
		},
	}
}

func TestGenerateWiki_ReturnsValidZip(t *testing.T) {
	s := minimalSector()
	swnHTML := "$$SECTOR_NAME $$SEED_TOKEN $$SECTOR_MAP $$PLANETARY_DIRECTORY $$PLANET_LIST $$STARMAP $$STARURL $$MAP_LINKS $$GMINFO $$TIME_STAMP <!--$$SECTOR_INFO-->"

	// Use minimal 1x1 white PNG for bg and dot
	bgPNG := minimalPNG()
	dotPNG := minimalPNG()

	result := GenerateWiki(s, swnHTML, s.Stars, s.Worlds, bgPNG, dotPNG, nil, false)
	if len(result) == 0 {
		t.Fatal("GenerateWiki returned empty result")
	}

	r, err := zip.NewReader(bytes.NewReader(result), int64(len(result)))
	if err != nil {
		t.Fatalf("Result is not valid ZIP: %v", err)
	}

	expectedFiles := map[string]bool{
		"README.text":          false,
		"SWN_wiki_ABC_GM.html": false,
		"SWN_wiki_ABC_PC.html": false,
	}
	for _, f := range r.File {
		expectedFiles[f.Name] = true
	}
	for name, found := range expectedFiles {
		if !found {
			t.Errorf("Missing file in ZIP: %s", name)
		}
	}
}

func TestGenerateWiki_GMHasNPCs(t *testing.T) {
	s := minimalSector()
	swnHTML := "$$SECTOR_NAME <!--$$SECTOR_INFO-->"
	bgPNG := minimalPNG()

	result := GenerateWiki(s, swnHTML, s.Stars, s.Worlds, bgPNG, bgPNG, nil, false)
	r, _ := zip.NewReader(bytes.NewReader(result), int64(len(result)))

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "_GM.html") {
			rc, _ := f.Open()
			buf := new(bytes.Buffer)
			buf.ReadFrom(rc)
			rc.Close()
			content := buf.String()
			if !strings.Contains(content, "John Doe") {
				t.Error("GM wiki does not contain NPC name")
			}
			if !strings.Contains(content, "NPC Listing") {
				t.Error("GM wiki does not contain NPC Listing")
			}
		}
	}
}

func TestGenerateWiki_PCLacksGMInfo(t *testing.T) {
	s := minimalSector()
	swnHTML := "$$SECTOR_NAME $$GMINFO <!--$$SECTOR_INFO-->"
	bgPNG := minimalPNG()

	result := GenerateWiki(s, swnHTML, s.Stars, s.Worlds, bgPNG, bgPNG, nil, false)
	r, _ := zip.NewReader(bytes.NewReader(result), int64(len(result)))

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "_PC.html") {
			rc, _ := f.Open()
			buf := new(bytes.Buffer)
			buf.ReadFrom(rc)
			rc.Close()
			content := buf.String()
			if strings.Contains(content, "NPC Listing") {
				t.Error("PC wiki should not contain NPC Listing")
			}
			if strings.Contains(content, "GM Info") {
				t.Error("PC wiki should not contain GM Info")
			}
		}
	}
}

func TestMarkupSectorMap(t *testing.T) {
	stars := []*generator.Star{
		{Name: "Alpha", Count: 1, Cell: "0101", ID: 0},
	}
	worlds := []*generator.World{
		{Name: "TestWorld", StarID: 0, SysPos: 1, Cell: "0101"},
	}

	result := markupSectorMap(worlds, stars)
	if !strings.Contains(result, "TestWorld") {
		t.Error("Sector map markup does not contain world name")
	}
	if !strings.Contains(result, "0101") {
		t.Error("Sector map markup does not contain cell ID")
	}
}

func TestMarkupPlanetaryLists_PCVsGM(t *testing.T) {
	stars := []*generator.Star{
		{Name: "Alpha", Count: 1, Cell: "0101", ID: 0},
	}
	worlds := []*generator.World{
		{
			Name: "TestWorld", StarID: 0, SysPos: 1, Cell: "0101",
			Atmosphere:  generator.WorldComponent{Short: "Breathable"},
			Temperature: generator.WorldComponent{Short: "Temperate"},
			Biosphere:   generator.WorldComponent{Short: "Miscible"},
			Population:  generator.WorldComponent{Short: "Millions"},
			TechLevel:   generator.WorldComponent{Short: "TL4"},
			Tags: []generator.WorldTag{
				{Short: "Tag1"},
				{Short: "Tag2"},
			},
		},
	}

	gmResult := markupPlanetaryLists(worlds, stars, false)
	pcResult := markupPlanetaryLists(worlds, stars, true)

	if !strings.Contains(gmResult, "Tags") {
		t.Error("GM planetary list should contain Tags column")
	}
	if strings.Contains(pcResult, "Tags") {
		t.Error("PC planetary list should not contain Tags column")
	}
}

// minimalPNG returns a valid 1x1 white PNG image.
func minimalPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG signature
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, // 8-bit RGB
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc,
		0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, // IEND chunk
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
