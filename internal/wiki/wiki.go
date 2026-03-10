package wiki

import (
	"archive/zip"
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"swn-go/internal/generator"
	"swn-go/internal/render"
	"swn-go/internal/util"
)

const gmColor = "FF0000"

// GenerateWiki creates TiddlyWiki output and returns a ZIP file as bytes.
// swnHTML is the TiddlyWiki template content.
func GenerateWiki(s *generator.Sector, swnHTML string, starMap []*generator.Star, worlds []*generator.World,
	bgPNG, dotPNG, fontData []byte, forIE bool) []byte {

	now := time.Now()
	timeStamp := fmt.Sprintf("%d%02d%02d%02d%02d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())

	sectorName := s.Name
	token := s.Token

	// Render map
	mapData, areas := render.RenderMap(sectorName, starMap, bgPNG, dotPNG, fontData, forIE)

	// Build wiki content
	smTxt := markupSectorMap(worlds, starMap)
	gmPdTxt := markupPlanetaryLists(worlds, starMap, false)
	pcPdTxt := markupPlanetaryLists(worlds, starMap, true)

	pcBkmk := strings.Join([]string{
		"----",
		"[[Alien Races]]",
		"[[Political Groups]]",
		"[[Corporations|Corporation Listing]]",
		"[[Religions|Sector Religions]]",
	}, "\n")

	gmBkmk := pcBkmk + "\n[[NPCs|NPC Listing]]"

	gmArtTxt := markupArticles(s, starMap, worlds, timeStamp, false)
	pcArtTxt := markupArticles(s, starMap, worlds, timeStamp, true)

	gmLink := "----\n[[GM Info]]"
	pcLink := "----"

	starImg := ""
	starURL := ""
	starFN := "SWN_wiki_" + token + ".png"
	if forIE {
		starURL = starFN
	} else {
		starImg = "data:image/png;base64," + string(mapData)
	}

	// Process template
	gmStr := swnHTML
	pcStr := swnHTML

	replacePairs := []struct{ placeholder, gmVal, pcVal string }{
		{"$$GMINFO", gmLink, pcLink},
		{"$$SEED_TOKEN", token, token},
		{"$$SECTOR_MAP", smTxt, smTxt},
		{"$$PLANETARY_DIRECTORY", gmPdTxt, pcPdTxt},
		{"$$PLANET_LIST", gmBkmk, pcBkmk},
		{"$$STARMAP", starImg, starImg},
		{"$$STARURL", starURL, starURL},
		{"$$MAP_LINKS", areas, areas},
		{"$$TIME_STAMP", timeStamp, timeStamp},
	}

	for _, rp := range replacePairs {
		gmStr = strings.ReplaceAll(gmStr, rp.placeholder, rp.gmVal)
		pcStr = strings.ReplaceAll(pcStr, rp.placeholder, rp.pcVal)
	}

	// Sector name can appear multiple times
	gmStr = strings.ReplaceAll(gmStr, "$$SECTOR_NAME", sectorName)
	pcStr = strings.ReplaceAll(pcStr, "$$SECTOR_NAME", sectorName)

	// Sector info articles
	gmStr = strings.ReplaceAll(gmStr, "<!--$$SECTOR_INFO-->", gmArtTxt)
	pcStr = strings.ReplaceAll(pcStr, "<!--$$SECTOR_INFO-->", pcArtTxt)

	// Create ZIP
	gmFN := "SWN_wiki_" + token + "_GM.html"
	pcFN := "SWN_wiki_" + token + "_PC.html"

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	addZipFile(w, "README.text", getReadme())
	addZipFile(w, gmFN, gmStr)
	addZipFile(w, pcFN, pcStr)
	if forIE {
		addZipBinary(w, starFN, mapData)
	}
	w.Close()

	return buf.Bytes()
}

func addZipFile(w *zip.Writer, name, content string) {
	f, _ := w.Create(name)
	f.Write([]byte(content))
}

func addZipBinary(w *zip.Writer, name string, data []byte) {
	f, _ := w.Create(name)
	f.Write(data)
}

func markupSectorMap(worlds []*generator.World, starMap []*generator.Star) string {
	wmap := make(map[int][]*generator.World)
	for _, w := range worlds {
		wmap[w.StarID] = append(wmap[w.StarID], w)
	}
	for _, wl := range wmap {
		sort.Slice(wl, func(i, j int) bool { return wl[i].SysPos < wl[j].SysPos })
	}

	txt := "|[img[StarMap]]<<imageMap StarMapLinks>>|!Hex|!World or Station|\n"

	sorted := make([]*generator.Star, len(starMap))
	copy(sorted, starMap)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Cell < sorted[j].Cell })

	for _, s := range sorted {
		sysname := util.SysName(s.Name, s.Cell)
		cellStr := "[[" + s.Cell + "|" + sysname + "]]"
		for _, w := range wmap[s.ID] {
			txt += "|~|" + cellStr + "|[[" + w.Name + "|Planet:" + w.Name + "]]|\n"
			cellStr = "~"
		}
	}

	return txt
}

