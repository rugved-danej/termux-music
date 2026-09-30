package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/rugved-danej/termux-music/internal/core"
	"github.com/rugved-danej/termux-music/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) startPlaying(song models.Song) (model, tea.Cmd) {
	if song.AddedAt == "" {
		song.AddedAt = time.Now().Format(time.RFC3339)
	}
	m.library.AddToHistory(song)
	m.library.Upsert(song)
	core.SaveLibrary(m.library)
	m.playing = song
	m.state = statePlaying
	m.mpvPos = 0
	m.paused = false
	m.scrobbled = false
	m.lyrics = nil
	m.lyricOffset = 0
	m.trackArmed = false
	core.AndroidNotification(song.Title, song.Channel, "Playing")
	core.LastfmNowPlaying(song, m.library.LastFM)
	startPos := 0.0
	if existing := m.library.FindSong(song.ID); existing != nil {
		if existing.LastPos > 15.0 && (song.Duration == 0 || existing.LastPos < song.Duration-15.0) {
			startPos = existing.LastPos
			m.statusMsg = fmt.Sprintf("Resumed at %s", models.FormatDuration(startPos))
		}
	}
	m.mpvPos = startPos

	return *m, tea.Batch(
		PlayMpvTrackCmd(song, models.MpvSocket, core.EqPresets[m.eqIndex].Values, startPos, m.speed),
		tickCmd(),
		FetchLyricsCmd(song),
	)
}

