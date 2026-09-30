package core

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rugved-danej/termux-music/internal/models"
)

var TrendingPlaylistURL = "https://music.youtube.com/playlist?list=PLHMlQVzOBYoQEMsFV-c8XYcUqLtVK9M-f"

func ParseYTDLPOutput(out []byte) []models.Song {
	var songs []models.Song
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		var r models.YTResult
		if json.Unmarshal([]byte(line), &r) == nil && r.ID != "" && r.Title != "" {
			songs = append(songs, models.Song{
				ID:       r.ID,
				Title:    r.Title,
				Channel:  r.Channel,
				Duration: r.Duration,
				AddedAt:  time.Now().Format(time.RFC3339),
			})
		}
	}
	return songs
}

func YtdlpSearch(query string, limit int) []models.Song {
	cmd := exec.Command("yt-dlp", "-j", "--flat-playlist",
		fmt.Sprintf("ytsearch%d:%s", limit, query))
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return nil
	}
	return ParseYTDLPOutput(out)
}

func FetchLyrics(title, artist string) []models.LyricLine {
	q := url.QueryEscape(title)
	a := url.QueryEscape(artist)
	resp, err := http.Get(
		fmt.Sprintf("https://lrclib.net/api/search?q=%s&artist_name=%s", q, a))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var results []struct {
		SyncedLyrics string `json:"syncedLyrics"`
	}
	json.NewDecoder(resp.Body).Decode(&results)
	if len(results) == 0 || results[0].SyncedLyrics == "" {
		return nil
	}
	var lines []models.LyricLine
	for _, raw := range strings.Split(results[0].SyncedLyrics, "\n") {
		raw = strings.TrimSpace(raw)
		if !strings.HasPrefix(raw, "[") {
			continue
		}
		end := strings.Index(raw, "]")
		if end < 0 {
			continue
		}
		timeStr := raw[1:end]
		text := strings.TrimSpace(raw[end+1:])
		parts := strings.Split(timeStr, ":")
		if len(parts) != 2 {
			continue
		}
		mins, _ := strconv.ParseFloat(parts[0], 64)
		secs, _ := strconv.ParseFloat(parts[1], 64)
		lines = append(lines, models.LyricLine{Time: mins*60 + secs, Text: text})
	}
	return lines
}

func lastfmSign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var s strings.Builder
	for _, k := range keys {
		s.WriteString(k)
		s.WriteString(params[k])
	}
	s.WriteString(secret)
	return fmt.Sprintf("%x", md5.Sum([]byte(s.String())))
}

func lastfmPost(params map[string]string) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	http.PostForm("http://ws.audioscrobbler.com/2.0/", form)
}

func splitArtistTrack(song models.Song) (artist, track string) {
	parts := strings.SplitN(song.Title, " - ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return song.Channel, song.Title
}

func LastfmNowPlaying(song models.Song, cfg models.LastFMConfig) {
	if !cfg.Enabled || cfg.SessionKey == "" {
		return
	}
	artist, track := splitArtistTrack(song)
	params := map[string]string{
		"method":   "track.updateNowPlaying",
		"api_key":  cfg.APIKey,
		"sk":       cfg.SessionKey,
		"artist":   artist,
		"track":    track,
		"duration": strconv.Itoa(int(song.Duration)),
	}
	params["api_sig"] = lastfmSign(params, cfg.APISecret)
	params["format"] = "json"
	go lastfmPost(params)
}

func LastfmScrobble(song models.Song, cfg models.LastFMConfig) {
	if !cfg.Enabled || cfg.SessionKey == "" {
		return
	}
	artist, track := splitArtistTrack(song)
	params := map[string]string{
		"method":    "track.scrobble",
		"api_key":   cfg.APIKey,
		"sk":        cfg.SessionKey,
		"artist":    artist,
		"track":     track,
		"timestamp": strconv.FormatInt(time.Now().Unix(), 10),
	}
	params["api_sig"] = lastfmSign(params, cfg.APISecret)
	params["format"] = "json"
	go lastfmPost(params)
}

func SetupLastFM() {
	fmt.Print("Last.fm API Key: ")
	var apiKey string
	fmt.Scan(&apiKey)
	fmt.Print("Last.fm API Secret: ")
	var apiSecret string
	fmt.Scan(&apiSecret)
	fmt.Print("Last.fm Username: ")
	var username string
	fmt.Scan(&username)
	fmt.Print("Last.fm Password: ")
	var password string
	fmt.Scan(&password)

	params := map[string]string{
		"method":   "auth.getMobileSession",
		"api_key":  apiKey,
		"username": username,
		"password": password,
	}
	params["api_sig"] = lastfmSign(params, apiSecret)
	params["format"] = "json"

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	resp, err := http.PostForm("http://ws.audioscrobbler.com/2.0/", form)
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Session struct {
			Key string `json:"key"`
		} `json:"session"`
		Error   int    `json:"error"`
		Message string `json:"message"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != 0 {
		fmt.Println("Last.fm error:", result.Message)
		return
	}

	lib := LoadLibrary()
	lib.LastFM = models.LastFMConfig{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		SessionKey: result.Session.Key,
		Username:   username,
		Enabled:    true,
	}
	SaveLibrary(lib)
	fmt.Printf("Last.fm connected! Welcome, %s\n", username)
}
