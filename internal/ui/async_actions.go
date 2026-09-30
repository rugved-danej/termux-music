package ui

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rugved-danej/termux-music/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rugved-danej/termux-music/internal/core"
)

func SearchYoutubeCmd(query string) tea.Cmd {
	return func() tea.Msg {
		songs := core.YtdlpSearch(query, 15)
		if songs == nil {
			return errMsg("Search failed. Check your connection.")
		}
		return searchResultMsg(songs)
	}
}

func FetchTrendingCmd() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("yt-dlp", "-j", "--flat-playlist",
			"--playlist-end", "30", core.TrendingPlaylistURL)
		out, err := cmd.Output()
		var songs []models.Song
		if err != nil || len(out) < 10 {
			songs = core.YtdlpSearch("top music hits official video 2024", 30)
		} else {
			songs = core.ParseYTDLPOutput(out)
			if len(songs) == 0 {
				songs = core.YtdlpSearch("top music hits official video 2024", 30)
			}
		}
		if songs == nil {
			return errMsg("Could not fetch trending music")
		}
		return trendingMsg(songs)
	}
}

func FetchArtistRadioCmd(artist string) tea.Cmd {
	return func() tea.Msg {
		songs := core.YtdlpSearch(artist+" best songs official", 20)
		if songs == nil {
			songs = core.YtdlpSearch(artist+" music", 15)
		}
		if songs == nil {
			return errMsg("Radio failed")
		}
		return artistRadioMsg(songs)
	}
}

func FetchLyricsCmd(song models.Song) tea.Cmd {
	return func() tea.Msg {
		if len(song.Lyrics) > 0 {
			return lyricsReadyMsg(song.Lyrics)
		}
		return lyricsReadyMsg(core.FetchLyrics(song.Title, song.Channel))
	}
}

func PlayMpvTrackCmd(song models.Song, socket, eqValues string, startAt float64, speed float64) tea.Cmd {
	return func() tea.Msg {
		core.AcquireWakeLock()
		playURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", song.ID)
		if song.LocalPath != "" {
			if _, err := os.Stat(song.LocalPath); err == nil {
				playURL = song.LocalPath
			}
		}

		af := "silenceremove=start_periods=1:start_threshold=-50dB,loudnorm=I=-14:TP=-1.5:LRA=11,afade=t=in:ss=0:d=1.5"
		if eqValues != "" && eqValues != "0:0:0:0:0:0:0:0:0:0" {
			af += ",@eq:equalizer=" + eqValues
		}
		core.MpvIPC(socket, fmt.Sprintf(`{"command":["set_property","af","%s"]}`, af))

		if speed != 1.0 && speed > 0 {
			core.MpvIPC(socket, fmt.Sprintf(`{"command":["set_property","speed",%f]}`, speed))
		} else {
			core.MpvIPC(socket, `{"command":["set_property","speed",1.0]}`)
		}

		core.MpvIPC(socket, fmt.Sprintf(`{"command":["loadfile","%s","replace"]}`, playURL))

		if startAt > 0 {
			core.MpvIPC(socket, fmt.Sprintf(`{"command":["set_property","time-pos",%f]}`, startAt))
		}
		return nil
	}
}
