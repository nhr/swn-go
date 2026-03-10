package generator

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	dbPath := filepath.Join(root, "static", "swn.sqlite")
	if _, err := os.Stat(dbPath); err != nil {
		t.Skipf("SQLite database not found at %s: %v", dbPath, err)
	}
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func testLangData(t *testing.T) map[string]string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	langData := make(map[string]string)
	for _, lang := range []string{"barsoomian", "glorantha", "jorune", "klingon", "lovecraftian", "sindarin", "tsolyani"} {
		data, err := os.ReadFile(filepath.Join(root, "static", "Includes", lang+".txt"))
		if err == nil {
			langData[lang] = string(data)
		}
	}
	return langData
}

func testGenerator(t *testing.T) *Generator {
	t.Helper()
	return New(testDB(t), testLangData(t))
}

func TestGenSector_StarCount(t *testing.T) {
	g := testGenerator(t)
	_, stars, worlds := g.GenSector()

	if len(stars) < 21 || len(stars) > 30 {
		t.Errorf("GenSector produced %d stars; want 21-30", len(stars))
	}
	if len(worlds) == 0 {
		t.Error("GenSector produced no worlds")
	}
	if len(worlds) > maxWorlds {
		t.Errorf("GenSector produced %d worlds; max is %d", len(worlds), maxWorlds)
	}
}

func TestGenSector_StarCellsUnique(t *testing.T) {
	g := testGenerator(t)
	_, stars, _ := g.GenSector()

	cells := make(map[string]bool)
	for _, s := range stars {
		if cells[s.Cell] {
			t.Errorf("Duplicate cell %s", s.Cell)
		}
		cells[s.Cell] = true
	}
}

func TestGenSector_WorldsHaveNames(t *testing.T) {
	g := testGenerator(t)
	_, _, worlds := g.GenSector()

	for i, w := range worlds {
		if w.Name == "" {
			t.Errorf("World %d has empty name", i)
		}
		if w.SysName == "" {
			t.Errorf("World %d has empty system name", i)
		}
		if w.Cell == "" {
			t.Errorf("World %d has empty cell", i)
		}
	}
}

func TestGenSector_WorldComponentsPopulated(t *testing.T) {
	g := testGenerator(t)
	_, _, worlds := g.GenSector()

	for i, w := range worlds {
		if w.Atmosphere.Name == "" {
			t.Errorf("World %d atmosphere name empty", i)
		}
		if w.Temperature.Name == "" {
			t.Errorf("World %d temperature name empty", i)
		}
		if w.Biosphere.Name == "" {
			t.Errorf("World %d biosphere name empty", i)
		}
		if w.Population.Name == "" {
			t.Errorf("World %d population name empty", i)
		}
		if w.TechLevel.Name == "" {
			t.Errorf("World %d tech level name empty", i)
		}
		if len(w.Tags) != 2 {
			t.Errorf("World %d has %d tags; want 2", i, len(w.Tags))
		}
	}
}

func TestGenNPCs(t *testing.T) {
	g := testGenerator(t)
	npcs := g.GenNPCs(10)

	if len(npcs) != 10 {
		t.Fatalf("GenNPCs returned %d; want 10", len(npcs))
	}
	for i, npc := range npcs {
		if npc.Name == "" {
			t.Errorf("NPC %d has empty name", i)
		}
		if npc.Gender != "Male" && npc.Gender != "Female" {
			t.Errorf("NPC %d has unexpected gender %q", i, npc.Gender)
		}
	}
}

func TestGenCorps(t *testing.T) {
	g := testGenerator(t)
	corps := g.GenCorps(10)

	if len(corps) != 10 {
		t.Fatalf("GenCorps returned %d; want 10", len(corps))
	}
	for i, c := range corps {
		if c.Name == "" {
			t.Errorf("Corp %d has empty name", i)
		}
		if c.Business == "" {
			t.Errorf("Corp %d has empty business", i)
		}
	}
}

func TestGenReligions(t *testing.T) {
	g := testGenerator(t)
	rels := g.GenReligions(5)

	if len(rels) != 5 {
		t.Fatalf("GenReligions returned %d; want 5", len(rels))
	}
	for i, r := range rels {
		if r.Name == "" {
			t.Errorf("Religion %d has empty name", i)
		}
		if r.Origin == "" {
			t.Errorf("Religion %d has empty origin", i)
		}
		if len(r.Offshoots) == 0 {
			t.Errorf("Religion %d has no heresies", i)
		}
	}
}

func TestGenPolParties(t *testing.T) {
	g := testGenerator(t)
	pols := g.GenPolParties(5)

	if len(pols) != 5 {
		t.Fatalf("GenPolParties returned %d; want 5", len(pols))
	}
	for i, p := range pols {
		if p.Name == "" {
			t.Errorf("Party %d has empty name", i)
		}
		if len(p.Issues) != 2 {
			t.Errorf("Party %d has %d issues; want 2", i, len(p.Issues))
		}
	}
}

func TestGenAliens(t *testing.T) {
	g := testGenerator(t)
	aliens := g.GenAliens(5)

	if len(aliens) != 5 {
		t.Fatalf("GenAliens returned %d; want 5", len(aliens))
	}
	for i, a := range aliens {
		if a.Name == "" {
			t.Errorf("Alien %d has empty name", i)
		}
		if len(a.Body) == 0 {
			t.Errorf("Alien %d has no body types", i)
		}
		if len(a.Lens) != 2 {
			t.Errorf("Alien %d has %d lenses; want 2", i, len(a.Lens))
		}
		if len(a.Social) == 0 {
			t.Errorf("Alien %d has no social structure", i)
		}
	}
}

func TestGenNPCs_DefaultCount(t *testing.T) {
	g := testGenerator(t)
	npcs := g.GenNPCs(0)
	if len(npcs) != 100 {
		t.Errorf("GenNPCs(0) returned %d; want 100", len(npcs))
	}
}

func TestGenSector_SectorNameNotEmpty(t *testing.T) {
	g := testGenerator(t)
	name, _, _ := g.GenSector()
	if name == "" {
		t.Error("GenSector returned empty sector name")
	}
	parts := strings.SplitN(name, " ", 2)
	if len(parts) != 2 {
		t.Errorf("Sector name %q does not have two parts", name)
	}
}