func markupPlanetaryLists(worlds []*generator.World, starMap []*generator.Star, forPC bool) string {
	dirtxt := "|!Hex|!World or Station|!Atmo.|!Temp.|!Biosphere|!Population|!TL|"
	if !forPC {
		dirtxt += "!@@color:#" + gmColor + ";Tags@@|"
	}
	dirtxt += "h\n"

	sorted := make([]*generator.World, len(worlds))
	copy(sorted, worlds)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	for _, w := range sorted {
		s := starMap[w.StarID]
		sysname := util.SysName(s.Name, s.Cell)
		tags := "@@color:#" + gmColor + ";" + w.Tags[0].Short + ", " + w.Tags[1].Short + "@@"
		plname := "[[" + w.Name + "|Planet:" + w.Name + "]]"

		flds := []string{
			"[[" + s.Cell + "|" + sysname + "]]",
			plname,
			w.Atmosphere.Short,
			w.Temperature.Short,
			w.Biosphere.Short,
			w.Population.Short,
			w.TechLevel.Short,
		}
		if !forPC {
			flds = append(flds, tags)
		}
		dirtxt += "|" + strings.Join(flds, "|") + "|\n"
	}

	return dirtxt
}

func markupArticles(s *generator.Sector, starMap []*generator.Star, worlds []*generator.World,
	timeStamp string, forPC bool) string {

	var txt strings.Builder

	// Sort worlds by name
	sorted := make([]*generator.World, len(worlds))
	copy(sorted, worlds)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	systems := make(map[string]*systemInfo)
	atms := make(map[string]*tagInfo)
	temps := make(map[string]*tagInfo)
	bios := make(map[string]*tagInfo)
	pops := make(map[string]*tagInfo)
	tls := make(map[string]*tagInfo)
	tags := make(map[string]*tagInfo)

	// Planet articles
	for _, w := range sorted {
		star := starMap[w.StarID]
		sysname := util.SysName(star.Name, star.Cell)

		if _, ok := systems[sysname]; !ok {
			systems[sysname] = &systemInfo{star: star.Name, cell: star.Cell, planets: make(map[int]string)}
		}
		systems[sysname].planets[w.SysPos] = w.Name

		p := strings.Join([]string{
			"''//Nav Designation//: [[" + star.Name + "|" + sysname + "]] " + w.SysNum + "''",
			"|!Atmosphere |" + w.Atmosphere.Name + "|",
			"|!Temperature|" + w.Temperature.Name + "|",
			"|!Biosphere  |" + w.Biosphere.Name + "|",
			"|!Population |" + w.Population.Name + "|",
			"|!Tech Level |" + w.TechLevel.Name + "|",
		}, "\n")

		if !forPC {
			p += "\n" + strings.Join([]string{
				"|!@@color:#" + gmColor + ";Tags@@|@@color:#" + gmColor + ";" + w.Tags[0].Name + ", " + w.Tags[1].Name + "@@|",
				"!!!@@color:#" + gmColor + ";Enemies@@",
				"@@color:#" + gmColor + ";" + w.Tags[0].Enemies + "; " + w.Tags[1].Enemies + "@@",
				"!!!@@color:#" + gmColor + ";Friends@@",
				"@@color:#" + gmColor + ";" + w.Tags[0].Friends + "; " + w.Tags[1].Friends + "@@",
				"!!!@@color:#" + gmColor + ";Complications@@",
				"@@color:#" + gmColor + ";" + w.Tags[0].Complications + "; " + w.Tags[1].Complications + "@@",
				"!!!@@color:#" + gmColor + ";Places@@",
				"@@color:#" + gmColor + ";" + w.Tags[0].Places + "; " + w.Tags[1].Places + "@@",
				"!!!@@color:#" + gmColor + ";Things@@",
				"@@color:#" + gmColor + ";" + w.Tags[0].Things + "; " + w.Tags[1].Things + "@@",
				"!!@@color:#" + gmColor + ";Capital and Government@@",
				"!!@@color:#" + gmColor + ";Cultural Notes@@",
				"!!@@color:#" + gmColor + ";Adventures Prepared@@",
				"!!@@color:#" + gmColor + ";Party Activities on this World@@",
			}, "\n")
		} else {
			p += "\n!!!Notes\n//Add your own notes here//\n"
		}

		atag := "Atmosphere:" + util.Tagify(w.Atmosphere.Short)
		ttag := "Temperature:" + util.Tagify(w.Temperature.Short)
		btag := "Biosphere:" + util.Tagify(w.Biosphere.Short)
		ptag := "Population:" + w.Population.Short
		ltag := "TechLevel:" + w.TechLevel.Short
		tag1 := "Tag:" + util.Tagify(w.Tags[0].Short)
		tag2 := "Tag:" + util.Tagify(w.Tags[1].Short)

		if _, ok := atms[atag]; !ok {
			atms[atag] = &tagInfo{name: w.Atmosphere.Name, desc: w.Atmosphere.Desc}
		}
		if _, ok := temps[ttag]; !ok {
			temps[ttag] = &tagInfo{name: w.Temperature.Name, desc: w.Temperature.Desc}
		}
		if _, ok := bios[btag]; !ok {
			bios[btag] = &tagInfo{name: w.Biosphere.Name, desc: w.Biosphere.Desc}
		}
		if _, ok := pops[ptag]; !ok {
			pops[ptag] = &tagInfo{name: w.Population.Name, desc: w.Population.Desc}
		}
		if _, ok := tls[ltag]; !ok {
			tls[ltag] = &tagInfo{name: w.TechLevel.Name, desc: w.TechLevel.Desc}
		}
		if _, ok := tags[tag1]; !ok {
			tags[tag1] = &tagInfo{name: w.Tags[0].Name, desc: w.Tags[0].Desc}
		}
		if _, ok := tags[tag2]; !ok {
			tags[tag2] = &tagInfo{name: w.Tags[1].Name, desc: w.Tags[1].Desc}
		}

		allTags := []string{"System:" + strings.ToUpper(starMap[w.StarID].Name), "Planet:" + w.Name, atag, ttag, btag, ptag, ltag}
		if !forPC {
			allTags = append(allTags, tag1, tag2)
		}

		txt.WriteString(divOpen("Planet:"+w.Name, timeStamp, strings.Join(allTags, " ")))
		txt.WriteString("<pre>" + p + "</pre>\n")
		txt.WriteString("</div>\n")
	}

	// System articles
	sysKeys := sortedKeys(systems)
	for _, skey := range sysKeys {
		sys := systems[skey]
		posKeys := sortedIntKeys(sys.planets)
		var planetTags []string
		var links []string
		planetTags = append(planetTags, "System:"+strings.ToUpper(sys.star))
		for _, pk := range posKeys {
			planetTags = append(planetTags, "Planet:"+sys.planets[pk])
			links = append(links, "[["+sys.planets[pk]+"|Planet:"+sys.planets[pk]+"]]")
		}

		txt.WriteString(divOpen(skey, timeStamp, strings.Join(planetTags, " ")))
		txt.WriteString("<pre>|!Nav Designation|''" + "GRID " + sys.cell + "''|\n")
		tl := "!System Planet"
		if len(links) > 1 {
			tl += "s"
		}
		for li, link := range links {
			c1 := tl
			if li > 0 {
				c1 = "~"
			}
			txt.WriteString("|" + c1 + "|" + link + "|\n")
		}
		txt.WriteString("</pre>\n</div>\n")
	}

	// Tag articles
	tagLists := []map[string]*tagInfo{atms, temps, bios, pops, tls}
	if !forPC {
		tagLists = append(tagLists, tags)
	}
	for _, list := range tagLists {
		for tag, info := range list {
			txt.WriteString(divOpenNoTags(tag, timeStamp))
			txt.WriteString("<pre>''" + info.name + "'' - " + info.desc + "\n</pre>\n")
			txt.WriteString("</div>\n")
		}
	}

	// Corporation articles
	if s.Corps != nil {
		writeCorpArticles(&txt, s, timeStamp, forPC)
	}

	// Religion articles
	if s.Rels != nil {
		writeRelArticles(&txt, s, timeStamp, forPC)
	}

	// Political articles
	if s.Pols != nil {
		writePolArticles(&txt, s, timeStamp, forPC)
	}

	// Alien articles
	if s.Aliens != nil {
		writeAlienArticles(&txt, s, timeStamp, forPC)
	}

	// NPC articles (GM only)
	if !forPC && s.NPCs != nil {
		writeNPCArticles(&txt, s, timeStamp)
	}

	// GM Info article
	if !forPC {
		writeGMInfo(&txt, timeStamp)
	}

	return txt.String()
}

