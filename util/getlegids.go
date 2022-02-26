package util

import (
	"encoding/json"
	"os"
	"path"
	"strconv"

	"inspector/config"
	"inspector/model"
)

func GetLegIds(term int) ([]model.Legislator, error) {
	dataDir := config.Get().DataDir()
	file, err := os.Open(path.Join(dataDir, "legislators-"+strconv.Itoa(term)+".json"))
	if err != nil {
		// TODO if not exist fetch the list online
		return nil, err
	}
	defer file.Close()

	var legislators []model.Legislator
	err = json.NewDecoder(file).Decode(&legislators)
	if err != nil {
		return nil, err
	}

	return legislators, nil
}