func (m *model) updateLyricOffset() {
	if len(m.lyrics) == 0 {
		return
	}
	for i, line := range m.lyrics {
		if line.Time > m.mpvPos {
			if i > 0 {
				m.lyricOffset = i - 1
			}
			return
		}
	}
	m.lyricOffset = len(m.lyrics) - 1
}
func (m model) nowPlayingView() string {
	B := bp(m.width)
	w := clamp(m.width-4, 20, 9999)
	innerW := w - 6

	dur := m.playing.Duration
	pos := m.mpvPos
	var pct float64
	if dur > 0 {
		pct = pos / dur
	}

	visualizer := ""
	if !m.paused {
		bars := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
		for i := 0; i < 5; i++ {
			idx := (m.animFrame + i*3) % len(bars)
			if (m.animFrame/len(bars))%2 != 0 {
				idx = len(bars) - 1 - idx
			}
			visualizer += bars[idx]
		}
		visualizer = " " + lipgloss.NewStyle().Foreground(colorCyan).Render(visualizer)
	}

	playState := lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("PLAYING")
	if m.paused {
		playState = lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Render("PAUSED")
	}
	playState += visualizer

	badges := ""
	if m.playing.Favorite {
		badges += " " + lipgloss.NewStyle().Foreground(colorPink).Render("[fav]")
	}
	if m.playing.LocalPath != "" {
		badges += " " + lipgloss.NewStyle().Foreground(colorGreen).Render("[local]")
	}
	if m.radioOn {
		badges += " " + lipgloss.NewStyle().Foreground(colorOrange).Render("[radio]")
	}
	if m.shuffle {
		badges += " " + lipgloss.NewStyle().Foreground(colorCyan).Render("[shuf]")
	}
	if m.repeatMode == repeatAll {
		badges += " " + lipgloss.NewStyle().Foreground(colorPurple).Render("[loop]")
	} else if m.repeatMode == repeatOne {
		badges += " " + lipgloss.NewStyle().Foreground(colorPurple).Render("[loop:1]")
	}

	maxTitleW := clamp(innerW-4, 10, 9999)
	displayTitle := m.playing.Title
	runes := []rune(displayTitle)
	if len(runes) > maxTitleW && !m.paused {
		scrollIdx := (m.animFrame / 5) % (len(runes) + 10)
		padded := append(runes, []rune("          ")...)
		padded = append(padded, runes...)
		if scrollIdx < len(runes) {
			displayTitle = string(padded[scrollIdx:])
		}
	}

	titleStr := lipgloss.NewStyle().Foreground(colorGreen).Bold(true).
		Render(truncate(displayTitle, maxTitleW))
	channelStr := lipgloss.NewStyle().Foreground(colorGray).Render(m.playing.Channel)
	timeStr := lipgloss.NewStyle().Foreground(colorGray).
		Render(fmt.Sprintf(" %s / %s", models.FormatDuration(pos), models.FormatDuration(dur)))

	var statusParts []string
	if len(m.queue) > 0 {
		statusParts = append(statusParts,
			lipgloss.NewStyle().Foreground(colorCyan).Render(fmt.Sprintf("Queue:%d", len(m.queue))))
	}
	if !m.sleepEnd.IsZero() {
		rem := time.Until(m.sleepEnd)
		sm := int(rem.Minutes())
		ss := int(rem.Seconds()) % 60
		statusParts = append(statusParts,
			lipgloss.NewStyle().Foreground(colorOrange).Render(fmt.Sprintf("Sleep:%d:%02d", sm, ss)))
	}
	if m.speed != 1.0 {
		statusParts = append(statusParts,
			lipgloss.NewStyle().Foreground(colorOrange).Render(fmt.Sprintf("Spd:%.1fx", m.speed)))
	}
	statusParts = append(statusParts,
		lipgloss.NewStyle().Foreground(colorPurple).Render("EQ:"+core.EqPresets[m.eqIndex].Name))
	statusParts = append(statusParts,
		lipgloss.NewStyle().Foreground(colorCyan).Render(fmt.Sprintf("Vol:%d%%", m.volume)))

	statusLine := playState + "  " + strings.Join(statusParts, "  |  ") + badges

	barWidth := progressWidth(w)
	progBar := customProgressBar(barWidth, pct, m.animFrame, m.paused)

	playerContent := titleStr + "\n" +
		channelStr + "\n\n" +
		progBar + timeStr + "\n\n" +
		divider(innerW) + "\n" +
		statusLine

	playerBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorGreen).
		Padding(1, 2).
		Width(w).
		Render(playerContent)

	fixedHeight := 18
	if B != "narrow" {
		fixedHeight = 17
	}
	lyricsLines := m.height - fixedHeight
	if lyricsLines < 2 {
		lyricsLines = 2
	}
	if lyricsLines > 12 {
		lyricsLines = 12
	}

	var lyricsContent string
	if len(m.lyrics) > 0 {
		start := clamp(m.lyricOffset-2, 0, len(m.lyrics)-1)
		end := clamp(start+lyricsLines, 0, len(m.lyrics))
		var lines []string
		for i := start; i < end; i++ {
			if i == m.lyricOffset {
				lines = append(lines,
					lipgloss.NewStyle().Foreground(colorYellow).Bold(true).
						Render("> "+truncate(m.lyrics[i].Text, innerW-4)))
			} else {
				lines = append(lines,
					lipgloss.NewStyle().Foreground(colorGray).
						Render("  "+truncate(m.lyrics[i].Text, innerW-4)))
			}
		}
		lyricsContent = strings.Join(lines, "\n")
	} else {
		if m.statusMsg == "No lyrics found" {
			lyricsContent = lipgloss.NewStyle().Foreground(colorDim).Italic(true).
				Render("  No lyrics found for this track.")
		} else {
			lyricsContent = lipgloss.NewStyle().Foreground(colorDim).Italic(true).
				Render("  Fetching lyrics...")
		}
	}

	lyricsTitle := lipgloss.NewStyle().Foreground(colorPurple).Bold(true).Render("Lyrics")
	lyricsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPurple).
		Padding(1, 2).
		Width(w).
		Render(lyricsTitle + "\n" + divider(innerW) + "\n" + lyricsContent)

	var ctrlLines []string
	if B == "narrow" {
		ctrlLines = []string{
			"[Space] Pause  [←→] Seek  [+/-] Vol  [[]] Spd",
			"[n] Skip  [s] Shuf  [l] Loop  [r] Radio",
			"[e] EQ  [t] Sleep  [f] Fav  [a] PL  [Esc] Back",
		}
	} else {
		ctrlLines = []string{
			"[Space] Pause  [←→] Seek  [+/-] Vol  [[]] Spd  [n] Skip  [s] Shuf  [l] Loop",
			"[r] Radio  [e] EQ  [t] Sleep  [f] Fav  [d] DL  [a] PL  [Esc] Back",
		}
	}
	controls := lipgloss.NewStyle().Foreground(colorGray).MarginTop(1).
		Render(strings.Join(ctrlLines, "\n"))

	return lipgloss.NewStyle().Margin(1, 1).
		Render(playerBox + "\n" + lyricsBox + "\n" + controls)
}
