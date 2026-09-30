package models

import "time"

func (lib *Library) FindSong(id string) *Song {
	for i := range lib.Songs {
		if lib.Songs[i].ID == id {
			return &lib.Songs[i]
		}
	}
	return nil
}

func (lib *Library) Upsert(s Song) {
	for i, existing := range lib.Songs {
		if existing.ID == s.ID {
			if s.LocalPath != "" {
				lib.Songs[i].LocalPath = s.LocalPath
			}
			lib.Songs[i].Favorite = s.Favorite
			return
		}
	}
	lib.Songs = append(lib.Songs, s)
}

func (lib *Library) AddToHistory(s Song) {
	entry := HistoryEntry{Song: s, PlayedAt: time.Now()}
	lib.History = append([]HistoryEntry{entry}, lib.History...)
	if len(lib.History) > 200 {
		lib.History = lib.History[:200]
	}
}

func (lib *Library) FindPlaylist(name string) *Playlist {
	for i := range lib.Playlists {
		if lib.Playlists[i].Name == name {
			return &lib.Playlists[i]
		}
	}
	return nil
}

func (lib *Library) CreatePlaylist(name string) bool {
	if lib.FindPlaylist(name) != nil {
		return false
	}
	lib.Playlists = append(lib.Playlists, Playlist{Name: name, SongIDs: []string{}})
	return true
}

func (lib *Library) AddSongToPlaylist(playlistName, songID string) {
	for i := range lib.Playlists {
		if lib.Playlists[i].Name == playlistName {
			for _, id := range lib.Playlists[i].SongIDs {
				if id == songID {
					return
				}
			}
			lib.Playlists[i].SongIDs = append(lib.Playlists[i].SongIDs, songID)
			return
		}
	}
}

func (lib *Library) GetPlaylistSongs(name string) []Song {
	pl := lib.FindPlaylist(name)
	if pl == nil {
		return nil
	}
	var songs []Song
	for _, id := range pl.SongIDs {
		if s := lib.FindSong(id); s != nil {
			songs = append(songs, *s)
		}
	}
	return songs
}

func (lib *Library) DeletePlaylist(name string) {
	for i, pl := range lib.Playlists {
		if pl.Name == name {
			lib.Playlists = append(lib.Playlists[:i], lib.Playlists[i+1:]...)
			return
		}
	}
}

func (lib *Library) RemoveSongFromPlaylist(playlistName, songID string) {
	for i := range lib.Playlists {
		if lib.Playlists[i].Name == playlistName {
			for j, id := range lib.Playlists[i].SongIDs {
				if id == songID {
					lib.Playlists[i].SongIDs = append(
						lib.Playlists[i].SongIDs[:j],
						lib.Playlists[i].SongIDs[j+1:]...,
					)
					return
				}
			}
		}
	}
}