func writeCorpArticles(txt *strings.Builder, s *generator.Sector, ts string, forPC bool) {
	corps := make([]*generator.Corp, len(s.Corps))
	copy(corps, s.Corps)
	sort.Slice(corps, func(i, j int) bool { return corps[i].Name < corps[j].Name })

	var corpArts []string
	corpTags := make(map[string]string)

	txt.WriteString(divOpenNoTags("Corporation Listing", ts))
	txt.WriteString("<pre>\n|!Company Name|!Business|h\n")

	for _, corp := range corps {
		title := "Corp:" + util.Tagify(corp.Name)
		biztag := "Business:" + util.Tagify(corp.Business)

		if _, ok := corpTags[biztag]; !ok {
			corpTags[biztag] = divOpenNoTags(biztag, ts) + "<pre></pre>\n</div>"
		}

		txt.WriteString("|[[" + corp.Name + "|" + title + "]]|" + corp.Business + "|\n")

		art := divOpen(title, ts, title+" "+biztag) +
			"<pre>\n|!Name      |" + corp.Name + "|\n|!Business  |" + corp.Business + "|\n"

		if !forPC {
			art += "|!@@color:#" + gmColor + ";Reputation@@|@@color:#" + gmColor + ";" + corp.Reputation + "@@|\n"
		}
		art += "\n!!!Notes\n</pre>\n</div>"
		corpArts = append(corpArts, art)
	}
	txt.WriteString("</pre>\n</div>\n")
	for _, a := range corpArts {
		txt.WriteString(a + "\n")
	}
	for _, t := range corpTags {
		txt.WriteString(t + "\n")
	}
}

