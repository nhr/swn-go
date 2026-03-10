package generator

// Star represents a star system on the sector map.
type Star struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Cell  string `json:"cell"`
	ID    int    `json:"id"`
}

// WorldComponent holds a rolled world attribute (atmosphere, biosphere, etc.)
type WorldComponent struct {
	Roll  int    `json:"roll"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Short string `json:"short"`
}

// WorldTag holds a rolled world tag.
type WorldTag struct {
	Roll          string `json:"roll"`
	Name          string `json:"name"`
	Desc          string `json:"desc"`
	Enemies       string `json:"enemies"`
	Friends       string `json:"friends"`
	Complications string `json:"complications"`
	Things        string `json:"things"`
	Places        string `json:"places"`
	Short         string `json:"short"`
}

// World represents a generated planet.
type World struct {
	Name        string         `json:"name"`
	Atmosphere  WorldComponent `json:"atmosphere"`
	Biosphere   WorldComponent `json:"biosphere"`
	Population  WorldComponent `json:"population"`
	TechLevel   WorldComponent `json:"tech_level"`
	Temperature WorldComponent `json:"temperature"`
	Tags        []WorldTag     `json:"tags"`
	SysCount    int            `json:"sys_count"`
	SysPos      int            `json:"sys_pos"`
	SysName     string         `json:"sys_name"`
	SysNum      string         `json:"sys_num"`
	StarID      int            `json:"star_id"`
	Cell        string         `json:"cell"`
	CellPos     string         `json:"cell_pos"`
	PNum        int            `json:"p_num,omitempty"`
}

// NPC represents a generated non-player character.
type NPC struct {
	Name    string `json:"name"`
	Sort    string `json:"sort"`
	Gender  string `json:"gender"`
	Age     string `json:"age"`
	Height  string `json:"height"`
	Problem string `json:"problem"`
	Motive  string `json:"motive"`
	Quirk   string `json:"quirk"`
}

// Corp represents a generated corporation.
type Corp struct {
	Name       string `json:"name"`
	Business   string `json:"business"`
	Reputation string `json:"reputation"`
}

// Heresy represents a heretical offshoot of a religion.
type Heresy struct {
	Name     string `json:"name"`
	Attitude string `json:"attitude"`
	Founder  string `json:"founder"`
	HerDesc  string `json:"heresy"`
	Quirk    string `json:"quirk"`
}

// Religion represents a generated religion.
type Religion struct {
	Name       string   `json:"name"`
	Origin     string   `json:"origin"`
	Leadership string   `json:"leadership"`
	Evolution  string   `json:"evolution"`
	Offshoots  []Heresy `json:"offshoots"`
}

// PoliticalIssue represents a political issue.
type PoliticalIssue struct {
	Issue string `json:"issue"`
	Tag   string `json:"tag"`
}

// PoliticalParty represents a generated political group.
type PoliticalParty struct {
	Name         string           `json:"name"`
	Issues       []PoliticalIssue `json:"issues"`
	Leadership   string           `json:"leadership"`
	Relationship string           `json:"relationship"`
	Policy       string           `json:"policy"`
}

// Architecture represents generated architectural features.
type Architecture struct {
	Name        string `json:"name"`
	Towers      string `json:"towers"`
	Foundations string `json:"foundations"`
	WallDecor   string `json:"wall_decor"`
	Supports    string `json:"supports"`
	Arches      string `json:"arches"`
	Extras      string `json:"extras"`
}

// Alien represents a generated alien race.
type Alien struct {
	Name   string       `json:"name"`
	Body   [][]string   `json:"body"`
	Lens   [][]string   `json:"lens"`
	Vars   []string     `json:"vars"`
	Social [][]string   `json:"social"`
	Arch   Architecture `json:"arch"`
}

// Sector holds all generated data for a sector.
type Sector struct {
	Name   string           `json:"name"`
	Token  string           `json:"token"`
	Map    interface{}      `json:"map"`
	Stars  []*Star          `json:"stars,omitempty"`
	Worlds []*World         `json:"worlds,omitempty"`
	NPCs   []*NPC           `json:"npcs,omitempty"`
	Corps  []*Corp          `json:"corps,omitempty"`
	Rels   []*Religion      `json:"rels,omitempty"`
	Pols   []*PoliticalParty `json:"pols,omitempty"`
	Aliens []*Alien         `json:"aliens,omitempty"`
}
