package core

import (
	"encoding/json"
	"os"

	"github.com/rugved-danej/termux-music/internal/models"
)

func LoadLibrary() models.Library {
	os.MkdirAll(models.MusicDir, 0755)
	data, err := os.ReadFile(models.DbFile)
	if err != nil {
		return models.Library{}
	}
	var lib models.Library
	json.Unmarshal(data, &lib)
	return lib
}

func SaveLibrary(lib models.Library) {
	data, _ := json.MarshalIndent(lib, "", "  ")
	os.WriteFile(models.DbFile, data, 0644)
}