func writeRelArticles(txt *strings.Builder, s *generator.Sector, ts string, forPC bool) {
	rels := make([]*generator.Religion, len(s.Rels))
	copy(rels, s.Rels)
	sort.Slice(rels, func(i, j int) bool { return rels[i].Origin < rels[j].Origin })

	var relArts, herArts []string
	relOTags := make(map[string]string)

	txt.WriteString(divOpenNoTags("Sector Religions", ts))
	txt.WriteString("<pre>\n|!Religion|!Origin|!Leadership|h\n")

	for _, rel := range rels {
		title := "Religion:" + util.Tagify(rel.Name)
		otag := "Origin:" + util.Tagify(rel.Origin)
		ltype := strings.SplitN(rel.Leadership, ".", 2)[0]

		txt.WriteString("|[[" + rel.Name + "|" + title + "]]|" + rel.Origin + "|" + ltype + "|\n")

		rtags := title + " " + otag

		if _, ok := relOTags[otag]; !ok {
			relOTags[otag] = divOpenNoTags(otag, ts) + "<pre></pre>\n</div>"
		}

		hlinks := ""
		if !forPC {
			sortedHers := make([]generator.Heresy, len(rel.Offshoots))
			copy(sortedHers, rel.Offshoots)
			sort.Slice(sortedHers, func(i, j int) bool { return sortedHers[i].Name < sortedHers[j].Name })

			for _, her := range sortedHers {
				htag := "Heresy:" + util.Tagify(her.Name)
				ftype := strings.SplitN(her.Founder, ":", 2)[0]
				hlinks += "|[[" + her.Name + "|" + htag + "]]|" + her.Quirk + "|" + ftype + "|\n"
				rtags += " " + htag

				ha := divOpen(htag, ts, title) +
					"<pre>\n|!Name   |" + her.Name + "|\n|!Origin |[[" + rel.Name + "|" + title + "]]|\n|!Founder|" + her.Founder + "|\n|!Quirk  |" + her.Quirk + "|\n!!!Major Heresy\n" + her.HerDesc + "\n!!!Attitude towards Orthodoxy\n" + her.Attitude + "\n</pre>\n</div>"
				herArts = append(herArts, ha)
			}
		}

		art := divOpen(title, ts, rtags) +
			"<pre>\n|!Name      |" + rel.Name + "|\n|!Origin    |" + rel.Origin + "|\n!!!Leadership\n" + rel.Leadership + "\n!!!Evolution\n" + rel.Evolution + "\n"

		if !forPC {
			art += "!!!@@color:#" + gmColor + ";Heresies@@\n|!Name|!Quirk|!Founder|h\n" + hlinks
		}
		art += "!!!Notes\n</pre>\n</div>"
		relArts = append(relArts, art)
	}
	txt.WriteString("</pre>\n</div>\n")
	for _, a := range relArts {
		txt.WriteString(a + "\n")
	}
	for _, t := range relOTags {
		txt.WriteString(t + "\n")
	}
	if !forPC {
		for _, h := range herArts {
			txt.WriteString(h + "\n")
		}
	}
}

