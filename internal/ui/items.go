package ui

import (
	"fmt"

	"github.com/rugved-danej/termux-music/internal/models"
)

type listItem struct {
	song models.Song
}

func (i listItem) Title() string {
	t := i.song.Title
	if i.song.Favorite {
		t += " [fav]"
	}
	if i.song.LocalPath != "" {
		t += " [dl]"
	}
	return t
}

func (i listItem) Description() string {
	return fmt.Sprintf("%s  %s", i.song.Channel, models.FormatDuration(i.song.Duration))
}

func (i listItem) FilterValue() string {
	return i.song.Title + " " + i.song.Channel
}

type playlistItem struct {
	pl models.Playlist
}

func (i playlistItem) Title() string { return i.pl.Name }
func (i playlistItem) Description() string {
	n := len(i.pl.SongIDs)
	if n == 1 {
		return "1 song"
	}
	return fmt.Sprintf("%d songs", n)
}
func (i playlistItem) FilterValue() string { return i.pl.Name }

type dlOptionItem struct {
	title string
	desc  string
	args  []string
	ext   string
}

func (i dlOptionItem) Title() string       { return i.title }
func (i dlOptionItem) Description() string { return i.desc }
func (i dlOptionItem) FilterValue() string { return i.title }
