package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path"

	"inspector/util"
)

var (
	term          = flag.Int("term", 10, "legislator term")
	InputDir      = flag.String("d", ".", "legislator term")
	workerCount   = flag.Int("w", 20, "number of workers")
	outputDir     = flag.String("o", "./dist", "output folder")
	startingIndex = flag.Int("i", 0, "start from")
)

func main() {
	flag.Parse()
	legs, err := util.GetLegIds(*term)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		os.Mkdir(*outputDir, 0755)
	}

	for i := *startingIndex; i < len(legs); i++ {
		leg := legs[i]
		log.Printf("%s: start fetching videos", leg.Name)

		f, err := os.Create(path.Join(*outputDir, leg.Name+".json"))
		if err != nil {
			log.Fatal(err)
		}

		leg.Videos = util.GetLegVideos(leg.Id, *workerCount)
		jsonBytes, err := json.Marshal(leg)
		if err != nil {
			log.Fatal(err)
		}

		n, err := f.WriteString(string(jsonBytes))
		if err != nil {
			log.Fatal(err)
		} else if n != len(jsonBytes) {
			log.Fatal("write error")
		}

		f.Close()
		log.Printf("%s: done fetching videos\n", leg.Name)
	}

}