func writePolArticles(txt *strings.Builder, s *generator.Sector, ts string, forPC bool) {
	pols := make([]*generator.PoliticalParty, len(s.Pols))
	copy(pols, s.Pols)
	sort.Slice(pols, func(i, j int) bool { return pols[i].Name < pols[j].Name })

	var polArts []string
	issTags := make(map[string]string)
	pcyTags := make(map[string]string)

	txt.WriteString(divOpenNoTags("Political Groups", ts))
	txt.WriteString("<pre>\n|!Organization|!Leadership|!Policy|!Outsiders|!Issues|h\n")

	for _, pol := range pols {
		title := "Group:" + util.Tagify(pol.Name)
		ptype := strings.SplitN(pol.Policy, ":", 2)[0]
		ltype := strings.SplitN(pol.Leadership, ":", 2)[0]
		otype := strings.SplitN(pol.Relationship, ":", 2)[0]
		ptag := "Ideology:" + util.Tagify(ptype)

		polTags := title + " " + ptag

		if _, ok := pcyTags[ptag]; !ok {
			pdesc := ""
			parts := strings.SplitN(pol.Policy, ":", 2)
			if len(parts) > 1 {
				pdesc = parts[1]
			}
			pcyTags[ptag] = divOpenNoTags(ptag, ts) + "<pre>''" + ptype + "'': " + pdesc + "</pre>\n</div>"
		}

		itbltxt := ""
		itagtxt := ""
		for _, iss := range pol.Issues {
			if itbltxt == "" {
				itbltxt = iss.Tag
			} else {
				itbltxt += ", " + iss.Tag
			}
			itag := "Issue:" + util.Tagify(iss.Tag)
			itagtxt += "* " + iss.Issue + "\n"
			polTags += " " + itag
			if _, ok := issTags[itag]; !ok {
				issTags[itag] = divOpenNoTags(itag, ts) + "<pre>" + iss.Issue + "</pre>\n</div>"
			}
		}

		txt.WriteString("|[[" + pol.Name + "|" + title + "]]|" + ltype + "|" + ptype + "|" + otype + "|" + itbltxt + "|\n")

		art := divOpen(title, ts, polTags) +
			"<pre>\n|!Name      |" + pol.Name + "|\n!!!Ideology\n" + pol.Policy + "\n!!!Leadership\n" + pol.Leadership + "\n!!!Key Issues\n" + itagtxt + "\n!!!Relationship toward Outsiders\n" + pol.Relationship + "\n!!!Notes\n</pre>\n</div>"
		polArts = append(polArts, art)
	}
	txt.WriteString("</pre>\n</div>\n")
	for _, a := range polArts {
		txt.WriteString(a + "\n")
	}
	for _, t := range pcyTags {
		txt.WriteString(t + "\n")
	}
	for _, t := range issTags {
		txt.WriteString(t + "\n")
	}
}

