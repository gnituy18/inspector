package util

import (
	"encoding/json"
	"os"
	"path"
	"strconv"
)

func GetLegNames(term int) ([]string, error) {
	data, err := ReadData(path.Join("leg-" + strconv.Itoa(term) + "-name-list" + ".json"))
	if err != nil && err != os.ErrNotExist {
		return nil, err
	} else if err == os.ErrNotExist {
		// TODO fetch from the web
		return nil, nil
	}

	var names []string
	if json.Unmarshal(data, &names); err != nil {
		return nil, err
	}

	return names, nil
}
