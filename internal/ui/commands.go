package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rugved-danej/termux-music/internal/core"
	"github.com/rugved-danej/termux-music/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func animTickCmd() tea.Cmd {
	return tea.Tick(60*time.Millisecond, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

func doDownloadCmd(song models.Song, args []string, ext string) tea.Cmd {
	return func() tea.Msg {
		outDir := models.MusicDir
		storageMusic := filepath.Join(os.Getenv("HOME"), "storage", "music")
		storageDownloads := filepath.Join(os.Getenv("HOME"), "storage", "downloads")

		isAudio := strings.Contains(strings.Join(args, " "), "-x")
		if isAudio {
			if _, err := os.Stat(storageMusic); err == nil {
				outDir = filepath.Join(storageMusic, "termux-music")
				os.MkdirAll(outDir, 0755)
			}
		} else {
			if _, err := os.Stat(storageDownloads); err == nil {
				outDir = storageDownloads
			}
		}
		outPath := filepath.Join(outDir, song.ID+".%(ext)s")
		finalArgs := append(args, "-o", outPath, fmt.Sprintf("https://www.youtube.com/watch?v=%s", song.ID))
		cmd := exec.Command("yt-dlp", finalArgs...)
		devnull, _ := os.Open(os.DevNull)
		defer devnull.Close()
		cmd.Stdout = devnull
		cmd.Stderr = devnull
		if err := cmd.Run(); err != nil {
			return errMsg("Download failed")
		}
		if isAudio {
			return downloadDoneMsg{path: filepath.Join(outDir, song.ID+"."+ext)}
		}
		return downloadDoneMsg{path: ""}
	}
}

type ytUpdateDoneMsg struct{ err error }

func updateYTDLPCmd() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("pip", "install", "--upgrade", "yt-dlp")
		err := cmd.Run()
		return ytUpdateDoneMsg{err: err}
	}
}

type importDoneMsg struct {
	added int
	lib   models.Library
}

func importLocalFilesCmd(lib models.Library) tea.Cmd {
	return func() tea.Msg {
		added := core.ImportLocalFiles(&lib)
		return importDoneMsg{added: added, lib: lib}
	}
}

func fetchDiscoverCmd(lib models.Library) tea.Cmd {
	return func() tea.Msg {
		seedQuery := "trending music"
		if len(lib.History) > 0 {
			seedQuery = lib.History[0].Song.Channel + " new songs recommended"
		} else if len(lib.Songs) > 0 {
			for _, s := range lib.Songs {
				if s.Favorite {
					seedQuery = s.Channel + " mix recommended"
					break
				}
			}
		}
		songs := core.YtdlpSearch(seedQuery, 20)
		if songs == nil {
			return errMsg("Could not fetch recommendations")
		}
		return discoverMsg(songs)
	}
}
