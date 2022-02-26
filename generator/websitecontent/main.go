package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"path"

	"inspector/model"
	"inspector/util"
)

var (
	term int
)

func init() {
	flag.IntVar(&term, "term", 0, "legislator term")
}

func main() {
	flag.Parse()
	// TODO: should get all the politicians instead of term 10 legs
	names, err := util.GetLegNames(term)
	if err != nil {
		log.Fatal(err)
	}

	for _, name := range names {
		data, err := util.ReadData(path.Join("legipvid", name+".json"))
		if err != nil {
			log.Fatal(err)
		}

		leg := &model.Legislator{}
		json.Unmarshal(data, leg)

		file, err := util.CreateData(path.Join("../website/content", leg.Name+".html"))
		if err != nil {
			log.Fatal(err)
		}


		file.WriteString("---\n")
		file.WriteString("title: " + leg.Name + "\n")
		file.WriteString("ipvids:\n")
		for _, vid := range leg.Videos {
			file.WriteString("  -\n")
			file.WriteString(fmt.Sprintf("    id: %d\n", vid.Id))
			file.WriteString(fmt.Sprintf("    url: %s\n", vid.PlaylistUrl))
		}
		file.WriteString("---\n\n")

		file.WriteString("<div>\n")
		for _, vid := range leg.Videos {
			title := fmt.Sprintf("第%d屆 第%d會期 主辦單位：%s", vid.Term, vid.Session, vid.Committee)
			file.WriteString(fmt.Sprintf("{{< legipvid id=\"%d\" title=\"%s\">}}\n", vid.Id, title))
		}
		file.WriteString("</div>\n")
		file.Close()
	}
}
