package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/rugved-danej/termux-music/internal/models"
)

func (m *model) toListItems(songs []models.Song, mergeLib bool) []list.Item {
	items := make([]list.Item, len(songs))
	for i, s := range songs {
		if mergeLib {
			if existing := m.library.FindSong(s.ID); existing != nil {
				s.Favorite = existing.Favorite
				s.LocalPath = existing.LocalPath
			}
		}
		items[i] = listItem{song: s}
	}
	return items
}

func (m *model) loadLibraryList() {
	l := makeList("Your Library")
	l.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
	l.SetItems(m.toListItems(m.library.Songs, false))
	m.list = l
}

func (m *model) loadPlaylistsList() {
	l := makeList("Playlists")
	l.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
	items := make([]list.Item, len(m.library.Playlists))
	for i, pl := range m.library.Playlists {
		items[i] = playlistItem{pl: pl}
	}
	l.SetItems(items)
	m.list = l
}

func (m *model) loadPlaylistSongsList(name string) {
	l := makeList(name)
	l.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
	l.SetItems(m.toListItems(m.library.GetPlaylistSongs(name), false))
	m.list = l
}

func (m *model) loadHistoryList() {
	l := makeList("Recently Played")
	l.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
	var songs []models.Song
	for _, entry := range m.library.History {
		songs = append(songs, entry.Song)
	}
	l.SetItems(m.toListItems(songs, true))
	m.list = l
}

func (m *model) loadQueueList() {
	l := makeList("Up Next")
	l.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
	l.SetItems(m.toListItems(m.queue, true))
	m.list = l
}

func (m *model) loadAddToPlaylistList() {
	l := makeList("Add to Playlist")
	l.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
	items := make([]list.Item, len(m.library.Playlists))
	for i, pl := range m.library.Playlists {
		items[i] = playlistItem{pl: pl}
	}
	l.SetItems(items)
	m.list = l
}

func (m *model) loadDownloadMenu() {
	l := makeList("Download: " + m.downloadTarget.Title)
	l.SetSize(clamp(m.width-4, 10, 9999), listHeight(m.height))
	l.SetItems([]list.Item{
		dlOptionItem{
			title: "MP3 Audio (Standard)", desc: "Best compatibility",
			args: []string{"-x", "--audio-format", "mp3", "--audio-quality", "0"}, ext: "mp3",
		},
		dlOptionItem{
			title: "M4A Audio (High Quality)", desc: "Better compression",
			args: []string{"-x", "--audio-format", "m4a"}, ext: "m4a",
		},
		dlOptionItem{
			title: "FLAC Audio (Lossless)", desc: "Highest quality audio",
			args: []string{"-x", "--audio-format", "flac"}, ext: "flac",
		},
		dlOptionItem{
			title: "MP4 Video (1080p)", desc: "Full HD Video",
			args: []string{"-f", "bestvideo[height<=1080][ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best"}, ext: "mp4",
		},
		dlOptionItem{
			title: "MP4 Video (720p)", desc: "HD Video (Smaller file)",
			args: []string{"-f", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best"}, ext: "mp4",
		},
	})
	m.list = l
}
