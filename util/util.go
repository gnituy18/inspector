package util

import (
	"inspector/config"
	"io/ioutil"
	"os"
	"path"
)

func ReadData(filepath string) ([]byte, error) {
	dataDir := config.Get().DataDir()
	return ioutil.ReadFile(path.Join(dataDir, filepath))
}

func CreateData(filepath string) (*os.File, error) {
	dataDir := config.Get().DataDir()
	if err := os.Truncate(path.Join(dataDir, filepath), 0); err != nil {
		return nil, err
	}
	return os.OpenFile(path.Join(dataDir, filepath), os.O_RDWR|os.O_CREATE, 0644)
}
