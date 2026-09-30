package models

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Song struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	Channel   string      `json:"channel"`
	Duration  float64     `json:"duration"`
	LocalPath string      `json:"local_path,omitempty"`
	Favorite  bool        `json:"favorite"`
	AddedAt   string      `json:"added_at"`
	Lyrics    []LyricLine `json:"lyrics,omitempty"`
	LastPos   float64     `json:"last_pos,omitempty"`
}

type Playlist struct {
	Name    string   `json:"name"`
	SongIDs []string `json:"song_ids"`
}

type HistoryEntry struct {
	Song     Song      `json:"song"`
	PlayedAt time.Time `json:"played_at"`
}

type LastFMConfig struct {
	APIKey     string `json:"api_key"`
	APISecret  string `json:"api_secret"`
	SessionKey string `json:"session_key"`
	Username   string `json:"username"`
	Enabled    bool   `json:"enabled"`
}

type Library struct {
	Songs     []Song         `json:"songs"`
	Playlists []Playlist     `json:"playlists"`
	History   []HistoryEntry `json:"history"`
	LastFM    LastFMConfig   `json:"lastfm"`
}

type LyricLine struct {
	Time float64
	Text string
}

type YTResult struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Channel  string  `json:"channel"`
	Duration float64 `json:"duration"`
}

func FormatDuration(seconds float64) string {
	if seconds <= 0 {
		return "0:00"
	}
	total := int(seconds)
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

var (
	MusicDir  = filepath.Join(os.Getenv("HOME"), "storage", "music", "termux-music")
	ThumbDir  = filepath.Join(MusicDir, ".thumbs")
	DbFile    = filepath.Join(MusicDir, "library.json")
	MpvSocket = "/data/data/com.termux/files/usr/tmp/termux-music.sock"
)