func writeAlienArticles(txt *strings.Builder, s *generator.Sector, ts string, forPC bool) {
	aliens := make([]*generator.Alien, len(s.Aliens))
	copy(aliens, s.Aliens)
	sort.Slice(aliens, func(i, j int) bool { return aliens[i].Name < aliens[j].Name })

	var alienArts []string
	bodyTags := make(map[string]string)
	lensTags := make(map[string]string)
	socTags := make(map[string]string)

	txt.WriteString(divOpenNoTags("Alien Races", ts))
	txt.WriteString("<pre>\n|!Name|!Body Type|")
	if !forPC {
		txt.WriteString("!@@color:#" + gmColor + ";Lenses@@|!@@color:#" + gmColor + ";Structure@@|")
	}
	txt.WriteString("h\n")

	for _, aln := range aliens {
		title := "Alien:" + util.Tagify(aln.Name)
		alnTags := title

		var btag, bname, bdesc, btxt, bpar string
		if len(aln.Body) > 1 {
			btag = "Body:Hybrid"
			bname = "Hybrid"
			btxt = "Hybrid: " + aln.Body[0][0] + " and " + aln.Body[1][0]
			bpar = "''" + aln.Body[0][0] + "'': " + aln.Body[0][1] + "&lt;br&gt;&lt;br&gt;''" + aln.Body[1][0] + "'': " + aln.Body[1][1]
		} else {
			btag = "Body:" + aln.Body[0][0]
			bname = aln.Body[0][0]
			bdesc = ": " + aln.Body[0][1]
			btxt = aln.Body[0][0]
			bpar = "''" + aln.Body[0][0] + "'': " + aln.Body[0][1]
		}
		alnTags += " " + btag

		if _, ok := bodyTags[btag]; !ok {
			bodyTags[btag] = divOpenNoTags(btag, ts) + "<pre>''" + bname + "''" + bdesc + "</pre>\n</div>"
		}

		var ltxt, lpar string
		for _, lens := range aln.Lens {
			lname := lens[0]
			ldesc := lens[1]
			ltag := "Lens:" + lname
			if !forPC {
				alnTags += " " + ltag
			}
			if ltxt == "" {
				ltxt = lname
				lpar = "''" + lname + "'': " + ldesc
			} else {
				ltxt += ", " + lname
				lpar += "&lt;br&gt;&lt;br&gt;''" + lname + "'': " + ldesc
			}
			if _, ok := lensTags[ltag]; !ok {
				lensTags[ltag] = divOpenNoTags(ltag, ts) + "<pre>''" + lname + "''" + ldesc + "</pre>\n</div>"
			}
		}

		var spar string
		structs := make(map[string]bool)
		for _, soc := range aln.Social {
			sname := soc[0]
			sdesc := soc[1]
			stag := "Structure:" + sname
			if !forPC {
				alnTags += " " + stag
			}
			structs[sname] = true
			if spar == "" {
				spar = "''" + sname + "'': " + sdesc
			} else {
				spar += "&lt;br&gt;&lt;br&gt;''" + sname + "'': " + sdesc
			}
			if _, ok := socTags[stag]; !ok {
				socTags[stag] = divOpenNoTags(stag, ts) + "<pre>''" + sname + "'': " + sdesc + "</pre>\n</div>"
			}
		}
		var structNames []string
		for k := range structs {
			structNames = append(structNames, k)
		}
		stxt := strings.Join(structNames, ", ")

		txt.WriteString("|[[" + aln.Name + "|" + title + "]]|" + btxt + "|")
		if !forPC {
			txt.WriteString("@@color:#" + gmColor + ";" + ltxt + "@@|@@color:#" + gmColor + ";" + stxt + "@@|")
		}
		txt.WriteString("\n")

		art := divOpen(title, ts, alnTags) +
			"<pre>\n|!Name      |" + aln.Name + "|\n|!Body Type |" + btxt + "|\n"
		if !forPC {
			art += "|!@@color:#" + gmColor + ";Lenses@@|@@color:#" + gmColor + ";" + ltxt + "@@|\n|!@@color:#" + gmColor + ";Structure@@|@@color:#" + gmColor + ";" + stxt + "@@|\n"
		}
		art += "!!!Distinguishing Features\n"
		if len(aln.Vars) > 0 {
			art += "* " + aln.Vars[0] + "\n"
		}
		if len(aln.Vars) > 1 {
			art += "* " + aln.Vars[1] + "\n"
		}
		if !forPC {
			art += "!!!@@color:#" + gmColor + ";Body Type@@\n@@color:#" + gmColor + ";" + bpar + "@@\n" +
				"!!!@@color:#" + gmColor + ";Lenses@@\n@@color:#" + gmColor + ";" + lpar + "@@\n" +
				"!!!@@color:#" + gmColor + ";Social Structure@@\n@@color:#" + gmColor + ";" + spar + "@@\n" +
				"!!!@@color:#" + gmColor + ";Architecture@@\n" +
				"* @@color:#" + gmColor + ";" + aln.Arch.Towers + "@@\n" +
				"* @@color:#" + gmColor + ";" + aln.Arch.Foundations + "@@\n" +
				"* @@color:#" + gmColor + ";" + aln.Arch.WallDecor + "@@\n" +
				"* @@color:#" + gmColor + ";" + aln.Arch.Supports + "@@\n" +
				"* @@color:#" + gmColor + ";" + aln.Arch.Arches + "@@\n" +
				"* @@color:#" + gmColor + ";" + aln.Arch.Extras + "@@\n"
		}
		art += "!!!Notes\n</pre>\n</div>"
		alienArts = append(alienArts, art)
	}
	txt.WriteString("</pre>\n</div>\n")
	for _, a := range alienArts {
		txt.WriteString(a + "\n")
	}
	for _, t := range bodyTags {
		txt.WriteString(t + "\n")
	}
	if !forPC {
		for _, t := range lensTags {
			txt.WriteString(t + "\n")
		}
		for _, t := range socTags {
			txt.WriteString(t + "\n")
		}
	}
}

