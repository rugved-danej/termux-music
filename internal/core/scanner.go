package core

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rugved-danej/termux-music/internal/models"
)

func ImportLocalFiles(lib *models.Library) int {
	dirsToScan := []string{
		filepath.Join(os.Getenv("HOME"), "Music"),
		filepath.Join(os.Getenv("HOME"), "storage", "music"),
		filepath.Join(os.Getenv("HOME"), "storage", "downloads"),
	}
	added := 0

	for _, baseDir := range dirsToScan {
		filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				if info.Name() == ".thumbs" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".mp3" || ext == ".m4a" || ext == ".flac" || ext == ".wav" {
				found := false
				for _, s := range lib.Songs {
					if s.LocalPath == path {
						found = true
						break
					}
				}
				if !found {
					hash := fmt.Sprintf("%x", md5.Sum([]byte(path)))
					id := "local_" + hash[:12]
					name := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
					song := models.Song{
						ID:        id,
						Title:     name,
						Channel:   "Local Audio",
						Duration:  0,
						LocalPath: path,
						AddedAt:   time.Now().Format(time.RFC3339),
					}
					lib.Upsert(song)
					added++
				}
			}
			return nil
		})
	}
	return added
}
