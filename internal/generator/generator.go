package generator

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"

	"swn-go/internal/conflux"
	"swn-go/internal/dice"
	"swn-go/internal/names"
	"swn-go/internal/util"
)

const maxWorlds = 36

// Generator holds state for sector generation.
type Generator struct {
	db        *sql.DB
	seenNames map[string]bool

	// World tag tracking
	tagIDs  []string
	tagUsed []bool
	totUsed int

	// Religion tracking
	relCombos map[string]bool

	// Heresy tracking
	heresyCombos map[string]bool

	// Political tracking
	polCombos map[string]bool
	issCombos map[string]bool
	seenElems map[string]map[int]bool

	// Architecture tracking
	seenArch map[string]bool

	// Alien tracking
	seenAliens map[string]bool

	// Alien name generation
	alienNameLists map[string][]string
	alienNameSeen  map[string]map[int]bool
	langData       map[string]string
}

// New creates a new Generator.
func New(db *sql.DB, langData map[string]string) *Generator {
	g := &Generator{
		db:             db,
		seenNames:      make(map[string]bool),
		relCombos:      make(map[string]bool),
		heresyCombos:   make(map[string]bool),
		polCombos:      make(map[string]bool),
		issCombos:      make(map[string]bool),
		seenArch:       make(map[string]bool),
		seenAliens:     make(map[string]bool),
		alienNameLists: make(map[string][]string),
		alienNameSeen:  make(map[string]map[int]bool),
		langData:       langData,
	}
	g.seenElems = map[string]map[int]bool{
		"animal":    {},
		"color":     {},
		"direction": {},
		"metal":     {},
	}
	g.initTagIDs()
	return g
}

func (g *Generator) initTagIDs() {
	g.tagIDs = nil
	g.tagUsed = nil
	for d6 := 1; d6 < 7; d6++ {
		for d10 := 1; d10 < 11; d10++ {
			id := fmt.Sprintf("%d%02d", d6, d10)
			g.tagIDs = append(g.tagIDs, id)
			g.tagUsed = append(g.tagUsed, false)
		}
	}
	g.totUsed = 0
}

func (g *Generator) checkName(name, typ string) bool {
	token := strings.ToUpper(typ) + ":" + strings.ToUpper(name)
	if g.seenNames[token] {
		return false
	}
	g.seenNames[token] = true
	return true
}

func (g *Generator) pickName(typ string) string {
	var listSet [][]string
	switch typ {
	case "cosmic":
		listSet = names.CosmicNames
	case "corp":
		listSet = names.LastNames
	case "male":
		listSet = names.MaleNames
	case "female":
		listSet = names.FemaleNames
	default:
		listSet = names.CosmicNames
	}

	for {
		subList := listSet[rand.Intn(len(listSet))]
		name := subList[rand.Intn(len(subList))]
		if typ == "male" || typ == "female" {
			lastList := names.LastNames[rand.Intn(len(names.LastNames))]
			name += " " + lastList[rand.Intn(len(lastList))]
		}
		if g.checkName(name, typ) {
			return name
		}
	}
}

var greekNames = []string{"Beta", "Gamma", "Delta", "Zeta", "Theta", "Kappa", "Lambda", "Rho", "Sigma", "Tau", "Psi", "Omega"}
var placeNames = []string{"Hades", "Amber", "Annwyn", "Asgard", "Avalon", "Dinas", "Elysian", "Reynes", "Saguenay", "Kvenland", "Lyonesse", "Meropis", "Olympus", "Shangra", "Tartarus", "Themis", "Valhallan"}

func (g *Generator) pickSectorName() string {
	pname := placeNames[rand.Intn(len(placeNames))]
	gname := greekNames[rand.Intn(len(greekNames))]
	return pname + " " + gname
}

var romanNums = map[int]string{
	3: "III", 4: "IV", 5: "V", 6: "VI",
	7: "VII", 8: "VIII", 9: "IX", 10: "X",
}