func writeNPCArticles(txt *strings.Builder, s *generator.Sector, ts string) {
	npcs := make([]*generator.NPC, len(s.NPCs))
	copy(npcs, s.NPCs)
	sort.Slice(npcs, func(i, j int) bool { return npcs[i].Sort < npcs[j].Sort })

	var npcArts []string
	npcSex := make(map[string]bool)
	npcAge := make(map[string]bool)
	npcHgt := make(map[string]bool)

	txt.WriteString(divOpenNoTags("NPC Listing", ts))
	txt.WriteString("<pre>\n|!Name|!Gender|!Age|!Height|h\n")

	for _, npc := range npcs {
		title := "NPC:" + util.Tagify(npc.Name)
		txt.WriteString("|[[" + npc.Name + "|" + title + "]]|" + npc.Gender + "|" + npc.Age + "|" + npc.Height + "|\n")

		stag := "Gender:" + npc.Gender
		atag := "Age:" + npc.Age
		htag := "Height:" + util.Tagify(npc.Height)
		npcSex[stag] = true
		npcAge[atag] = true
		npcHgt[htag] = true

		npcTags := strings.Join([]string{title, stag, atag, htag}, " ")
		art := divOpen(title, ts, npcTags) +
			"<pre>\n|!Name  |" + npc.Name + "|\n|!Gender|" + npc.Gender + "|\n|!Age   |" + npc.Age + "|\n|!Height|" + npc.Height + "|\n|!Quirk |" + npc.Quirk + "|\n!!!Motive\n" + npc.Motive + "\n!!!Problem\n" + npc.Problem + "\n!!!Notes\n</pre>\n</div>"
		npcArts = append(npcArts, art)
	}
	txt.WriteString("</pre>\n</div>\n")
	for _, a := range npcArts {
		txt.WriteString(a + "\n")
	}
	for tag := range npcSex {
		txt.WriteString(divOpenNoTags(tag, ts) + "<pre></pre>\n</div>\n")
	}
	for tag := range npcAge {
		txt.WriteString(divOpenNoTags(tag, ts) + "<pre></pre>\n</div>\n")
	}
	for tag := range npcHgt {
		txt.WriteString(divOpenNoTags(tag, ts) + "<pre></pre>\n</div>\n")
	}
}