func (g *Generator) worldNums(count int) []string {
	useNums := make(map[int]bool)
	goodNums := 0
	for goodNums < count {
		num := dice.DInt(8) + 2
		if !useNums[num] {
			useNums[num] = true
			goodNums++
		}
	}

	// Sort keys
	sorted := make([]int, 0, len(useNums))
	for k := range useNums {
		sorted = append(sorted, k)
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	result := make([]string, 0, len(sorted))
	for _, n := range sorted {
		result = append(result, romanNums[n])
	}
	return result
}

func dWorldCount() int {
	firstRoll := dice.DInt(10)
	if firstRoll < 6 {
		return 1
	}
	if firstRoll < 9 {
		return 2
	}
	return 3
}

// GenSector generates a complete sector.
func (g *Generator) GenSector() (string, []*Star, []*World) {
	starCount := dice.DInt(10) + 20
	sectorName := g.pickSectorName()

	var starMap []*Star
	seenCells := make(map[string]bool)
	var worlds []*World
	currentStars := 0

	for i := 0; i < starCount; i++ {
		var row, col int
		var cell string
		for {
			row = dice.DInt(10) - 1
			col = dice.DInt(8) - 1
			cell = util.CellID(col, row)
			if !seenCells[cell] {
				break
			}
		}
		seenCells[cell] = true
		currentStars++
		starName := g.pickName("cosmic")

		worldCount := dWorldCount()
		remainingStars := starCount - currentStars
		remainingWorlds := maxWorlds - len(worlds)
		if remainingWorlds-worldCount < remainingStars {
			worldCount = 1
		}

		starMap = append(starMap, &Star{
			Name:  starName,
			Count: worldCount,
			Cell:  cell,
			ID:    i,
		})

		worldNums := g.worldNums(worldCount)

		for j := 0; j < worldCount; j++ {
			world := g.genWorld()
			world.SysCount = worldCount
			world.SysPos = j + 1
			world.SysName = starName + " " + worldNums[j]
			world.SysNum = worldNums[j]
			world.StarID = i
			world.Cell = cell
			world.CellPos = cell + fmt.Sprintf("%02d", j+1)
			worlds = append(worlds, world)
		}
	}

	return sectorName, starMap, worlds
}

func (g *Generator) genWorld() *World {
	return &World{
		Name:        g.pickName("cosmic"),
		Atmosphere:  g.genWorldComponent("atmosphere"),
		Biosphere:   g.genWorldComponent("biosphere"),
		Population:  g.genWorldComponent("population"),
		TechLevel:   g.genWorldComponent("tech_level"),
		Temperature: g.genWorldComponent("temperature"),
		Tags:        g.genWorldTags(),
	}
}

var worldTableMap = map[string]string{
	"atmosphere":  "WORLD_ATMOSPHERE",
	"biosphere":   "WORLD_BIOSPHERE",
	"population":  "WORLD_POPULATION",
	"tech_level":  "WORLD_TECH_LEVEL",
	"temperature": "WORLD_TEMPERATURE",
}

func (g *Generator) genWorldComponent(typ string) WorldComponent {
	roll := dice.D("6x2")
	table := worldTableMap[typ]

	var name, desc string
	var short sql.NullString
	err := g.db.QueryRow(
		fmt.Sprintf("SELECT name, desc, short FROM %s WHERE id = ?", table),
		roll,
	).Scan(&name, &desc, &short)
	if err != nil {
		return WorldComponent{Roll: roll}
	}

	shortStr := name
	if short.Valid {
		shortStr = short.String
	}

	return WorldComponent{Roll: roll, Name: name, Desc: desc, Short: shortStr}
}

func (g *Generator) genWorldTags() []WorldTag {
	tag1 := g.genWorldTag("")
	tag2 := g.genWorldTag(tag1.Roll)
	return []WorldTag{tag1, tag2}
}

func (g *Generator) genWorldTag(skip string) WorldTag {
	totTags := len(g.tagIDs)

	if g.totUsed >= totTags {
		for i := range g.tagUsed {
			g.tagUsed[i] = false
		}
		g.totUsed = 0
	}

	idx := rand.Intn(totTags)
	skipIdx := -1

	if skip != "" {
		for i, id := range g.tagIDs {
			if id == skip {
				skipIdx = i
				break
			}
		}
	}

	for g.tagUsed[idx] || (skipIdx >= 0 && idx == skipIdx) {
		idx++
		if idx >= totTags {
			idx = 0
		}
	}

	g.tagUsed[idx] = true
	g.totUsed++

	id := g.tagIDs[idx]

	var name, desc, enemies, friends, complications, things, places string
	var short sql.NullString
	err := g.db.QueryRow(
		"SELECT name, desc, enemies, friends, complications, things, places, short FROM WORLD_TAG WHERE id = ?",
		id,
	).Scan(&name, &desc, &enemies, &friends, &complications, &things, &places, &short)
	if err != nil {
		return WorldTag{Roll: id}
	}

	shortStr := name
	if short.Valid {
		shortStr = short.String
	}

	return WorldTag{
		Roll: id, Name: name, Desc: desc,
		Enemies: enemies, Friends: friends,
		Complications: complications, Things: things,
		Places: places, Short: shortStr,
	}
}

// GenNPCs generates NPCs.
func (g *Generator) GenNPCs(max int) []*NPC {
	if max == 0 {
		max = 100
	}

	npcs := make([]*NPC, 0, max)
	for i := 0; i < max; i++ {
		gr := dice.DInt(4)
		ar := dice.DInt(6)
		hr := dice.DInt(8)
		pr := dice.DInt(10)
		mr := dice.DInt(12)
		qr := dice.DInt(20)

		gender := g.queryString("SELECT gender FROM NPC_GENDER WHERE id = ?", gr)
		age := g.queryString("SELECT age FROM NPC_AGE WHERE id = ?", ar)
		height := g.queryString("SELECT height FROM NPC_HEIGHT WHERE id = ?", hr)
		problem := g.queryString("SELECT problem FROM NPC_PROBLEMS WHERE id = ?", pr)
		motive := g.queryString("SELECT motive FROM NPC_MOTIVATION WHERE id = ?", mr)
		quirk := g.queryString("SELECT quirk FROM NPC_QUIRKS WHERE id = ?", qr)

		// Handle gender-specific quirks: "male:X,female:Y"
		if strings.HasPrefix(quirk, "male:") && strings.Contains(quirk, ",female:") {
			parts := strings.SplitN(quirk, ",", 2)
			maleQuirk := strings.TrimPrefix(parts[0], "male:")
			femaleQuirk := strings.TrimPrefix(parts[1], "female:")
			if gender == "male" {
				quirk = maleQuirk
			} else {
				quirk = femaleQuirk
			}
		}

		name := g.pickName(gender)
		nameParts := strings.SplitN(name, " ", 2)
		sortName := name
		if len(nameParts) == 2 {
			sortName = nameParts[1] + " " + nameParts[0]
		}

		npcs = append(npcs, &NPC{
			Name:    name,
			Sort:    sortName,
			Gender:  strings.ToUpper(gender[:1]) + gender[1:],
			Age:     age,
			Height:  height,
			Problem: problem,
			Motive:  motive,
			Quirk:   quirk,
		})
	}
	return npcs
}

// GenCorps generates corporations.
func (g *Generator) GenCorps(max int) []*Corp {
	if max == 0 {
		max = 50
	}

	corps := make([]*Corp, 0, max)
	for i := 0; i < max; i++ {
		nameSource := dice.DInt(3)
		var name string

		if nameSource == 1 {
			for {
				roll := dice.DInt(25)
				name = g.queryString("SELECT name FROM QUICK_CORP_NAME WHERE id = ?", roll)
				if g.checkName(name, "corp") {
					break
				}
			}
		} else {
			name = g.pickName("corp")
		}

		org := g.queryString("SELECT organization FROM QUICK_CORP_NAME WHERE id = ?", dice.DInt(25))
		biz := g.queryString("SELECT business FROM QUICK_CORP_BUSINESS WHERE id = ?", dice.DInt(100))
		rep := g.queryString("SELECT reputation FROM QUICK_CORP_REPUTATION WHERE id = ?", dice.DInt(100))

		corps = append(corps, &Corp{
			Name:       name + " " + org,
			Business:   biz,
			Reputation: rep,
		})
	}
	return corps
}

// GenReligions generates religions.
func (g *Generator) GenReligions(max int) []*Religion {
	if max == 0 {
		max = 24
	}

	rels := make([]*Religion, 0, max)
	for i := 0; i < max; i++ {
		var oroll1, lroll, eroll int
		var token string
		for {
			oroll1 = dice.DInt(12)
			lroll = dice.DInt(6)
			eroll = dice.DInt(8)
			token = fmt.Sprintf("%d:%d", oroll1, eroll)
			if !g.relCombos[token] {
				break
			}
		}
		g.relCombos[token] = true

		var origin, oname string
		g.db.QueryRow("SELECT origin, name, adj FROM QUICK_RELIGION_ORIGIN WHERE id = ?", oroll1).Scan(&origin, &oname, new(string))

		leadership := g.queryString("SELECT leadership FROM QUICK_RELIGION_LEADERSHIP WHERE id = ?", lroll)

		var evolution, adjective string
		g.db.QueryRow("SELECT evolution, adjective FROM QUICK_RELIGION_EVOLUTION WHERE id = ?", eroll).Scan(&evolution, &adjective)

		rname := adjective

		// Handle Syncretism (eroll == 3)
		if eroll == 3 {
			var oroll2 int
			for {
				oroll2 = dice.DInt(12)
				if oroll2 != oroll1 {
					break
				}
			}
			var oadj2 string
			g.db.QueryRow("SELECT origin, name, adj FROM QUICK_RELIGION_ORIGIN WHERE id = ?", oroll2).Scan(new(string), new(string), &oadj2)
			rname += " " + oadj2 + " " + oname
		} else {
			rname += " " + oname
		}

		// Handle Leadership
		if lroll == 6 {
			lroll2 := dice.DInt(6)
			if lroll2 == 6 {
				leadership = "No universal leadership. This faith has no hierarchy."
			} else {
				leadership = "No universal leadership. Each region governs itself differently."
			}
		}

		// Heresies
		hroll := dice.DInt(6)
		hmax := 1
		if hroll == 6 {
			hmax = 3
		} else if hroll == 5 {
			hmax = 2
		}

		heresies := g.genHeresies(hmax, oroll1)

		rels = append(rels, &Religion{
			Name:       rname,
			Origin:     origin,
			Leadership: leadership,
			Evolution:  evolution,
			Offshoots:  heresies,
		})
	}
	return rels
}

func (g *Generator) genHeresies(max, origin int) []Heresy {
	heresies := make([]Heresy, 0, max)
	for i := 0; i < max; i++ {
		heresies = append(heresies, g.genHeresy(origin))
	}
	return heresies
}

func (g *Generator) genHeresy(origin int) Heresy {
	var aroll, hroll, qroll int
	var token string
	for {
		aroll = dice.DInt(10)
		_ = dice.DInt(8) // froll consumed but not used in token
		hroll = dice.DInt(12)
		qroll = dice.DInt(20)
		token = fmt.Sprintf("%d:%d", aroll, hroll)
		if !g.heresyCombos[token] {
			break
		}
	}

	// Handle syncretic heresies
	var sorigin string
	if hroll == 10 {
		var sroll int
		for {
			sroll = dice.DInt(12)
			if sroll != origin {
				break
			}
		}
		sorigin = g.queryString("SELECT origin FROM QUICK_RELIGION_ORIGIN WHERE id = ?", sroll)
	}

	g.heresyCombos[token] = true

	var attitude, adjective string
	g.db.QueryRow("SELECT attitude, adjective FROM QUICK_HERESY_ATTITUDE WHERE id = ?", aroll).Scan(&attitude, &adjective)
	founder := g.queryString("SELECT founder FROM QUICK_HERESY_FOUNDER WHERE id = ?", aroll)

	var heresy, hname string
	g.db.QueryRow("SELECT heresy, name FROM QUICK_HERESY_HERESY WHERE id = ?", hroll).Scan(&heresy, &hname)
	quirk := g.queryString("SELECT quirk FROM QUICK_HERESY_QUIRK WHERE id = ?", qroll)

	if hroll == 10 {
		heresy = strings.ReplaceAll(heresy, "$$RELIGION", sorigin)
	}

	return Heresy{
		Name:     adjective + " " + hname,
		Attitude: attitude,
		Founder:  founder,
		HerDesc:  heresy,
		Quirk:    quirk,
	}
}

// Political party generation
var polColors = []string{"Black", "Yellow", "Red", "Blue", "White", "Green", "Charteuse", "Violet", "Brown", "Crimson",
	"Amber", "Indigo", "Azure", "Carnelian", "Cerulean", "Emerald", "Lavender", "Ivory",
	"Jade", "Onyx", "Tyrian", "Vermilion", "Viridian"}
var polAnimals = []string{"Alligator", "Fire-Ant", "Bison", "Armadillo", "Badger", "Barracuda", "Bear", "Boar",
	"Buffalo", "Caribou", "Stag", "Cheetah", "Rooster", "Cobra", "Cormorant", "Coyote",
	"Crab", "Crane", "Crow", "Deer", "Dolphin", "Dove", "Dragonfly", "Eagle", "Wapiti", "Falcon",
	"Ferret", "Finch", "Fox", "Gazelle", "Panda", "Giraffe", "Goat", "Gorilla", "Gull", "Hawk",
	"Heron", "Hornet", "Thoroughbred", "Human", "Iguana", "Jackal", "Jaguar", "Kangaroo",
	"Koala", "Komodo", "Leopard", "Lion", "Locust", "Mallard", "Manatee", "Meerkat", "Moose",
	"Nightingale", "Ostrich", "Otter", "Owl", "Ox", "Oyster", "Panther", "Pelican", "Penguin",
	"Pigeon", "Platypus", "Porcupine", "Ram", "Raven", "Salamander", "Sea-Lion", "Seahorse",
	"Shark", "Snake", "Spider", "Swan", "Tiger", "Tortoise", "Walrus", "Wasp", "Wolf"}
var polDirections = []string{"North", "South", "East", "West", "Turnwise", "Widdershins", "Northeast", "Northwest",
	"Southeast", "Southwest", "Outer", "Inner", "Higher", "Lower"}
var polMetals = []string{"Lithium", "Magnesium", "Aluminum", "Titanium", "Chromium", "Iron", "Cobalt", "Copper",
	"Palladium", "Silver", "Iridium", "Platinum", "Gold", "Mercury", "Cerium", "Neodymium",
	"Uranium", "Plutonium", "Steel", "Bronze"}

var polLists = map[string][]string{
	"direction": polDirections,
	"animal":    polAnimals,
	"color":     polColors,
	"metal":     polMetals,
}

var polManaged = map[int]string{
	11: "color",
	14: "animal",
	17: "direction",
	20: "metal",
}

// GenPolParties generates political parties.
func (g *Generator) GenPolParties(max int) []*PoliticalParty {
	if max == 0 {
		max = 24
	}

	parties := make([]*PoliticalParty, 0, max)
	polIdx := 0
	for i := 0; i < max; i++ {
		var lroll, nroll1, nroll2, oroll, proll int
		var token string
		for {
			lroll = dice.DInt(8)
			nroll1 = dice.DInt(20)
			nroll2 = dice.DInt(20)
			oroll = dice.DInt(4)
			proll = dice.DInt(6)
			token = fmt.Sprintf("%d", nroll1)
			if _, managed := polManaged[nroll1]; managed {
				token += fmt.Sprintf("%d", polIdx)
				polIdx++
			}
			if !g.polCombos[token] {
				break
			}
		}
		g.polCombos[token] = true

		var iroll1, iroll2 int
		var itok string
		for {
			iroll1 = dice.DInt(12)
			for {
				iroll2 = dice.DInt(12)
				if iroll2 != iroll1 {
					break
				}
			}
			if iroll1 > iroll2 {
				itok = fmt.Sprintf("%d:%d", iroll2, iroll1)
			} else {
				itok = fmt.Sprintf("%d:%d", iroll1, iroll2)
			}
			if !g.issCombos[itok] {
				break
			}
		}
		g.issCombos[itok] = true

		var name string
		if lkey, managed := polManaged[nroll1]; managed {
			list := polLists[lkey]
			llen := len(list)
			var roll int
			for {
				roll = dice.DInt(llen)
				if !g.seenElems[lkey][roll] {
					break
				}
			}
			g.seenElems[lkey][roll] = true
			name = list[roll-1]
		} else {
			name = g.queryString("SELECT element1 FROM QUICK_POLITICAL_NAME WHERE id = ?", nroll1)
		}

		name2 := g.queryString("SELECT element2 FROM QUICK_POLITICAL_NAME WHERE id = ?", nroll2)
		name += " " + name2

		issue1, itag1 := g.queryTwoStrings("SELECT issue, tag FROM QUICK_POLITICAL_ISSUES WHERE id = ?", iroll1)
		issue2, itag2 := g.queryTwoStrings("SELECT issue, tag FROM QUICK_POLITICAL_ISSUES WHERE id = ?", iroll2)
		leadership := g.queryString("SELECT leadership FROM QUICK_POLITICAL_LEADERSHIP WHERE id = ?", lroll)
		relationship := g.queryString("SELECT relationship FROM QUICK_POLITICAL_OUTSIDERS WHERE id = ?", oroll)
		policy := g.queryString("SELECT policy FROM QUICK_POLITICAL_POLICY WHERE id = ?", proll)

		parties = append(parties, &PoliticalParty{
			Name: name,
			Issues: []PoliticalIssue{
				{Issue: issue1, Tag: itag1},
				{Issue: issue2, Tag: itag2},
			},
			Leadership:   leadership,
			Relationship: relationship,
			Policy:       policy,
		})
	}
	return parties
}

// GenAliens generates alien races.
func (g *Generator) GenAliens(max int) []*Alien {
	if max == 0 {
		max = 10
	}

	aliens := make([]*Alien, 0, max)
	for i := 0; i < max; i++ {
		aliens = append(aliens, g.genAlien())
	}
	return aliens
}

var alienVarTables = map[string]string{
	"Humanlike":  "ALIEN_VAR_MAMMAL",
	"Avian":      "ALIEN_VAR_AVIAN",
	"Reptilian":  "ALIEN_VAR_REPTILE",
	"Insectile":  "ALIEN_VAR_INSECT",
	"Exotic":     "ALIEN_VAR_EXOTIC",
}

func (g *Generator) genAlien() *Alien {
	var broll1, broll2, lroll1, lroll2, sroll int
	var token string
	hybrid := false

	for {
		broll1 = dice.DInt(6)
		if broll1 == 6 {
			hybrid = true
			broll1 = dice.DInt(5)
			for {
				broll2 = dice.DInt(5)
				if broll2 != broll1 {
					break
				}
			}
		} else {
			broll2 = 0
			hybrid = false
		}

		lroll1 = dice.DInt(20)
		for {
			lroll2 = dice.DInt(20)
			if lroll2 != lroll1 {
				break
			}
		}
		sroll = dice.DInt(6)
		stok := sroll
		if sroll == 6 {
			stok = 5
		}
		token = fmt.Sprintf("%d:%d:%d:%d:%d", broll1, broll2, lroll1, lroll2, stok)
		if !g.seenAliens[token] {
			break
		}
	}
	g.seenAliens[token] = true

	// Social structure
	var social [][]string
	totSoc := 1
	if sroll > 4 {
		totSoc = 1 + dice.DInt(3)
	}
	for i := 0; i < totSoc; i++ {
		roll := sroll
		if sroll > 4 {
			roll = dice.DInt(4)
		}
		sname, sdesc := g.queryTwoStrings("SELECT name, desc FROM ALIEN_STRUCTURE WHERE id = ?", roll)
		social = append(social, []string{sname, sdesc})
	}

	// Body types
	var body [][]string
	var vars []string

	body1name, body1desc := g.queryTwoStrings("SELECT name, desc FROM ALIEN_BODY_TYPE WHERE id = ?", broll1)
	body = append(body, []string{body1name, body1desc})

	vroll1 := dice.DInt(20)
	if varTable, ok := alienVarTables[body1name]; ok {
		var1 := g.queryString(fmt.Sprintf("SELECT variation FROM %s WHERE id = ?", varTable), vroll1)
		vars = append(vars, rollVar(var1))
	}

	var vroll2 int
	if hybrid {
		body2name, body2desc := g.queryTwoStrings("SELECT name, desc FROM ALIEN_BODY_TYPE WHERE id = ?", broll2)
		body = append(body, []string{body2name, body2desc})
		vroll2 = dice.DInt(20)
		if varTable, ok := alienVarTables[body2name]; ok {
			var2 := g.queryString(fmt.Sprintf("SELECT variation FROM %s WHERE id = ?", varTable), vroll2)
			vars = append(vars, rollVar(var2))
		}
	} else {
		for {
			vroll2 = dice.DInt(20)
			if vroll2 != vroll1 {
				break
			}
		}
		if varTable, ok := alienVarTables[body1name]; ok {
			var2 := g.queryString(fmt.Sprintf("SELECT variation FROM %s WHERE id = ?", varTable), vroll2)
			vars = append(vars, rollVar(var2))
		}
	}

	// Lenses
	lens1name, lens1desc := g.queryTwoStrings("SELECT name, desc FROM ALIEN_LENSES WHERE id = ?", lroll1)
	lens2name, lens2desc := g.queryTwoStrings("SELECT name, desc FROM ALIEN_LENSES WHERE id = ?", lroll2)

	name := g.genAlienName()

	return &Alien{
		Name:   name,
		Body:   body,
		Lens:   [][]string{{lens1name, lens1desc}, {lens2name, lens2desc}},
		Vars:   vars,
		Social: social,
		Arch:   g.genArch(1),
	}
}

func rollVar(txt string) string {
	d3 := fmt.Sprintf("%d", dice.DInt(3))
	d4 := fmt.Sprintf("%d", dice.DInt(4))
	d4x2 := fmt.Sprintf("%d", dice.D("4x2"))
	d10x2 := fmt.Sprintf("%d", dice.D("10x2"))
	d6p1 := fmt.Sprintf("%d", dice.D("6p1"))

	txt = strings.Replace(txt, "$$1D3", d3, 1)
	txt = strings.Replace(txt, "$$1D4", d4, 1)
	txt = strings.Replace(txt, "$$2D4", d4x2, 1)
	txt = strings.Replace(txt, "$$2D10", d10x2, 1)
	txt = strings.Replace(txt, "$$1D6P1", d6p1, 1)
	return txt
}

// genArch generates an architecture.
func (g *Generator) genArch(idx int) Architecture {
	var troll, froll, wroll, sroll, aroll, eroll int
	var token string
	for {
		troll = archID(1, dice.DInt(10))
		froll = archID(2, dice.DInt(10))
		wroll = archID(3, dice.DInt(10))
		sroll = archID(4, dice.DInt(10))
		aroll = archID(5, dice.DInt(10))
		eroll = archID(6, dice.DInt(10))
		token = fmt.Sprintf("%d:%d:%d:%d:%d:%d", troll, froll, wroll, sroll, aroll, eroll)
		if !g.seenArch[token] {
			break
		}
	}
	g.seenArch[token] = true

	return Architecture{
		Name:        fmt.Sprintf("Architecture %d", idx),
		Towers:      g.queryString("SELECT element FROM QUICK_ARCHITECTURE WHERE id = ?", troll),
		Foundations: g.queryString("SELECT element FROM QUICK_ARCHITECTURE WHERE id = ?", froll),
		WallDecor:   g.queryString("SELECT element FROM QUICK_ARCHITECTURE WHERE id = ?", wroll),
		Supports:    g.queryString("SELECT element FROM QUICK_ARCHITECTURE WHERE id = ?", sroll),
		Arches:      g.queryString("SELECT element FROM QUICK_ARCHITECTURE WHERE id = ?", aroll),
		Extras:      g.queryString("SELECT element FROM QUICK_ARCHITECTURE WHERE id = ?", eroll),
	}
}

func archID(idx, roll int) int {
	return idx*100 + roll
}

var alienLangLists = []string{"barsoomian", "glorantha", "jorune", "klingon", "lovecraftian", "sindarin", "tsolyani"}

func (g *Generator) genAlienName() string {
	list := alienLangLists[rand.Intn(len(alienLangLists))]

	if _, ok := g.alienNameLists[list]; !ok {
		data, ok := g.langData[list]
		if !ok {
			return "Unknown"
		}
		nl := conflux.Generate(data, 100)
		g.alienNameLists[list] = nl
		if _, exists := g.alienNameSeen[list]; !exists {
			g.alienNameSeen[list] = make(map[int]bool)
		}
	}

	thisList := g.alienNameLists[list]
	listLen := len(thisList)
	if listLen == 0 {
		return "Unknown"
	}

	var idx int
	for {
		idx = rand.Intn(listLen)
		if !g.alienNameSeen[list][idx] {
			break
		}
	}
	g.alienNameSeen[list][idx] = true

	name := thisList[idx]
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	return name
}

func (g *Generator) queryString(query string, args ...interface{}) string {
	var result string
	err := g.db.QueryRow(query, args...).Scan(&result)
	if err != nil {
		return ""
	}
	return result
}

func (g *Generator) queryTwoStrings(query string, args ...interface{}) (string, string) {
	var a, b string
	err := g.db.QueryRow(query, args...).Scan(&a, &b)
	if err != nil {
		return "", ""
	}
	return a, b
}