func writeGMInfo(txt *strings.Builder, ts string) {
	txt.WriteString(divOpenNoTags("GM Info", ts))
	txt.WriteString(`<pre>
!!General Stuff to Know
A couple of quick notes on the differences between the GM and PC versions of this file:
* The PC version does not contain //any// info on:
** ~NPCs
** Heretical groups
* The PC version contains //limited// info on:
** Planets
** Corporations
** Alien Races
Otherwise, the PC version of the file is identical to the GM version. Where possible, "~GM-only" information in this document is highlighted in @@color:#` + gmColor + `;this color@@. Sometimes only a header is highlighted; the paragraphs within the section can be treated as ~GM-only knowledge.

You can alter the level of information that ~PCs have by editing the PC copy of the file before handing it out. Or, you can skip handing out the PC version entirely. I attempted to limit PC info to what I thought the characters could glean from the post-Scream version of the Internet, but the very existence of a post-Scream Internet in your sector is really up to you.

!!Other Comments
While you could start a campaign without making any changes to the contents of this file, you should be aware of a few things:

''1. Alien Races''
The generator always rolls up ten alien races and generates a distinct set of architectural features for each race. That presumes that every alien race in the sector is sentient and feels inclined to build structures, which may not make sense. Look at the alien races, decide which ones are sentient participants in the overall universe and adjust the articles for each race accordingly.

''2. Corporations, Political Parties, ~NPCs''
A sizeable number of each of these has been generated for your convenience. However, they are not linked to any of the planets or to each other. If you want J. Random NPC to be the president of ACME Corp, based on planet Zanzibar, then a brief introduction to the extremely simple-to-use ~TiddlyWiki markup is in order.

''3. Religions and Heresies''
By design, the generator very frequently produces hitherto unknown intepretations of known world religions. If you are put off by a religion that is generated by this utility, by all means, edit it in your copies of the sector wiki. Regarding heresies; the generator rolls up 1-3 heretical offshoots for every generated religion. Heresies don't get their own listing page because they are listed on the page of their "parent religion".

As an aside, this article is also not included in the PC version of the file. Happy gaming!
</pre>
</div>
`)
}

// Helper types and functions

type systemInfo struct {
	star    string
	cell    string
	planets map[int]string
}

type tagInfo struct {
	name string
	desc string
}

func divOpen(title, ts, tags string) string {
	return `<div title="` + title + `" creator="SWN Sector Generator" modifier="SWN Sector Generator" created="` + ts + `" tags="` + tags + `" changecount="1">` + "\n"
}

func divOpenNoTags(title, ts string) string {
	return `<div title="` + title + `" creator="SWN Sector Generator" modifier="SWN Sector Generator" created="` + ts + `" changecount="1">` + "\n"
}

func sortedKeys(m map[string]*systemInfo) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedIntKeys(m map[int]string) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func getReadme() string {
	return `SWN Sector Generator
===================

A utility for use with Stars Without Number (http://www.sinenomine-pub.com/).


## Contents
The zip file that you've unpacked contains the following items:

* This README file
* A 'GM version' of the sector that you generated
* A 'PC version' of the sector that you generated
* A star map image file (if you generated the sector in Internet Explorer)


## How to Use
Open the GM version of the sector wiki in your browser and have a look around.
In particular, check out the 'GM Info' article. It will help you to use these
pre-generated sectors to their best effect. If you aren't familiar with how
TiddlyWiki works, check out the excellent documentation at the TiddlyWiki
website: http://www.tiddlywiki.com/


## Credits
The SWN Sector Generator
Copyright 2011 N. Harrison Ripps

This utility is based on and written for use with the Stars Without Number
Science Fiction Roleplaying Game by Sine Nomine Publishing
`
}
